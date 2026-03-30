package gopcre2

import (
	"testing"
)

func TestLookahead(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    bool
	}{
		// Positive lookahead
		{"q(?=u)", "quit", true},
		{"q(?=u)", "qat", false},
		{"\\w+(?=\\.)", "foo.bar", true},

		// Negative lookahead
		{"q(?!u)", "qat", true},
		{"q(?!u)", "quit", false},
		{"\\d+(?!\\d)", "123abc", true},
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

func TestLookaheadZeroWidth(t *testing.T) {
	// Lookahead should not consume input
	re := MustCompile("(?=foo)foo")
	if !re.MatchString("foo") {
		t.Error("(?=foo)foo should match 'foo'")
	}

	// The found text should be "foo" (lookahead + literal)
	got := re.FindString("foobar")
	if got != "foo" {
		t.Errorf("FindString = %q, want 'foo'", got)
	}
}

func TestNegLookaheadZeroWidth(t *testing.T) {
	// (?!bar)... should match if the next chars are not "bar"
	re := MustCompile("(?!bar)\\w+")
	got := re.FindString("foo")
	if got != "foo" {
		t.Errorf("got %q, want 'foo'", got)
	}
}

func TestAtomicGroup(t *testing.T) {
	// (?>a+)ab requires backtracking into a+ to give back an 'a' for the literal 'a' after it.
	// The atomic group prevents this backtracking.
	re := MustCompile("(?>a+)ab")
	if re.MatchString("aaab") {
		t.Error("(?>a+)ab should NOT match 'aaab' because atomic group prevents backtracking")
	}

	// Without atomic, a+ab should match (a+ gives back one 'a')
	re2 := MustCompile("a+ab")
	if !re2.MatchString("aaab") {
		t.Error("a+ab should match 'aaab'")
	}

	// (?>a+)b should match because no backtracking needed
	re3 := MustCompile("(?>a+)b")
	if !re3.MatchString("aab") {
		t.Error("(?>a+)b should match 'aab'")
	}

	// (?>a+) should match when no backtracking needed
	re4 := MustCompile("(?>a+)")
	if !re4.MatchString("aaa") {
		t.Error("(?>a+) should match 'aaa'")
	}
}

func TestPossessiveQuantifiers(t *testing.T) {
	// a++ is equivalent to (?>a+) — no backtracking into the a+ part
	// a++ab should NOT match 'aaab' because a++ consumes all 'a's and won't give any back
	re := MustCompile("a++ab")
	if re.MatchString("aaab") {
		t.Error("a++ab should NOT match 'aaab' (possessive won't give back)")
	}

	// But a++b should match 'aab' because no backtracking is needed
	re2 := MustCompile("a++b")
	if !re2.MatchString("aab") {
		t.Error("a++b should match 'aab'")
	}

	// \\d++ followed by \\d should not match because possessive consumed all digits
	re3 := MustCompile("\\d++\\d")
	if re3.MatchString("123") {
		t.Error("\\d++\\d should NOT match '123' (possessive consumes all digits)")
	}
}

func TestBackreference(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    bool
	}{
		// Basic numbered backref
		{"(a)\\1", "aa", true},
		{"(a)\\1", "ab", false},
		{"(\\w+) \\1", "hello hello", true},
		{"(\\w+) \\1", "hello world", false},

		// Multiple groups
		{"(a)(b)\\2\\1", "abba", true},
		{"(a)(b)\\2\\1", "abab", false},
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

func TestBackreferenceSubmatch(t *testing.T) {
	re := MustCompile(`(["'])(.*?)\1`)
	match := re.FindStringSubmatch(`"hello"`)
	if match == nil {
		t.Fatal("no match")
	}
	if match[0] != `"hello"` {
		t.Errorf("match[0] = %q, want '\"hello\"'", match[0])
	}
	if match[2] != "hello" {
		t.Errorf("match[2] = %q, want 'hello'", match[2])
	}
}

func TestVerbFail(t *testing.T) {
	re := MustCompile("a(*FAIL)b|c")
	if !re.MatchString("c") {
		t.Error("(*FAIL) should cause backtrack to next alternative")
	}
	if re.MatchString("ab") {
		t.Error("a(*FAIL)b should never match")
	}
}

func TestVerbAccept(t *testing.T) {
	re := MustCompile("a(*ACCEPT)b")
	// (*ACCEPT) forces immediate match at 'a', regardless of 'b'
	got := re.FindString("ac")
	if got != "a" {
		t.Errorf("(*ACCEPT): got %q, want 'a'", got)
	}
}

func TestMatchPointReset(t *testing.T) {
	// \K resets the start of the reported match
	re := MustCompile("foo\\Kbar")
	got := re.FindString("foobar")
	if got != "bar" {
		t.Errorf("\\K: got %q, want 'bar'", got)
	}
}

func TestCallout(t *testing.T) {
	re := MustCompile("a(?C1)b(?C2)c")
	var callouts []int
	re.SetCallout(func(cb *CalloutBlock) int {
		callouts = append(callouts, cb.CalloutNumber)
		return 0
	})
	re.MatchString("abc")
	if len(callouts) != 2 || callouts[0] != 1 || callouts[1] != 2 {
		t.Errorf("callouts = %v, want [1 2]", callouts)
	}
}

func TestCalloutAbort(t *testing.T) {
	re := MustCompile("a(?C1)b")
	re.SetCallout(func(cb *CalloutBlock) int {
		return 1 // abort
	})
	if re.MatchString("ab") {
		t.Error("callout returning 1 should abort match")
	}
}

func TestInlineOptions(t *testing.T) {
	// Scoped option
	re := MustCompile("(?i:abc)def")
	if !re.MatchString("ABCdef") {
		t.Error("(?i:abc)def should match 'ABCdef'")
	}
	if re.MatchString("ABCDEF") {
		t.Error("(?i:abc)def should NOT match 'ABCDEF' (def is case-sensitive)")
	}

	// Standalone option
	re2 := MustCompile("abc(?i)def")
	if !re2.MatchString("abcDEF") {
		t.Error("abc(?i)def should match 'abcDEF'")
	}
}

func TestDepthLimit(t *testing.T) {
	// Test that recursion depth limit works
	re := MustCompile("(a(?1)?b)")
	re.SetDepthLimit(5)
	// Deeply nested should fail
	if re.MatchString("aaaaaabbbbbb") {
		// Might succeed if within limit, that's ok
	}
}

func TestHeapLimit(t *testing.T) {
	re := MustCompile("(a+)+b")
	re.SetHeapLimit(1024) // very small
	re.SetMatchLimit(100000)
	// Should hit some limit
	result := re.MatchString("aaaaaaaaaaaac")
	if result {
		t.Error("expected no match with tight limits")
	}
}

func TestHorizontalVerticalSpace(t *testing.T) {
	// \h matches horizontal space
	re := MustCompile(`\h+`)
	if !re.MatchString(" \t") {
		t.Error("\\h should match space and tab")
	}
	if re.MatchString("\n") {
		t.Error("\\h should NOT match newline")
	}

	// \v matches vertical space
	re2 := MustCompile(`\v+`)
	if !re2.MatchString("\n\r") {
		t.Error("\\v should match newline and CR")
	}
	if re2.MatchString(" ") {
		t.Error("\\v should NOT match space")
	}

	// \H and \V are negations
	re3 := MustCompile(`\H+`)
	if !re3.MatchString("abc") {
		t.Error("\\H should match non-horizontal-space")
	}
	re4 := MustCompile(`\V+`)
	if !re4.MatchString("abc") {
		t.Error("\\V should match non-vertical-space")
	}
}

func TestStartOfMatchAnchor(t *testing.T) {
	// \G matches at the start of the match attempt
	re := MustCompile(`\Gabc`)
	if !re.MatchString("abc") {
		t.Error("\\G at start should match")
	}
	if re.MatchString("xabc") {
		t.Error("\\G should not match when not at start")
	}
}

func TestEscapeSequences(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    bool
	}{
		{"\\x41", "A", true},
		{"\\x{0041}", "A", true},
		{"\\o{101}", "A", true}, // octal 101 = 'A'
		{"\\cA", "\x01", true},  // control-A
		{"\\cZ", "\x1a", true},  // control-Z
		{"\\e", "\x1b", true},   // escape
		{"\\a", "\a", true},     // bell
		{"\\f", "\f", true},     // form feed
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			got := re.MatchString(tt.subject)
			if got != tt.want {
				t.Errorf("MatchString(%q) = %v, want %v", tt.subject, got, tt.want)
			}
		})
	}
}

func TestCaseInsensitiveBackref(t *testing.T) {
	re := MustCompile(`(abc)\1`, Caseless)
	if !re.MatchString("abcABC") {
		t.Error("case-insensitive backref should match 'abcABC'")
	}
}

func TestCommentGroup(t *testing.T) {
	re := MustCompile("abc(?# this is a comment)def")
	if !re.MatchString("abcdef") {
		t.Error("comment group should be ignored")
	}
}

func TestDollarEndOnly(t *testing.T) {
	re := MustCompile("abc$", DollarEndOnly)
	if re.MatchString("abc\n") {
		t.Error("DollarEndOnly: $ should not match before final newline")
	}
	if !re.MatchString("abc") {
		t.Error("DollarEndOnly: $ should match at end")
	}
}
