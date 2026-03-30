package gopcre2

import (
	"testing"
)

func TestCompileBasic(t *testing.T) {
	tests := []struct {
		pattern string
		flags   Flag
		wantErr bool
	}{
		{"abc", 0, false},
		{"a.c", 0, false},
		{"^abc$", 0, false},
		{"a|b|c", 0, false},
		{"(abc)", 0, false},
		{"(?:abc)", 0, false},
		{"a*b+c?", 0, false},
		{"a{2,5}", 0, false},
		{"a{3}", 0, false},
		{"a{2,}", 0, false},
		{"\\d+", 0, false},
		{"\\w\\s\\D", 0, false},
		{"[abc]", 0, false},
		{"[^a-z]", 0, false},
		{"[a-z0-9]", 0, false},
		{"\\n\\r\\t", 0, false},
		{"\\x41", 0, false},
		{"\\x{41}", 0, false},
		{"(a)(b)\\1", 0, false},
		{"(?<name>abc)", 0, false},
		{"(?P<name>abc)", 0, false},
		{"(?'name'abc)", 0, false},
		{"(?=abc)", 0, false},
		{"(?!abc)", 0, false},
		{"(?<=abc)", 0, false},
		{"(?<!abc)", 0, false},
		{"(?>abc)", 0, false},
		{"a*?", 0, false},
		{"a*+", 0, false},
		{"a++", 0, false},
		{"a?+", 0, false},
		{"\\b\\B", 0, false},
		{"\\A\\z\\Z", 0, false},
		{"\\K", 0, false},
		{"(?i)abc", 0, false},
		{"(?i:abc)", 0, false},
		{".", DotAll, false},
		{"^.+$", Multiline, false},
		// Errors
		{"(", 0, true},           // unterminated group
		{"[", 0, true},           // unterminated character class
		{"\\", 0, true},          // trailing backslash
		{"(?<name", 0, true},     // unterminated named group
		{"\\p{", 0, true},        // unterminated property
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			_, err := Compile(tt.pattern, tt.flags)
			if tt.wantErr && err == nil {
				t.Errorf("Compile(%q) expected error, got nil", tt.pattern)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Compile(%q) unexpected error: %v", tt.pattern, err)
			}
		})
	}
}

