package gopcre2

import "testing"

func TestCharClassEdgeCases(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    bool
	}{
		// Literal ] as first char in class
		{"[]abc]", "]", true},
		{"[]abc]", "a", true},
		{"[]abc]", "d", false},

		// Literal - at start or end
		{"[-abc]", "-", true},
		{"[abc-]", "-", true},

		// Mixed ranges and literals
		{"[a-zA-Z0-9_]", "z", true},
		{"[a-zA-Z0-9_]", "Z", true},
		{"[a-zA-Z0-9_]", "5", true},
		{"[a-zA-Z0-9_]", "_", true},
		{"[a-zA-Z0-9_]", " ", false},

		// Negated class
		{"[^0-9]", "a", true},
		{"[^0-9]", "5", false},

		// Escaped chars in class
		{"[\\n\\t]", "\n", true},
		{"[\\n\\t]", "\t", true},
		{"[\\n\\t]", "a", false},

		// Hex escape in class
		{"[\\x41-\\x5A]", "A", true},
		{"[\\x41-\\x5A]", "Z", true},
		{"[\\x41-\\x5A]", "a", false},

		// \d \w \s in class
		{"[\\d]+", "123", true},
		{"[\\w]+", "abc", true},
		{"[\\s]+", " \t", true},
		{"[\\d\\s]+", "1 2", true},
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

func TestCaseInsensitiveCharClass(t *testing.T) {
	re := MustCompile("[a-z]+", Caseless)
	if !re.MatchString("ABC") {
		t.Error("case-insensitive [a-z]+ should match 'ABC'")
	}

	re2 := MustCompile("[A-Z]+", Caseless)
	if !re2.MatchString("abc") {
		t.Error("case-insensitive [A-Z]+ should match 'abc'")
	}
}
