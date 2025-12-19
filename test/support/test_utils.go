package support

import (
	"os"
	"unicode"
)

// CapitalizeFirst capitalizes the first letter of a string
func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
}

// IsDryRun checks if dry run mode is enabled via DRY_RUN environment variable
func IsDryRun() bool {
	return os.Getenv("DRY_RUN") == "true" || os.Getenv("DRY_RUN") == "1"
}
