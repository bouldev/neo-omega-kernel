package fsm

import "strings"

type JumpChar struct {
	Condition string
	Next      AutomataNode
}

type NodeJumpCharIf struct {
	JumpChar
	Else AutomataNode
}

func (n *NodeJumpCharIf) Do(r TextReader, currentRead *Text) {
	if r.Current() != "" && strings.ContainsAny(r.Current(), n.Condition) {
		currentRead.Text += r.CurrentThenNext()
		n.Next.Do(r, currentRead)
		return
	}
	n.Else.Do(r, currentRead)
}

type NodeJumpCharNWay struct {
	Jumps []JumpChar
	Else  AutomataNode
}

func (n *NodeJumpCharNWay) Do(r TextReader, currentRead *Text) {
	if r.Current() != "" {
		for _, j := range n.Jumps {
			if strings.ContainsAny(r.Current(), j.Condition) {
				currentRead.Text += r.CurrentThenNext()
				j.Next.Do(r, currentRead)
				return
			}
		}
	}
	n.Else.Do(r, currentRead)
}

type NodeCharAny struct {
	Next AutomataNode
	Else AutomataNode
}

func (n *NodeCharAny) Do(r TextReader, currentRead *Text) {
	if r.Current() == "" {
		n.Else.Do(r, currentRead)
	} else {
		currentRead.Text += r.CurrentThenNext()
		n.Next.Do(r, currentRead)
	}
}

type NodeCharWhiteSpace struct {
	Next AutomataNode
	Else AutomataNode
}

func (n *NodeCharWhiteSpace) Do(r TextReader, currentRead *Text) {
	(&NodeJumpCharIf{
		JumpChar: JumpChar{
			Condition: " \n\t\r",
			Next:      n.Next,
		},
		Else: n.Else,
	}).Do(r, currentRead)
}

type NodeCharNonWhiteSpace struct {
	Next AutomataNode
	Else AutomataNode
}

func (n *NodeCharNonWhiteSpace) Do(r TextReader, currentRead *Text) {
	if r.Current() != "" && (!strings.ContainsAny(r.Current(), " \n\t\r")) {
		currentRead.Text += r.CurrentThenNext()
		n.Next.Do(r, currentRead)
		return
	}
	n.Else.Do(r, currentRead)
}

type NodeCharNumber struct {
	Next AutomataNode
	Else AutomataNode
}

func (n *NodeCharNumber) Do(r TextReader, currentRead *Text) {
	(&NodeJumpCharIf{
		JumpChar: JumpChar{
			Condition: "0123456789",
			Next:      n.Next,
		},
		Else: n.Else,
	}).Do(r, currentRead)
}

type NodePosOrNegSign struct {
	NextPosSign AutomataNode
	NextNegSign AutomataNode
	Else        AutomataNode
}

func (n *NodePosOrNegSign) Do(r TextReader, currentRead *Text) {
	(&NodeJumpCharNWay{
		Jumps: []JumpChar{
			{"+", n.NextPosSign},
			{"-", n.NextNegSign},
		},
		Else: n.Else,
	}).Do(r, currentRead)
}

type NodeDot struct {
	Next AutomataNode
	Else AutomataNode
}

func (n *NodeDot) Do(r TextReader, currentRead *Text) {
	(&NodeJumpCharIf{
		JumpChar: JumpChar{
			Condition: ".",
			Next:      n.Next,
		},
		Else: n.Else,
	}).Do(r, currentRead)
}

type DoubleQuote struct {
	Next AutomataNode
	Else AutomataNode
}

func (n *DoubleQuote) Do(r TextReader, currentRead *Text) {
	(&NodeJumpCharIf{
		JumpChar: JumpChar{
			Condition: "\"",
			Next:      n.Next,
		},
		Else: n.Else,
	}).Do(r, currentRead)
}

type DoubleQuoteOrSlash struct {
	NextQuote AutomataNode
	NextSlash AutomataNode
	Else      AutomataNode
}

func (n *DoubleQuoteOrSlash) Do(r TextReader, currentRead *Text) {
	(&NodeJumpCharNWay{
		Jumps: []JumpChar{
			{"\"", n.NextQuote},
			{"\\", n.NextSlash},
		},
		Else: n.Else,
	}).Do(r, currentRead)
}

type SingleQuote struct {
	Next AutomataNode
	Else AutomataNode
}

func (n *SingleQuote) Do(r TextReader, currentRead *Text) {
	(&NodeJumpCharIf{
		JumpChar: JumpChar{
			Condition: "'",
			Next:      n.Next,
		},
		Else: n.Else,
	}).Do(r, currentRead)
}

type SingleQuoteOrSlash struct {
	NextQuote AutomataNode
	NextSlash AutomataNode
	Else      AutomataNode
}

func (n *SingleQuoteOrSlash) Do(r TextReader, currentRead *Text) {
	(&NodeJumpCharNWay{
		Jumps: []JumpChar{
			{"'", n.NextQuote},
			{"\\", n.NextSlash},
		},
		Else: n.Else,
	}).Do(r, currentRead)
}
