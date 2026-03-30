package gopcre2

import "fmt"

// Program is a compiled PCRE2 pattern ready for execution by the VM.
type Program struct {
	Inst        []Inst           // instruction array
	Start       int              // index of first instruction
	NumCapture  int              // total number of capture groups (0 = whole match)
	CapNames    []string         // capture group names by index (empty string if unnamed)
	NameToGroup map[string][]int // name -> group indices (for DUPNAMES)
	Flags       Flag             // compile-time flags

	// Optimization hints
	Prefix      string // literal prefix for fast scan (empty if none)
	AnchorStart bool   // pattern is anchored at start
	AnchorEnd   bool   // pattern is anchored at end
	MinLen      int    // minimum subject length that can match

	// Safety limits. Set via Regexp.Set*Limit() methods. Inline pattern
	// directives like (*LIMIT_MATCH=N) are only honored when the
	// AllowInlineLimits flag is passed at compile time.
	MatchLimit int // 0 = use default
	DepthLimit int // 0 = use default
	HeapLimit  int // 0 = use default

	// Subroutine entry points: group index -> instruction index
	GroupEntry map[int]int
	GroupEnd   map[int]int
}

// Dump returns a human-readable representation of the program for debugging.
func (p *Program) Dump() string {
	var s string
	for i, inst := range p.Inst {
		s += fmt.Sprintf("%4d: %s\n", i, instString(inst))
	}
	return s
}

func instString(inst Inst) string {
	switch inst.Op {
	case OpMatch:
		return "Match"
	case OpFail:
		return "Fail"
	case OpNop:
		return fmt.Sprintf("Nop -> %d", inst.Out)
	case OpRune:
		return fmt.Sprintf("Rune %q -> %d", string(inst.Runes), inst.Out)
	case OpRuneFold:
		return fmt.Sprintf("RuneFold %q -> %d", string(inst.Runes), inst.Out)
	case OpCharClass:
		neg := ""
		if inst.Negate {
			neg = "^"
		}
		return fmt.Sprintf("CharClass %s%v -> %d", neg, inst.Runes, inst.Out)
	case OpAnyCharNotNL:
		return fmt.Sprintf("AnyCharNotNL -> %d", inst.Out)
	case OpAnyChar:
		return fmt.Sprintf("AnyChar -> %d", inst.Out)
	case OpCharType:
		return fmt.Sprintf("CharType(%d) -> %d", inst.N, inst.Out)
	case OpSplit:
		return fmt.Sprintf("Split -> %d, %d", inst.Out, inst.Arg)
	case OpSplitLazy:
		return fmt.Sprintf("SplitLazy -> %d, %d", inst.Out, inst.Arg)
	case OpJump:
		return fmt.Sprintf("Jump -> %d", inst.Out)
	case OpCaptureStart:
		return fmt.Sprintf("CaptureStart(%d) -> %d", inst.N, inst.Out)
	case OpCaptureEnd:
		return fmt.Sprintf("CaptureEnd(%d) -> %d", inst.N, inst.Out)
	case OpAssertBeginLine:
		return fmt.Sprintf("AssertBeginLine -> %d", inst.Out)
	case OpAssertEndLine:
		return fmt.Sprintf("AssertEndLine -> %d", inst.Out)
	case OpAssertBeginText:
		return fmt.Sprintf("AssertBeginText -> %d", inst.Out)
	case OpAssertEndText:
		return fmt.Sprintf("AssertEndText -> %d", inst.Out)
	case OpAssertEndTextOrNewline:
		return fmt.Sprintf("AssertEndTextOpt -> %d", inst.Out)
	case OpAssertWordBoundary:
		return fmt.Sprintf("AssertWordBoundary -> %d", inst.Out)
	case OpAssertNonWordBoundary:
		return fmt.Sprintf("AssertNonWordBoundary -> %d", inst.Out)
	case OpBackref:
		return fmt.Sprintf("Backref(%d) -> %d", inst.N, inst.Out)
	case OpResetMatchStart:
		return fmt.Sprintf("ResetMatchStart -> %d", inst.Out)
	case OpLookaheadStart:
		return fmt.Sprintf("LookaheadStart -> %d", inst.Out)
	case OpLookaheadEnd:
		return fmt.Sprintf("LookaheadEnd -> %d", inst.Out)
	case OpNegLookaheadStart:
		return fmt.Sprintf("NegLookaheadStart -> %d", inst.Out)
	case OpNegLookaheadEnd:
		return fmt.Sprintf("NegLookaheadEnd -> %d", inst.Out)
	case OpAtomicStart:
		return fmt.Sprintf("AtomicStart -> %d", inst.Out)
	case OpAtomicEnd:
		return fmt.Sprintf("AtomicEnd -> %d", inst.Out)
	case OpAccept:
		return "Accept"
	case OpVerbFail:
		return "VerbFail"
	default:
		return fmt.Sprintf("Op(%d) -> %d", inst.Op, inst.Out)
	}
}
