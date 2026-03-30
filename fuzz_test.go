package gopcre2

import (
	"testing"
)

func FuzzCompile(f *testing.F) {
	// Seed corpus with valid patterns
	seeds := []string{
		"abc",
		"a+b*c?",
		"(a|b)+",
		"[a-z0-9]+",
		"\\d{3}-\\d{4}",
		"^hello$",
		"(?:foo|bar)",
		"(?<name>\\w+)",
		"(?=abc)",
		"(?!abc)",
		"a*?b+?",
		"\\p{L}+",
		"[[:alpha:]]",
		"(?>a+)",
		"\\bword\\b",
		"(?i)abc",
		"a{2,5}",
		"",
		".",
		"\\n\\t\\r",
		"\\x41",
		"(a)(b)\\1",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, pattern string) {
		// Compile should not panic
		re, err := Compile(pattern)
		if err != nil {
			return // compile errors are expected for fuzz input
		}

		// Set very tight limits to prevent hangs
		re.SetMatchLimit(10000)
		re.SetDepthLimit(10)
		re.SetHeapLimit(1024 * 1024)

		// These should not panic
		_ = re.MatchString("test input string 12345")
		_ = re.FindString("test input string 12345")
		_ = re.FindAllString("test input string 12345", 5)
		_ = re.NumSubexp()
		_ = re.String()
	})
}

func FuzzMatch(f *testing.F) {
	// Seed with pattern+subject pairs
	f.Add("\\d+", "hello123world")
	f.Add("[a-z]+", "ABC123def")
	f.Add("(a+)(b+)", "aaabbb")
	f.Add("(?:foo|bar)", "foobar")
	f.Add("^abc$", "abc")
	f.Add(".", "x")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, pattern, subject string) {
		re, err := Compile(pattern)
		if err != nil {
			return
		}

		re.SetMatchLimit(10000)
		re.SetDepthLimit(10)
		re.SetHeapLimit(1024 * 1024)

		// Should not panic
		_ = re.MatchString(subject)
		_ = re.FindString(subject)
		_ = re.FindStringSubmatch(subject)
		_ = re.FindAllString(subject, 10)
		_ = re.ReplaceAllString(subject, "X")
	})
}
