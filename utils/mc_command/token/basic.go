package token

import "neo-omega-kernel/utils/mc_command/fsm"

type TextReader interface {
	// if "" means EOF
	Current() string
	CurrentThenNext() string
	Next()
	Snapshot() func()
}

func ReadWhiteSpace(r TextReader) (ok bool, token string) {
	return fsm.DoLogic(r, fsm.WhiteSpaceStringLogic)
}

func ReadSignedInteger(r TextReader) (ok bool, token string) {
	return fsm.DoLogic(r, fsm.SignedIntegerLogic)
}

func ReadNonWhiteSpace(r TextReader) (ok bool, token string) {
	return fsm.DoLogic(r, fsm.NonWhiteSpaceStringLogic)
}

func ReadMCSelector(r TextReader) (ok bool, token string) {
	return fsm.DoLogic(r, fsm.MCSelectorOrNameLogic)
}

func ReadPosition(r TextReader) (ok bool, token string) {
	return fsm.DoLogic(r, fsm.PositionGroupLogic)
}

func ReadUntilEnd(r TextReader) (ok bool, token string) {
	return fsm.DoLogic(r, fsm.ReadUntilEndLogic)
}
