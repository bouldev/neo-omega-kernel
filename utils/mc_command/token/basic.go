package token

import "github.com/OmineDev/neomega-core/utils/mc_command/fsm"

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

func ReadSpecific(r TextReader, target string, fold bool) (ok bool, token string) {
	return fsm.DoLogic(r, fsm.MakeMatchSpecificLogic(target, fold))
}

func ReadAnySpecifics(r TextReader, targets []string, fold bool) (ok bool, token string) {
	return fsm.DoLogic(r, fsm.MakeMatchSpecificsLogic(targets, fold))
}

func ReadAnyExcept(r TextReader, except string) (ok bool, token string) {
	return fsm.DoLogic(r, fsm.MakeNonExceptStringLogic(except))
}
