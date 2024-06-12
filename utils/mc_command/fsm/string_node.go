package fsm

import "strings"

// the node it self is a fsm machine
type MachineNodeWrapper struct {
	Next  AutomataNode
	Else  AutomataNode
	Logic func(end, fail AutomataNode) AutomataNode
}

func (n *MachineNodeWrapper) Do(r TextReader, currentRead *Text) {
	back := r.Snapshot()
	_end := &NodeEnd{}
	_fail := &NodeBasic{}
	n.Logic(_end, _fail).Do(r, &Text{})
	if _end.Accessed {
		if _fail.Accessed {
			panic("illegal fsm define")
		}
		currentRead.Text += _end.Read
		n.Next.Do(r, currentRead)
	} else {
		back()
		n.Else.Do(r, currentRead)
	}
}

func DoLogic(r TextReader, logic func(end, fail AutomataNode) AutomataNode) (bool, string) {
	back := r.Snapshot()
	end := &NodeEnd{}
	fail := &NodeBasic{}
	logic(end, fail).Do(r, &Text{})
	if end.Accessed {
		if fail.Accessed {
			panic("illegal fsm define")
		}
		return true, end.Read
	} else {
		back()
		return false, ""
	}
}

// Start -> [1.WhiteSpace]
// [1] -> [2.WhiteSpaceOrElse]
// [2] -IfWhiteSpace-> [2]
// [2] -Else-> Next
func WhiteSpaceStringLogic(end, fail AutomataNode) AutomataNode {
	_1 := &NodeCharWhiteSpace{}
	_2 := &NodeCharWhiteSpace{}
	_1.Else = fail
	_1.Next = _2
	_2.Next = _2
	_2.Else = end
	return _1
}

func MakeWhiteSpaceStringNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: WhiteSpaceStringLogic,
	}
}

// Start -> [1.NonWhiteSpace]
// [1] -> [2.NonWhiteSpaceOrElse]
// [2] -IfNonWhiteSpace-> [2]
// [2] -Else-> Next
func NonWhiteSpaceStringLogic(end, fail AutomataNode) AutomataNode {
	_1 := &NodeCharNonWhiteSpace{}
	_2 := &NodeCharNonWhiteSpace{}
	_1.Else = fail
	_1.Next = _2
	_2.Next = _2
	_2.Else = end
	return _1
}

func MakeNonWhiteSpaceStringNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: WhiteSpaceStringLogic,
	}
}

// Start -> [1.Number]
// [1] -> [2.NumberOrElse]
// [2] -IfNumber-> [2]
// [2] -Else-> Next
// "09876"
func NumberStringLogic(end, fail AutomataNode) AutomataNode {
	_1 := &NodeCharNumber{}
	_2 := &NodeCharNumber{}
	_1.Else = fail
	_1.Next = _2
	_2.Next = _2
	_2.Else = end
	return _1
}

func MakeNumberStringNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: NumberStringLogic,
	}
}

// Start -> [1.SignOrElse]
// [1] -IfSign-> [2.NumberString]
// [1] -IfElse-> [2.NumberString]
// [2] -> Next
// "123", "-123", "+123"
func SignedIntegerLogic(end, fail AutomataNode) AutomataNode {
	_1 := &NodePosOrNegSign{}
	_2 := MakeNumberStringNode()
	_1.Else = _2
	_1.NextNegSign = _2
	_1.NextPosSign = _2
	_2.Else = fail
	_2.Next = end
	return _1
}

func MakeSignedIntegerStringNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: SignedIntegerLogic,
	}
}

// Start -> [1.NumberStringOrElse]
// [1] -IfNumberString-> [2.DotOrElse]
// [1] -IfElse-> [3.Dot]
// [2] -IfDot-> [4.NumberStringOrElse]
// [2] -IfElse-> Next
// [3] -> [5.NumberString]
// [4] -> Next
// [5] -> Next
// "123" ".123" "123.123" "123."
func SimpleUnsignedFloatLogic(end, fail AutomataNode) AutomataNode {
	_1 := MakeNumberStringNode()
	_2 := &NodeDot{}
	_3 := &NodeDot{}
	_4 := MakeNumberStringNode()
	_5 := MakeNumberStringNode()
	_1.Next = _2
	_1.Else = _3
	_2.Next = _4
	_2.Else = end
	_3.Next = _5
	_3.Else = fail
	_4.Next = end
	_4.Else = end
	_5.Next = end
	_5.Else = fail
	return _1
}

func MakeSimpleUnsignedFloatStringNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: SimpleUnsignedFloatLogic,
	}
}

