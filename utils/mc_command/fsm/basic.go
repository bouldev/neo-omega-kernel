package fsm

type TextReader interface {
	// if "" means EOF
	Current() string
	CurrentThenNext() string
	Next()
	Snapshot() func()
}

type Text struct {
	Text string
}

func (t *Text) Clone() *Text {
	return &Text{
		Text: t.Text,
	}
}

type AutomataNode interface {
	Do(r TextReader, currentRead *Text)
}

type NodeBasic struct {
	Accessed bool
}

func (n *NodeBasic) Do(r TextReader, currentRead *Text) {
	n.Accessed = true
}

type NodeEnd struct {
	Accessed bool
	Read     string
}

func (e *NodeEnd) Do(r TextReader, currentRead *Text) {
	e.Accessed = true
	e.Read = currentRead.Text
}

type NodeStart struct {
	Next AutomataNode
}

func (n *NodeStart) Do(r TextReader, currentRead *Text) {
	n.Next.Do(r, currentRead)
}

// type NodeBack struct {
// 	r           TextReader
// 	currentRead *Text
// 	Next        AutomataNode
// }

// func (n *NodeBack) Do(r TextReader, currentRead *Text) {
// 	n.Next.Do(n.r, n.currentRead)
// }

type NodeLazy struct {
	Make func() AutomataNode
}

func (n *NodeLazy) Do(r TextReader, currentRead *Text) {
	n.Make().Do(r, currentRead)
}
