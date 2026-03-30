package gopcre2

import "testing"

func TestCompilerCaptureCount(t *testing.T) {
	re := MustCompile("(a)(b)(c)")
	if re.NumSubexp() != 3 {
		t.Errorf("NumSubexp() = %d, want 3", re.NumSubexp())
	}
}

func TestCompilerNamedCaptures(t *testing.T) {
	re := MustCompile("(?<x>a)(?P<y>b)(?'z'c)")
	if re.NumSubexp() != 3 {
		t.Errorf("NumSubexp() = %d, want 3", re.NumSubexp())
	}
	names := re.SubexpNames()
	if names[1] != "x" || names[2] != "y" || names[3] != "z" {
		t.Errorf("names = %v", names)
	}
}

func TestCompilerAlternation(t *testing.T) {
	re := MustCompile("a|b|c")
	prog := re.prog
	// Should have splits for alternation
	hasSplit := false
	for _, inst := range prog.Inst {
		if inst.Op == OpSplit {
			hasSplit = true
			break
		}
	}
	if !hasSplit {
		t.Error("alternation should produce splits")
	}
}

func TestCompilerRepeat(t *testing.T) {
	tests := []struct {
		pattern  string
		hasSplit bool
	}{
		{"a*", true},
		{"a+", true},
		{"a?", true},
		{"a{3}", false}, // exactly 3: just 3 rune ops
		{"a{2,4}", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			hasSplit := false
			for _, inst := range re.prog.Inst {
				if inst.Op == OpSplit || inst.Op == OpSplitLazy {
					hasSplit = true
					break
				}
			}
			if hasSplit != tt.hasSplit {
				t.Errorf("hasSplit = %v, want %v", hasSplit, tt.hasSplit)
			}
		})
	}
}

func TestCompilerDot(t *testing.T) {
	re := MustCompile("a.b")
	hasAny := false
	for _, inst := range re.prog.Inst {
		if inst.Op == OpAnyCharNotNL {
			hasAny = true
		}
	}
	if !hasAny {
		t.Error("dot should produce OpAnyCharNotNL")
	}

	re2 := MustCompile("a.b", DotAll)
	hasAnyAll := false
	for _, inst := range re2.prog.Inst {
		if inst.Op == OpAnyChar {
			hasAnyAll = true
		}
	}
	if !hasAnyAll {
		t.Error("dot with DotAll should produce OpAnyChar")
	}
}

func TestCompilerCaseFold(t *testing.T) {
	re := MustCompile("abc", Caseless)
	hasFold := false
	for _, inst := range re.prog.Inst {
		if inst.Op == OpRuneFold {
			hasFold = true
		}
	}
	if !hasFold {
		t.Error("caseless pattern should produce OpRuneFold")
	}
}

func TestOptimizeAnchor(t *testing.T) {
	re := MustCompile("^abc")
	if !re.prog.AnchorStart {
		t.Error("^abc should be detected as anchored")
	}

	re2 := MustCompile("abc")
	if re2.prog.AnchorStart {
		t.Error("abc should not be detected as anchored")
	}

	re3 := MustCompile("\\Aabc")
	if !re3.prog.AnchorStart {
		t.Error("\\Aabc should be detected as anchored")
	}
}

func TestOptimizePrefix(t *testing.T) {
	re := MustCompile("hello.*world")
	if re.prog.Prefix != "hello" {
		t.Errorf("prefix = %q, want 'hello'", re.prog.Prefix)
	}
}