// Start -> [1.SignOrElse]
// [1] -IfSign-> [2.UnsignedFloat]
// [1] -IfElse-> [2.UnsignedFloat]
// [2] -> Next
// "123" ".123" "123.123" "123."
// "+123" "+.123" "+123.123" "+123."
// "-123" "-.123" "-123.123" "-123."
func SignedFloatLogic(end, fail AutomataNode) AutomataNode {
	_1 := &NodePosOrNegSign{}
	_2 := MakeSimpleUnsignedFloatStringNode()
	_1.Else = _2
	_1.NextNegSign = _2
	_1.NextPosSign = _2
	_2.Else = fail
	_2.Next = end
	return _1
}

func MakeSignedFloatStringNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: SignedFloatLogic,
	}
}

// Start -> [1."]
// [1] ->[2."or\orElse]
// [2] -IfElse-> [3.Any]
// [2] -if"->Next
// [2] -if\-> [3]
// [3] -> [2]
// "123" "12\"34" "\""
func DoubleQuoteStringLogic(end, fail AutomataNode) AutomataNode {
	_1 := &DoubleQuote{}
	_2 := &DoubleQuoteOrSlash{}
	_3 := &NodeCharAny{}
	_1.Else = fail
	_1.Next = _2
	_2.Else = _3
	_2.NextQuote = end
	_2.NextSlash = _3
	_3.Next = _2
	_3.Else = fail
	return _1
}

func MakeDoubleQuoteStringNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: DoubleQuoteStringLogic,
	}
}

// Start -> [1.']
// [1] ->[2.'or\orElse]
// [2] -IfElse-> [3.Any]
// [2] -if'->Next
// [2] -if\-> [3]
// [3] -> [2]
// "123" "1 2\'3 4" "\""
func SingleQuoteStringLogic(end, fail AutomataNode) AutomataNode {
	_1 := &SingleQuote{}
	_2 := &SingleQuoteOrSlash{}
	_3 := &NodeCharAny{}
	_1.Else = fail
	_1.Next = _2
	_2.Else = _3
	_2.NextQuote = end
	_2.NextSlash = _3
	_3.Next = _2
	_3.Else = fail
	return _1
}

func MakeSingleQuoteStringNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: SingleQuoteStringLogic,
	}
}

// Start -> [1.QuoteOrNonExcept]
// [1] -IfSingleQuote(noRead)-> [2.SingleQuoteString]
// [1] -IfDoubleQuote(noRead)-> [3.DoubleQuoteString]
// [1] -IfNonExcept-> [4.QuoteOrNonExceptOrElse]
// [2] -> [4]
// [3] -> [4]
// [4] -IfSingleQuote(noRead)-> [2]
// [4] -IfDoubleQuote(noRead)-> [3]
// [4] -IfNonExcept-> [4]
// [4] -IfElse-> Next
// when Except = " \t\n\r"
// `123"4 5 6"789` `123'456'789` `'123'` `"123"` `123`
type quoteOrNonExcept struct {
	NextSingleQuoteNoRead AutomataNode
	NextDoubleQuoteNoRead AutomataNode
	NextNonExcept         AutomataNode
	Except                string
	Else                  AutomataNode
}

func (n *quoteOrNonExcept) Do(r TextReader, currentRead *Text) {
	if r.Current() == "" {
		n.Else.Do(r, currentRead)
		return
	}
	if r.Current() == `"` {
		n.NextDoubleQuoteNoRead.Do(r, currentRead)
		return
	}
	if r.Current() == `'` {
		n.NextSingleQuoteNoRead.Do(r, currentRead)
		return
	}
	if !strings.ContainsAny(r.Current(), n.Except) {
		currentRead.Text += r.CurrentThenNext()
		n.NextNonExcept.Do(r, currentRead)
		return
	}
	n.Else.Do(r, currentRead)
}
func MakeQuoteMixtureNonExceptStringLogic(except string) func(end, fail AutomataNode) AutomataNode {
	return func(end, fail AutomataNode) AutomataNode {
		_1 := &quoteOrNonExcept{Except: except}
		_2 := MakeSingleQuoteStringNode()
		_3 := MakeDoubleQuoteStringNode()
		_4 := &quoteOrNonExcept{Except: except}
		_1.NextSingleQuoteNoRead = _2
		_1.NextDoubleQuoteNoRead = _3
		_1.NextNonExcept = _4
		_1.Else = fail
		_2.Next = _4
		_2.Else = fail
		_3.Next = _4
		_3.Else = fail
		_4.NextSingleQuoteNoRead = _2
		_4.NextDoubleQuoteNoRead = _3
		_4.NextNonExcept = _4
		_4.Else = end
		return _1
	}
}

// except =" \n\r\t"
func MakeQuoteMixtureNonExceptStringNode(except string) *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: MakeQuoteMixtureNonExceptStringLogic(except),
	}
}