func TestMatchString(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		flags   Flag
		want    bool
	}{
		// Literals
		{"abc", "abc", 0, true},
		{"abc", "xabcx", 0, true},
		{"abc", "ab", 0, false},
		{"abc", "ABC", 0, false},

		// Case insensitive
		{"abc", "ABC", Caseless, true},
		{"abc", "AbC", Caseless, true},

		// Dot
		{"a.c", "abc", 0, true},
		{"a.c", "aXc", 0, true},
		{"a.c", "a\nc", 0, false},
		{"a.c", "a\nc", DotAll, true},

		// Anchors
		{"^abc", "abc", 0, true},
		{"^abc", "xabc", 0, false},
		{"abc$", "abc", 0, true},
		{"abc$", "abcx", 0, false},
		{"^abc$", "abc", 0, true},
		{"^abc$", "abc\n", 0, true}, // $ matches before final \n
		{"\\Aabc", "abc", 0, true},
		{"\\Aabc", "xabc", 0, false},
		{"abc\\z", "abc", 0, true},
		{"abc\\z", "abc\n", 0, false},
		{"abc\\Z", "abc", 0, true},
		{"abc\\Z", "abc\n", 0, true},

		// Alternation
		{"cat|dog", "cat", 0, true},
		{"cat|dog", "dog", 0, true},
		{"cat|dog", "fish", 0, false},
		{"a|b|c", "b", 0, true},
		{"a|b|c", "d", 0, false},

		// Quantifiers
		{"ab*c", "ac", 0, true},
		{"ab*c", "abc", 0, true},
		{"ab*c", "abbc", 0, true},
		{"ab+c", "ac", 0, false},
		{"ab+c", "abc", 0, true},
		{"ab+c", "abbc", 0, true},
		{"ab?c", "ac", 0, true},
		{"ab?c", "abc", 0, true},
		{"ab?c", "abbc", 0, false},
		{"a{2}", "a", 0, false},
		{"a{2}", "aa", 0, true},
		{"a{2}", "aaa", 0, true},
		{"a{2,4}", "a", 0, false},
		{"a{2,4}", "aa", 0, true},
		{"a{2,4}", "aaaa", 0, true},
		{"a{2,4}", "aaaaa", 0, true}, // matches first 4
		{"a{2,}", "a", 0, false},
		{"a{2,}", "aa", 0, true},
		{"a{2,}", "aaaaaa", 0, true},

		// Groups
		{"(abc)", "abc", 0, true},
		{"(?:abc)", "abc", 0, true},
		{"(a)(b)(c)", "abc", 0, true},

		// Character classes
		{"[abc]", "a", 0, true},
		{"[abc]", "b", 0, true},
		{"[abc]", "d", 0, false},
		{"[a-z]", "m", 0, true},
		{"[a-z]", "M", 0, false},
		{"[^a-z]", "M", 0, true},
		{"[^a-z]", "m", 0, false},
		{"[a-z0-9]", "5", 0, true},

		// Character types
		{"\\d+", "123", 0, true},
		{"\\d+", "abc", 0, false},
		{"\\w+", "hello", 0, true},
		{"\\w+", "!!!", 0, false},
		{"\\s+", "  ", 0, true},
		{"\\s+", "ab", 0, false},
		{"\\D+", "abc", 0, true},
		{"\\W+", "!!!", 0, true},
		{"\\S+", "abc", 0, true},

		// Escape sequences
		{"\\n", "\n", 0, true},
		{"\\t", "\t", 0, true},
		{"\\x41", "A", 0, true},
		{"\\x{0041}", "A", 0, true},

		// Word boundary
		{"\\bword\\b", "a word here", 0, true},
		{"\\bword\\b", "sword", 0, false},
		{"\\Bword", "sword", 0, true},

		// Multiline
		{"^abc", "xxx\nabc", Multiline, true},
		{"^abc", "xxx\nabc", 0, false},
		{"abc$", "abc\nxxx", Multiline, true},

		// Empty pattern
		{"", "abc", 0, true},
		{"", "", 0, true},

		// Complex patterns
		{"(a+)(b+)", "aaabbb", 0, true},
		{"(a|b)+", "ababab", 0, true},
		{"(?:a|b)+c", "ababc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.subject, func(t *testing.T) {
			re, err := Compile(tt.pattern, tt.flags)
			if err != nil {
				t.Fatalf("Compile(%q): %v", tt.pattern, err)
			}
			got := re.MatchString(tt.subject)
			if got != tt.want {
				t.Errorf("Compile(%q).MatchString(%q) = %v, want %v", tt.pattern, tt.subject, got, tt.want)
			}
		})
	}
}

func TestFindString(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    string
	}{
		{"abc", "xxabcxx", "abc"},
		{"a+", "baaab", "aaa"},
		{"\\d+", "abc123def", "123"},
		{"[a-z]+", "123abc456", "abc"},
		{"(a)(b)", "ab", "ab"},
		{"x", "abc", ""},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			got := re.FindString(tt.subject)
			if got != tt.want {
				t.Errorf("FindString(%q) = %q, want %q", tt.subject, got, tt.want)
			}
		})
	}
}

