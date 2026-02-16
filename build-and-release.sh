#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version from argument or prompt
VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    echo -n "Enter version (e.g., v0.1.0): "
    read VERSION
fi

# Validate version format
if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Version must be in format vX.Y.Z${NC}"
    exit 1
fi

echo -e "${GREEN}🚀 Starting release process for $VERSION${NC}"

# Step 1: Ensure all changes are committed
echo -e "\n${YELLOW}Step 1: Checking for uncommitted changes...${NC}"
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo -e "${RED}Error: Uncommitted changes found. Please commit or stash them first.${NC}"
    exit 1
fi

# Step 2: Create and push tag
echo -e "\n${YELLOW}Step 2: Creating tag...${NC}"
if git tag | grep -q "^${VERSION}$"; then
    echo -e "${RED}Error: Tag $VERSION already exists${NC}"
    exit 1
else
    echo "Enter release notes (press Ctrl-D when done):"
    RELEASE_NOTES=$(cat)
    git tag -a "$VERSION" -m "$RELEASE_NOTES"
fi

# Step 3: Push tag to GitHub
echo -e "\n${YELLOW}Step 3: Pushing tag to GitHub...${NC}"
git push origin "$VERSION"

# Step 4: Build binaries for current platform
echo -e "\n${YELLOW}Step 4: Building binaries...${NC}"
rm -rf dist
mkdir -p dist

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

# Build the binary
echo "Building for $OS/$ARCH..."
go build -o "dist/atask" .

# Create archive
ARCHIVE_NAME="atask_${VERSION}_${OS}_${ARCH}.tar.gz"
cd dist
cp ../README.md .
FILES_TO_ARCHIVE="atask README.md"
if [ -f ../test-config.toml ]; then
    cp ../test-config.toml ./config.example.toml
    FILES_TO_ARCHIVE="$FILES_TO_ARCHIVE config.example.toml"
fi
if [ -f ../LICENSE ]; then
    cp ../LICENSE .
    FILES_TO_ARCHIVE="$FILES_TO_ARCHIVE LICENSE"
fi
# Include shell completions in the binary archive
if [ -d ../completions ]; then
    cp -r ../completions .
    FILES_TO_ARCHIVE="$FILES_TO_ARCHIVE completions"
fi
tar czf "$ARCHIVE_NAME" $FILES_TO_ARCHIVE
rm README.md
if [ -f config.example.toml ]; then
    rm config.example.toml
fi
if [ -f LICENSE ]; then
    rm LICENSE
fi
if [ -d completions ]; then
    rm -rf completions
fi
cd ..

# Create checksums
cd dist
shasum -a 256 *.tar.gz > checksums.txt
cd ..

# Step 5: Create GitHub release
echo -e "\n${YELLOW}Step 5: Creating GitHub release...${NC}"
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) not found. Install with: brew install gh${NC}"
    echo -e "${YELLOW}Alternatively, create the release manually at:${NC}"
    echo "https://github.com/pdxmph/atask/releases/new?tag=$VERSION"
    echo "Upload these files:"
    ls -la dist/*.tar.gz dist/checksums.txt
else
    gh release create "$VERSION" \
      --title "$VERSION" \
      --notes "$RELEASE_NOTES" \
      dist/*.tar.gz \
      dist/checksums.txt
fi

echo -e "\n${GREEN}✅ Release $VERSION completed successfully!${NC}"
echo -e "${GREEN}View at: https://github.com/pdxmph/atask/releases/tag/$VERSION${NC}"

# Step 6: Update Homebrew formula
echo -e "\n${YELLOW}Step 6: Updating Homebrew formula...${NC}"
if [ -d "$HOME/code/homebrew-tap" ]; then
    # Calculate SHA256 for source archive
    SOURCE_URL="https://github.com/pdxmph/atask/archive/refs/tags/${VERSION}.tar.gz"
    echo "Downloading source archive for SHA calculation..."
    TEMP_SOURCE="/tmp/atask-source-${VERSION}.tar.gz"
    curl -sL -o "$TEMP_SOURCE" "$SOURCE_URL"
    SOURCE_SHA=$(shasum -a 256 "$TEMP_SOURCE" | cut -d' ' -f1)
    rm -f "$TEMP_SOURCE"
    
    # Calculate SHA256 for binary archive
    BINARY_SHA=$(shasum -a 256 "dist/$ARCHIVE_NAME" | cut -d' ' -f1)
    
    # Update or create the formula
    FORMULA_PATH="$HOME/code/homebrew-tap/Formula/atask.rb"
    cat > "$FORMULA_PATH" << EOF
class DenoteTasks < Formula
  desc "Task management tool using Denote file naming convention"
  homepage "https://github.com/pdxmph/atask"
  version "${VERSION#v}"
  license "MIT"
  
  # Use binary release for ARM64 Macs
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/pdxmph/atask/releases/download/${VERSION}/${ARCHIVE_NAME}"
    sha256 "${BINARY_SHA}"
  else
    # Fall back to building from source
    url "${SOURCE_URL}"
    sha256 "${SOURCE_SHA}"
    depends_on "go" => :build
  end
  
  depends_on arch: :arm64  # Currently only ARM64 builds available

  def install
    if File.exist?("go.mod")
      # Building from source
      system "go", "build", *std_go_args(ldflags: "-s -w")
      
      # Install completions from source
      bash_completion.install "completions/atask.bash"
      zsh_completion.install "completions/_atask"
    else
      # Installing pre-built binary
      bin.install "atask"
      
      # Install completions from binary archive
      bash_completion.install "completions/atask.bash" if File.exist?("completions/atask.bash")
      zsh_completion.install "completions/_atask" if File.exist?("completions/_atask")
    end
    
    # Install documentation
    doc.install "README.md" if File.exist?("README.md")
  end

  def caveats
    <<~EOS
      To use atask, create ~/.config/atask/config.toml with:

        notes_directory = "~/tasks"
        editor = "vim"  # or your preferred editor
    EOS
  end

  test do
    assert_match "atask", shell_output("#{bin}/atask --help 2>&1")
  end
end
EOF
    
    echo -e "${GREEN}✅ Homebrew formula updated${NC}"
    
    # Commit and push the formula update
    cd "$HOME/code/homebrew-tap"
    git add "Formula/atask.rb"
    git commit -m "atask ${VERSION}" || echo "No changes to commit"
    git push || echo "Failed to push - you may need to push manually"
    cd - > /dev/null
    
    echo -e "\n${YELLOW}Homebrew installation will be available after push completes:${NC}"
    echo "  brew tap pdxmph/homebrew-tap  # if not already tapped"
    echo "  brew install pdxmph/homebrew-tap/atask"
else
    echo -e "${YELLOW}Homebrew tap directory not found at ~/code/homebrew-tap${NC}"
    echo "Skipping Homebrew formula update."
fi

# Show installation instructions
echo -e "\n${YELLOW}Installation instructions:${NC}"
echo "Via Homebrew (recommended):"
echo "  brew tap pdxmph/homebrew-tap"
echo "  brew install pdxmph/homebrew-tap/atask"
echo ""
echo "Or manually:"
echo "  curl -L https://github.com/pdxmph/atask/releases/download/$VERSION/$ARCHIVE_NAME | tar xz"
echo "  sudo mv atask /usr/local/bin/"
echo ""
echo "Or download directly from: https://github.com/pdxmph/atask/releases/tag/$VERSION"
