package gopcre2

import "testing"

func TestLexerBasicTokens(t *testing.T) {
	lex := newLexer("abc", 0)
	tokens, err := lex.tokenize()
	if err != nil {
		t.Fatal(err)
	}
	// Should be: 'a', 'b', 'c', EOF
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d", len(tokens))
	}
	for i := 0; i < 3; i++ {
		if tokens[i].Kind != TokLiteral {
			t.Errorf("token %d: got kind %d, want TokLiteral", i, tokens[i].Kind)
		}
	}
	if tokens[3].Kind != TokEOF {
		t.Error("last token should be EOF")
	}
}

func TestLexerQuantifiers(t *testing.T) {
	lex := newLexer("a*b+c?d{2,5}", 0)
	tokens, err := lex.tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expected := []TokenKind{
		TokLiteral, TokStar,
		TokLiteral, TokPlus,
		TokLiteral, TokQuestion,
		TokLiteral, TokRepeat,
		TokEOF,
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, exp := range expected {
		if tokens[i].Kind != exp {
			t.Errorf("token %d: got kind %d, want %d", i, tokens[i].Kind, exp)
		}
	}
}

func TestLexerEscapes(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"\\d\\w\\s", false},
		{"\\n\\r\\t", false},
		{"\\x41", false},
		{"\\x{0041}", false},
		{"\\o{101}", false},
		{"\\cA", false},
		{"\\p{Lu}", false},
		{"\\P{Nd}", false},
		{"\\pL", false},
		{"\\b\\B", false},
		{"\\A\\z\\Z", false},
		{"\\K", false},
		{"\\1", false},
		{"\\g{1}", false},
		{"\\k<name>", false},
		{"\\", true},  // trailing backslash
		{"\\Q", true}, // unrecognized escape
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lex := newLexer(tt.input, 0)
			_, err := lex.tokenize()
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLexerGroups(t *testing.T) {
	tests := []struct {
		input    string
		firstTok TokenKind
	}{
		{"(abc)", TokGroupOpen},
		{"(?:abc)", TokNonCapture},
		{"(?<name>abc)", TokNamedCapture},
		{"(?P<name>abc)", TokNamedCapture},
		{"(?'name'abc)", TokNamedCapture},
		{"(?>abc)", TokAtomicGroup},
		{"(?=abc)", TokLookahead},
		{"(?!abc)", TokNegLookahead},
		{"(?<=abc)", TokLookbehind},
		{"(?<!abc)", TokNegLookbehind},
		{"(?|abc)", TokBranchReset},
		{"(?i)", TokInlineOption},
		{"(?i:abc)", TokInlineOption},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lex := newLexer(tt.input, 0)
			tokens, err := lex.tokenize()
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if tokens[0].Kind != tt.firstTok {
				t.Errorf("first token: got %d, want %d", tokens[0].Kind, tt.firstTok)
			}
		})
	}
}

func TestLexerVerbs(t *testing.T) {
	tests := []struct {
		input    string
		verbKind VerbKind
	}{
		{"(*ACCEPT)", VerbAccept},
		{"(*FAIL)", VerbFail},
		{"(*F)", VerbFail},
		{"(*COMMIT)", VerbCommit},
		{"(*PRUNE)", VerbPrune},
		{"(*SKIP)", VerbSkip},
		{"(*SKIP:name)", VerbSkipName},
		{"(*THEN)", VerbThen},
		{"(*MARK:name)", VerbMark},
		{"(*:name)", VerbMark},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lex := newLexer(tt.input, 0)
			tokens, err := lex.tokenize()
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if tokens[0].Kind != TokVerb {
				t.Fatalf("first token kind: got %d, want TokVerb", tokens[0].Kind)
			}
			if tokens[0].VerbType != tt.verbKind {
				t.Errorf("verb kind: got %d, want %d", tokens[0].VerbType, tt.verbKind)
			}
		})
	}
}

func TestLexerCharClass(t *testing.T) {
	lex := newLexer("[a-z]", 0)
	tokens, err := lex.tokenize()
	if err != nil {
		t.Fatal(err)
	}
	// [, 'a', -, 'z', ], EOF
	if tokens[0].Kind != TokCharClassOpen {
		t.Errorf("expected TokCharClassOpen, got %d", tokens[0].Kind)
	}

	lex2 := newLexer("[^abc]", 0)
	tokens2, err := lex2.tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if tokens2[0].Kind != TokCharClassNeg {
		t.Errorf("expected TokCharClassNeg, got %d", tokens2[0].Kind)
	}
}

func TestLexerExtendedMode(t *testing.T) {
	lex := newLexer("a b c # comment\nd", Extended)
	tokens, err := lex.tokenize()
	if err != nil {
		t.Fatal(err)
	}
	// Should ignore spaces and comments: a, b, c, d, EOF
	count := 0
	for _, tok := range tokens {
		if tok.Kind == TokLiteral {
			count++
		}
	}
	if count != 4 {
		t.Errorf("expected 4 literals in extended mode, got %d", count)
	}
}

func TestLexerRepeatFormats(t *testing.T) {
	tests := []struct {
		input string
		min   int
		max   int
	}{
		{"a{3}", 3, 3},
		{"a{2,5}", 2, 5},
		{"a{2,}", 2, -1},
		{"a{0,1}", 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lex := newLexer(tt.input, 0)
			tokens, err := lex.tokenize()
			if err != nil {
				t.Fatal(err)
			}
			// Find the TokRepeat token
			for _, tok := range tokens {
				if tok.Kind == TokRepeat {
					if tok.Num != tt.min || tok.Num2 != tt.max {
						t.Errorf("got {%d,%d}, want {%d,%d}", tok.Num, tok.Num2, tt.min, tt.max)
					}
					return
				}
			}
			t.Error("no TokRepeat found")
		})
	}
}