// [This]=
// Start -> [1.`[`]
// [1] -> [2.`]`or`[`orElse]
// [2] -If`]`-> Next
// [2] -If`[`(noRead) -> [3.This]
// [2] -ifElse-> [4.QuoteMixtureExcept`[]`orElse]
// [3] -> [2]
// [4] -> [2]
// `[]` `[abc]` `[a[bc]d]` `[12[a[x]b]cd[45]67]`
type nodeBracketLorR struct {
	LB, RB               string
	NextLBNoRead, NextRB AutomataNode
	Else                 AutomataNode
	EOF                  AutomataNode
}

func (n *nodeBracketLorR) Do(r TextReader, currentRead *Text) {
	if r.Current() == "" {
		n.EOF.Do(r, currentRead)
		return
	}
	if r.Current() == n.LB {
		n.NextLBNoRead.Do(r, currentRead)
		return
	}
	if r.Current() == n.RB {
		currentRead.Text += r.CurrentThenNext()
		n.NextRB.Do(r, currentRead)
		return
	}
	n.Else.Do(r, currentRead)
}

func MakeMiddleBracketStringNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: MiddleBracketStringLogic,
	}
}

func MiddleBracketStringLogic(end, fail AutomataNode) AutomataNode {
	_1 := &NodeJumpCharIf{JumpChar: JumpChar{"[", nil}}
	_2 := &nodeBracketLorR{LB: `[`, RB: `]`}
	_4 := MakeQuoteMixtureNonExceptStringNode("[]")
	_3 := &NodeLazy{Make: func() AutomataNode {
		r := MakeMiddleBracketStringNode()
		r.Next = _2
		r.Else = fail
		return r
	}}
	_1.Else = fail
	_1.JumpChar.Next = _2
	_2.NextRB = end
	_2.NextLBNoRead = _3
	_2.EOF = fail
	_2.Else = _4
	_4.Else = fail
	_4.Next = _2
	return _1
}

// Start -> [1.AnyOrElse]
// [1] -IfAny-> [1]
// [1] -IfElse-> Next
func ReadUntilEndLogic(end, fail AutomataNode) AutomataNode {
	_1 := &NodeCharAny{}
	_1.Next = _1
	_1.Else = end
	return _1
}

func MakeReadUntilEndNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: ReadUntilEndLogic,
	}
}

type MatchSpecificNode struct {
	Target string
	Fold   bool
	Else   AutomataNode
	Next   AutomataNode
}

func (n *MatchSpecificNode) Do(r TextReader, currentRead *Text) {
	back := r.Snapshot()
	newRead := ""
	runesToRead := len([]rune(n.Target))
	for i := 0; i < runesToRead; i++ {
		if r.Current() == "" {
			break
		} else {
			newRead += r.CurrentThenNext()
		}
	}
	if n.Fold {
		if strings.EqualFold(newRead, n.Target) {
			currentRead.Text += newRead
			n.Next.Do(r, currentRead)
			return
		} else {
			back()
			n.Else.Do(r, currentRead)
			return
		}
	} else {
		if newRead == n.Target {
			currentRead.Text += newRead
			n.Next.Do(r, currentRead)
			return
		} else {
			back()
			n.Else.Do(r, currentRead)
			return
		}
	}
}

func MakeMatchSpecificLogic(target string, fold bool) func(end, fail AutomataNode) AutomataNode {
	return func(end, fail AutomataNode) AutomataNode {
		_1 := &MatchSpecificNode{
			Else:   fail,
			Next:   end,
			Target: target,
			Fold:   fold,
		}
		return _1
	}
}

type MatchSpecificsNode struct {
	Targets []struct {
		Target string
		Fold   bool
		Next   AutomataNode
	}
	Else AutomataNode
}

func (n *MatchSpecificsNode) Do(r TextReader, currentRead *Text) {
	nodes := []*MatchSpecificNode{}
	for _, t := range n.Targets {
		node := &MatchSpecificNode{
			Target: t.Target,
			Fold:   t.Fold,
			Next:   t.Next,
			Else:   n.Else,
		}
		if len(nodes) != 0 {
			nodes[len(nodes)-1].Else = node
		}
		nodes = append(nodes, node)
	}
	if len(nodes) == 0 {
		n.Else.Do(r, currentRead)
	} else {
		nodes[0].Do(r, currentRead)
	}
}

func MakeMatchSpecificsLogic(targets []string, fold bool) func(end, fail AutomataNode) AutomataNode {
	return func(end, fail AutomataNode) AutomataNode {
		targetsCond := []struct {
			Target string
			Fold   bool
			Next   AutomataNode
		}{}
		for _, target := range targets {
			targetsCond = append(targetsCond, struct {
				Target string
				Fold   bool
				Next   AutomataNode
			}{target, fold, end})
		}
		_1 := &MatchSpecificsNode{
			Targets: targetsCond,
			Else:    fail,
		}
		return _1
	}
}