func TestFindStringSubmatch(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    []string
	}{
		{"(a+)(b+)", "aaabbb", []string{"aaabbb", "aaa", "bbb"}},
		{"(\\w+)@(\\w+)", "user@host", []string{"user@host", "user", "host"}},
		{"(?P<first>\\w+) (?P<last>\\w+)", "John Doe", []string{"John Doe", "John", "Doe"}},
		{"(a)|(b)", "b", []string{"b", "", "b"}},
		{"x(y)?z", "xz", []string{"xz", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			got := re.FindStringSubmatch(tt.subject)
			if got == nil && tt.want != nil {
				t.Fatalf("FindStringSubmatch(%q) = nil, want %v", tt.subject, tt.want)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("FindStringSubmatch(%q) len = %d, want %d: got %v", tt.subject, len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("FindStringSubmatch(%q)[%d] = %q, want %q", tt.subject, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFindAllString(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		n       int
		want    []string
	}{
		{"\\d+", "a1b22c333", -1, []string{"1", "22", "333"}},
		{"[a-z]+", "abc123def456ghi", -1, []string{"abc", "def", "ghi"}},
		{"a", "aaa", -1, []string{"a", "a", "a"}},
		{"a", "aaa", 2, []string{"a", "a"}},
		{"x", "abc", -1, nil},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			got := re.FindAllString(tt.subject, tt.n)
			if len(got) != len(tt.want) {
				t.Fatalf("FindAllString(%q, %d) = %v, want %v", tt.subject, tt.n, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("FindAllString(%q, %d)[%d] = %q, want %q", tt.subject, tt.n, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestReplaceAllString(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		repl    string
		want    string
	}{
		{"\\d+", "abc123def456", "NUM", "abcNUMdefNUM"},
		{"(\\w+)@(\\w+)", "user@host", "$2@$1", "host@user"},
		{"a+", "aaabbb", "x", "xbbb"},
		{"(a)(b)", "ab", "$2$1", "ba"},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			got := re.ReplaceAllString(tt.subject, tt.repl)
			if got != tt.want {
				t.Errorf("ReplaceAllString(%q, %q) = %q, want %q", tt.subject, tt.repl, got, tt.want)
			}
		})
	}
}

func TestNamedCaptures(t *testing.T) {
	re := MustCompile("(?P<year>\\d{4})-(?P<month>\\d{2})-(?P<day>\\d{2})")
	if re.NumSubexp() != 3 {
		t.Errorf("NumSubexp() = %d, want 3", re.NumSubexp())
	}
	if re.SubexpIndex("year") != 1 {
		t.Errorf("SubexpIndex(year) = %d, want 1", re.SubexpIndex("year"))
	}
	if re.SubexpIndex("month") != 2 {
		t.Errorf("SubexpIndex(month) = %d, want 2", re.SubexpIndex("month"))
	}
	if re.SubexpIndex("day") != 3 {
		t.Errorf("SubexpIndex(day) = %d, want 3", re.SubexpIndex("day"))
	}
	if re.SubexpIndex("nonexistent") != -1 {
		t.Errorf("SubexpIndex(nonexistent) = %d, want -1", re.SubexpIndex("nonexistent"))
	}

	names := re.SubexpNames()
	expected := []string{"", "year", "month", "day"}
	if len(names) != len(expected) {
		t.Fatalf("SubexpNames() len = %d, want %d", len(names), len(expected))
	}
	for i := range names {
		if names[i] != expected[i] {
			t.Errorf("SubexpNames()[%d] = %q, want %q", i, names[i], expected[i])
		}
	}

	match := re.FindStringSubmatch("2026-03-30")
	if match == nil {
		t.Fatal("no match")
	}
	if match[1] != "2026" || match[2] != "03" || match[3] != "30" {
		t.Errorf("match = %v, want [2026 03 30]", match[1:])
	}
}

func TestFindStringIndex(t *testing.T) {
	re := MustCompile("\\d+")
	idx := re.FindStringIndex("abc123def")
	if idx == nil {
		t.Fatal("no match")
	}
	if idx[0] != 3 || idx[1] != 6 {
		t.Errorf("FindStringIndex = %v, want [3 6]", idx)
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		n       int
		want    []string
	}{
		{"\\s+", "hello world foo", -1, []string{"hello", "world", "foo"}},
		{",", "a,b,c", -1, []string{"a", "b", "c"}},
		{",", "a,b,c", 2, []string{"a", "b,c"}},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			got := re.Split(tt.subject, tt.n)
			if len(got) != len(tt.want) {
				t.Fatalf("Split(%q, %d) = %v, want %v", tt.subject, tt.n, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Split(%q, %d)[%d] = %q, want %q", tt.subject, tt.n, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestMatchLimit(t *testing.T) {
	// This pattern causes catastrophic backtracking
	re := MustCompile("(a+)+b")
	re.SetMatchLimit(1000) // very low limit

	// This should hit the limit rather than hang
	result := re.MatchString("aaaaaaaaaaaaaaaaaac")
	if result {
		t.Error("expected no match (limit should be hit)")
	}
}

func TestMustCompilePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from MustCompile with invalid pattern")
		}
	}()
	MustCompile("(")
}

func TestLazyQuantifiers(t *testing.T) {
	re := MustCompile("a+?")
	got := re.FindString("aaa")
	if got != "a" {
		t.Errorf("lazy a+? on 'aaa': got %q, want 'a'", got)
	}

	re2 := MustCompile("<.+?>")
	got2 := re2.FindString("<b>text</b>")
	if got2 != "<b>" {
		t.Errorf("lazy <.+?> got %q, want '<b>'", got2)
	}
}

func TestUngreedyFlag(t *testing.T) {
	re := MustCompile("a+", Ungreedy)
	got := re.FindString("aaa")
	if got != "a" {
		t.Errorf("Ungreedy a+ on 'aaa': got %q, want 'a'", got)
	}
}

func TestDotAll(t *testing.T) {
	re := MustCompile("a.b", DotAll)
	if !re.MatchString("a\nb") {
		t.Error("DotAll a.b should match 'a\\nb'")
	}
}

func TestLiteralBrace(t *testing.T) {
	// { that doesn't form a valid repeat should be treated as literal
	re := MustCompile("a{b")
	if !re.MatchString("a{b") {
		t.Error("a{b should match literal 'a{b'")
	}
}

func TestString(t *testing.T) {
	re := MustCompile("abc")
	if re.String() != "abc" {
		t.Errorf("String() = %q, want 'abc'", re.String())
	}
}

func TestFindAllSubmatch(t *testing.T) {
	re := MustCompile("(\\w+)")
	got := re.FindAllStringSubmatch("hello world", -1)
	if len(got) != 2 {
		t.Fatalf("FindAllStringSubmatch len = %d, want 2", len(got))
	}
	if got[0][0] != "hello" || got[0][1] != "hello" {
		t.Errorf("match 0 = %v, want [hello hello]", got[0])
	}
	if got[1][0] != "world" || got[1][1] != "world" {
		t.Errorf("match 1 = %v, want [world world]", got[1])
	}
}

func TestReplaceAllStringFunc(t *testing.T) {
	re := MustCompile("\\w+")
	got := re.ReplaceAllStringFunc("hello world", func(s string) string {
		return "[" + s + "]"
	})
	if got != "[hello] [world]" {
		t.Errorf("ReplaceAllStringFunc got %q, want '[hello] [world]'", got)
	}
}

func TestByteSliceMethods(t *testing.T) {
	re := MustCompile("\\d+")
	b := []byte("abc123def456")

	if !re.Match(b) {
		t.Error("Match(b) should be true")
	}

	found := re.Find(b)
	if string(found) != "123" {
		t.Errorf("Find = %q, want '123'", found)
	}

	all := re.FindAll(b, -1)
	if len(all) != 2 || string(all[0]) != "123" || string(all[1]) != "456" {
		t.Errorf("FindAll = %v", all)
	}

	replaced := re.ReplaceAll(b, []byte("NUM"))
	if string(replaced) != "abcNUMdefNUM" {
		t.Errorf("ReplaceAll = %q", replaced)
	}
}

func TestEmptyMatches(t *testing.T) {
	re := MustCompile("")
	all := re.FindAllString("abc", -1)
	// Empty pattern matches at every position
	if len(all) < 3 {
		t.Errorf("FindAllString with empty pattern: got %d matches, want >= 3", len(all))
	}
}

func TestPOSIXClasses(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    bool
	}{
		{"[[:digit:]]", "5", true},
		{"[[:digit:]]", "a", false},
		{"[[:alpha:]]", "a", true},
		{"[[:alpha:]]", "5", false},
		{"[[:alnum:]]", "a", true},
		{"[[:alnum:]]", "5", true},
		{"[[:alnum:]]", "!", false},
		{"[[:upper:]]", "A", true},
		{"[[:upper:]]", "a", false},
		{"[[:lower:]]", "a", true},
		{"[[:lower:]]", "A", false},
		{"[[:space:]]", " ", true},
		{"[[:space:]]", "a", false},
		{"[[:xdigit:]]", "f", true},
		{"[[:xdigit:]]", "g", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.subject, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			got := re.MatchString(tt.subject)
			if got != tt.want {
				t.Errorf("MatchString(%q) = %v, want %v", tt.subject, got, tt.want)
			}
		})
	}
}
