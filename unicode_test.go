package gopcre2

import (
	"testing"
)

func TestUnicodeProperties(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    bool
	}{
		// General categories
		{`\p{L}`, "a", true},
		{`\p{L}`, "1", false},
		{`\p{Lu}`, "A", true},
		{`\p{Lu}`, "a", false},
		{`\p{Ll}`, "a", true},
		{`\p{Ll}`, "A", false},
		{`\p{N}`, "5", true},
		{`\p{N}`, "a", false},
		{`\p{Nd}`, "5", true},
		{`\p{Nd}`, "a", false},
		{`\p{P}`, "!", true},
		{`\p{P}`, "a", false},
		{`\p{S}`, "$", true},
		{`\p{Z}`, " ", true},

		// Negation
		{`\P{L}`, "1", true},
		{`\P{L}`, "a", false},

		// Single-letter shorthand
		{`\pL`, "a", true},
		{`\pL`, "1", false},
		{`\PL`, "1", true},
		{`\PL`, "a", false},

		// Scripts
		{`\p{Greek}`, "\u03B1", true},  // α
		{`\p{Greek}`, "a", false},
		{`\p{Latin}`, "a", true},
		{`\p{Latin}`, "\u03B1", false},
		{`\p{Han}`, "\u4e00", true},    // 一
		{`\p{Cyrillic}`, "\u0410", true}, // А (Cyrillic A)

		// Negated with ^
		{`\p{^L}`, "1", true},
		{`\p{^L}`, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.subject, func(t *testing.T) {
			re, err := Compile(tt.pattern)
			if err != nil {
				t.Fatalf("Compile(%q): %v", tt.pattern, err)
			}
			got := re.MatchString(tt.subject)
			if got != tt.want {
				t.Errorf("MatchString(%q) = %v, want %v", tt.subject, got, tt.want)
			}
		})
	}
}

func TestUCPMode(t *testing.T) {
	// In UCP mode, \d matches Unicode digits, \w matches Unicode word chars
	re := MustCompile(`\d+`, UCP)
	// Arabic-Indic digit
	if !re.MatchString("\u0660") {
		t.Error("UCP \\d should match Arabic-Indic digit")
	}

	// \w in UCP mode
	reW := MustCompile(`\w+`, UCP)
	if !reW.MatchString("\u00e9") { // é
		t.Error("UCP \\w should match accented letter")
	}
}

func TestPropertyInCharClass(t *testing.T) {
	re := MustCompile(`[\p{Lu}\p{Ll}]+`)
	got := re.FindString("Hello123World")
	if got != "Hello" {
		t.Errorf("got %q, want 'Hello'", got)
	}
}

func TestResolveProperty(t *testing.T) {
	tests := []string{
		"L", "Lu", "Ll", "Lt", "Lm", "Lo",
		"M", "Mn", "Mc", "Me",
		"N", "Nd", "Nl", "No",
		"P", "Pc", "Pd", "Ps", "Pe", "Pi", "Pf", "Po",
		"S", "Sm", "Sc", "Sk", "So",
		"Z", "Zs", "Zl", "Zp",
		"C", "Cc", "Cf", "Co", "Cs",
		"Latin", "Greek", "Cyrillic", "Han", "Arabic",
		"Any", "L&",
	}
	for _, name := range tests {
		if resolveProperty(name) == nil {
			t.Errorf("resolveProperty(%q) = nil", name)
		}
	}
}
