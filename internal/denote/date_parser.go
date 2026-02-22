package denote

import "github.com/mph-llm-experiments/acore"

// ParseNaturalDate parses natural language dates into YYYY-MM-DD format.
// Delegates to acore.ParseNaturalDate.
func ParseNaturalDate(input string) (string, error) {
	return acore.ParseNaturalDate(input)
}
