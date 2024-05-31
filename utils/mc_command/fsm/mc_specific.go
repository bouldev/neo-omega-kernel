package fsm

import "strings"

type selectorOrNameLeadingCharNode struct {
	Selector AutomataNode
	Name     AutomataNode
	Else     AutomataNode
}

func (n *selectorOrNameLeadingCharNode) Do(r TextReader, currentRead *Text) {
	if r.Current() == "" || strings.Contains(r.Current(), " \r\n\t") {
		n.Else.Do(r, currentRead)
		return
	}
	if r.Current() == "@" {
		currentRead.Text += r.CurrentThenNext()
		n.Selector.Do(r, currentRead)
		return
	}
	n.Name.Do(r, currentRead)
}

// Start -> [1.@orNot(WhiteSpaceOrEof)]
// [1] -if@->[2.AnySpecifics:p,r,a,e,s,c,v,initiator]
// [1] -ifNot(WhiteSpaceOrEof)->[3.StringWithQuote]
// [2] -> ([4.WhiteSpaceOrElse->5.MiddleBracketString] or Else)
// [3] -> Next
// [4->5] -> Next
// `@abc` `@abc[ ... ]` `" sth "` `name`
func MCSelectorOrNameLogic(end, fail AutomataNode) AutomataNode {
	_1 := &selectorOrNameLeadingCharNode{}
	_3 := MakeQuoteMixtureNonExceptStringNode("\r\n\t ")
	_4 := MakeWhiteSpaceStringNode()
	_5 := MakeMiddleBracketStringNode()
	_4_then_5 := (&MachineNodeWrapper{
		Logic: func(end, fail AutomataNode) AutomataNode {
			_4.Next = _5
			_4.Else = _5
			_5.Next = end
			_5.Else = fail
			return _4
		},
	})
	_2 := &MatchSpecificsNode{
		Targets: []struct {
			Target string
			Fold   bool
			Next   AutomataNode
		}{
			{"p", true, _4_then_5},
			{"r", true, _4_then_5},
			{"a", true, _4_then_5},
			{"e", true, _4_then_5},
			{"s", true, _4_then_5},
			{"c", true, _4_then_5},
			{"v", true, _4_then_5},
			{"initiator", true, _4_then_5},
		},
		Else: fail,
	}
	_1.Selector = _2
	_1.Name = _3
	_1.Else = fail
	_3.Next = end
	_3.Else = fail
	_4_then_5.Next = end
	_4_then_5.Else = end
	return _1
}

func MakeMCSelectorOrNameLogic() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: MCSelectorOrNameLogic,
	}
}

// Start -> [1.WhiteSpaceOrElse]
// [1]->[2.^or~orElse]
// [2]-IfRelative->[3.SignedFloatNumberOrElse]
// [2]-Else->[4.SignedFloatNumber]
// [3]-> Next
// [4]-> Next
// `^` `~` `123` `^+/-123` `~123.456`
func PositionLogic(end, fail AutomataNode) AutomataNode {
	_1 := MakeWhiteSpaceStringNode()
	_2 := &NodeJumpCharIf{JumpChar: JumpChar{Condition: "^~"}}
	_3 := MakeSignedFloatStringNode()
	_4 := MakeSignedFloatStringNode()
	_1.Next = _2
	_1.Else = _2
	_2.JumpChar.Next = _3
	_2.Else = _4
	_3.Next = end
	_3.Else = end
	_4.Next = end
	_4.Else = fail
	return _1
}
func MakePositionNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: PositionLogic,
	}
}

func PositionGroupLogic(end, fail AutomataNode) AutomataNode {
	_1 := MakePositionNode()
	_2 := MakePositionNode()
	_3 := MakePositionNode()
	_1.Next = _2
	_1.Else = fail
	_2.Next = _3
	_2.Else = fail
	_3.Next = end
	_3.Else = fail
	return _1
}
func MakePositionGroupNode() *MachineNodeWrapper {
	return &MachineNodeWrapper{
		Logic: PositionGroupLogic,
	}
}
