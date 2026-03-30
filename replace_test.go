package gopcre2

import (
	"testing"
)

func TestReplaceNamedGroups(t *testing.T) {
	re := MustCompile("(?P<y>\\d{4})-(?P<m>\\d{2})-(?P<d>\\d{2})")
	got := re.ReplaceAllString("2026-03-30", "${d}/${m}/${y}")
	if got != "30/03/2026" {
		t.Errorf("got %q, want '30/03/2026'", got)
	}
}

func TestReplaceZeroRef(t *testing.T) {
	re := MustCompile("\\w+")
	got := re.ReplaceAllString("hello world", "[$0]")
	if got != "[hello] [world]" {
		t.Errorf("got %q, want '[hello] [world]'", got)
	}
}

func TestReplaceDollarDollar(t *testing.T) {
	re := MustCompile("\\d+")
	got := re.ReplaceAllString("price: 100", "$$$$")
	if got != "price: $$" {
		t.Errorf("got %q, want 'price: $$'", got)
	}
}

func TestReplaceAllFunc(t *testing.T) {
	re := MustCompile("\\d+")
	got := re.ReplaceAllFunc([]byte("a1b22c333"), func(b []byte) []byte {
		return []byte("[" + string(b) + "]")
	})
	if string(got) != "a[1]b[22]c[333]" {
		t.Errorf("got %q", string(got))
	}
}
