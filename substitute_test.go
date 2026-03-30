package gopcre2

import (
	"testing"
)

func TestSubstitute(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		repl    string
		want    string
	}{
		// Basic substitution
		{"(\\w+) (\\w+)", "hello world", "$2 $1", "world hello"},
		// Named groups
		{"(?P<first>\\w+) (?P<last>\\w+)", "John Doe", "${last}, ${first}", "Doe, John"},
		// Case conversion
		{"(\\w+)", "hello", "\\U$1", "HELLO"},
		{"(\\w+)", "HELLO", "\\L$1", "hello"},
		{"(\\w+)", "hello", "\\u$1", "Hello"},
		{"(\\w+)", "HELLO", "\\l$1", "hELLO"},
		// $0 whole match
		{"\\w+", "hello", "[$0]", "[hello]"},
		// $& whole match
		{"\\w+", "hello", "[$&]", "[hello]"},
		// Literal $$
		{"\\w+", "hello", "$$100", "$100"},
	}

	for _, tt := range tests {
		t.Run(tt.repl, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			got, err := re.Substitute(tt.subject, tt.repl)
			if err != nil {
				t.Fatalf("Substitute error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Substitute(%q, %q) = %q, want %q", tt.subject, tt.repl, got, tt.want)
			}
		})
	}
}

func TestSubstituteAll(t *testing.T) {
	re := MustCompile("\\d+")
	got, err := re.SubstituteAll("a1b22c333", "[NUM]")
	if err != nil {
		t.Fatal(err)
	}
	if got != "a[NUM]b[NUM]c[NUM]" {
		t.Errorf("SubstituteAll = %q", got)
	}
}

func TestSubstituteNoMatch(t *testing.T) {
	re := MustCompile("xyz")
	got, err := re.Substitute("hello", "replaced")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("Substitute with no match should return original, got %q", got)
	}
}

func TestSubstituteCaseConversion(t *testing.T) {
	re := MustCompile("(\\w+) (\\w+)")
	got, err := re.Substitute("hello world", "\\U$1\\E $2")
	if err != nil {
		t.Fatal(err)
	}
	if got != "HELLO world" {
		t.Errorf("got %q, want 'HELLO world'", got)
	}
}
