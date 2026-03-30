package gopcre2

import "testing"

func TestParserLiteral(t *testing.T) {
	tokens, _ := newLexer("abc", 0).tokenize()
	p := newParser(tokens, 0)
	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != NdConcat {
		t.Fatalf("expected NdConcat, got %d", node.Kind)
	}
	if len(node.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.Children))
	}
	for i, ch := range node.Children {
		if ch.Kind != NdLiteral {
			t.Errorf("child %d: expected NdLiteral, got %d", i, ch.Kind)
		}
	}
}

func TestParserAlternation(t *testing.T) {
	tokens, _ := newLexer("a|b|c", 0).tokenize()
	p := newParser(tokens, 0)
	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != NdAlternate {
		t.Fatalf("expected NdAlternate, got %d", node.Kind)
	}
	if len(node.Children) != 3 {
		t.Fatalf("expected 3 alternatives, got %d", len(node.Children))
	}
}

func TestParserRepeat(t *testing.T) {
	tokens, _ := newLexer("a+", 0).tokenize()
	p := newParser(tokens, 0)
	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != NdRepeat {
		t.Fatalf("expected NdRepeat, got %d", node.Kind)
	}
	if node.Min != 1 || node.Max != -1 {
		t.Errorf("expected {1,-1}, got {%d,%d}", node.Min, node.Max)
	}
	if node.Greedy != Greedy {
		t.Error("expected Greedy")
	}
}

func TestParserLazyRepeat(t *testing.T) {
	tokens, _ := newLexer("a+?", 0).tokenize()
	p := newParser(tokens, 0)
	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != NdRepeat {
		t.Fatalf("expected NdRepeat, got %d", node.Kind)
	}
	if node.Greedy != Lazy {
		t.Error("expected Lazy")
	}
}

func TestParserPossessiveRepeat(t *testing.T) {
	tokens, _ := newLexer("a++", 0).tokenize()
	p := newParser(tokens, 0)
	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != NdRepeat {
		t.Fatalf("expected NdRepeat, got %d", node.Kind)
	}
	if node.Greedy != Possessive {
		t.Error("expected Possessive")
	}
}

func TestParserCapture(t *testing.T) {
	tokens, _ := newLexer("(abc)", 0).tokenize()
	p := newParser(tokens, 0)
	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != NdCapture {
		t.Fatalf("expected NdCapture, got %d", node.Kind)
	}
	if node.Index != 1 {
		t.Errorf("expected group 1, got %d", node.Index)
	}
}

func TestParserNamedCapture(t *testing.T) {
	tokens, _ := newLexer("(?<foo>abc)", 0).tokenize()
	p := newParser(tokens, 0)
	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != NdNamedCapture {
		t.Fatalf("expected NdNamedCapture, got %d", node.Kind)
	}
	if node.Name != "foo" {
		t.Errorf("expected name 'foo', got %q", node.Name)
	}
}

func TestParserNested(t *testing.T) {
	tokens, _ := newLexer("(a(b)c)", 0).tokenize()
	p := newParser(tokens, 0)
	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != NdCapture {
		t.Fatalf("expected NdCapture, got %d", node.Kind)
	}
	// Inner should be concat: literal 'a', capture(literal 'b'), literal 'c'
	inner := node.Children[0]
	if inner.Kind != NdConcat {
		t.Fatalf("expected NdConcat, got %d", inner.Kind)
	}
	if len(inner.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(inner.Children))
	}
	if inner.Children[1].Kind != NdCapture {
		t.Error("expected nested capture")
	}
}

func TestParserEmpty(t *testing.T) {
	tokens, _ := newLexer("", 0).tokenize()
	p := newParser(tokens, 0)
	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != NdEmpty {
		t.Fatalf("expected NdEmpty, got %d", node.Kind)
	}
}

func TestParserMultipleCaptures(t *testing.T) {
	tokens, _ := newLexer("(a)(b)(c)", 0).tokenize()
	p := newParser(tokens, 0)
	node, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != NdConcat {
		t.Fatalf("expected NdConcat, got %d", node.Kind)
	}
	if len(node.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.Children))
	}
	for i, ch := range node.Children {
		if ch.Kind != NdCapture {
			t.Errorf("child %d: expected NdCapture, got %d", i, ch.Kind)
		}
		if ch.Index != i+1 {
			t.Errorf("child %d: expected group %d, got %d", i, i+1, ch.Index)
		}
	}
}
