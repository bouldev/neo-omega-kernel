package access_point

import (
	"flag"
	"neo-omega-kernel/neomega/entries/minimal_client_entry"
	"neo-omega-kernel/neomega/entries/minimal_end_point_entry"
	"neo-omega-kernel/neomega/rental_server_impact/access_helper"
)

type ArgsPlaceHolder struct {
}

type Args struct {
	*access_helper.ImpactOption
	AccessArgs *minimal_end_point_entry.Args
}

var argsPlaceHolder *ArgsPlaceHolder

func RegArgs() {
	minimal_client_entry.RegArgs()
	minimal_end_point_entry.RegArgs()
	if argsPlaceHolder != nil {
		return
	}
	argsPlaceHolder = &ArgsPlaceHolder{}
}

func GetArgs() *Args {
	RegArgs()
	if !flag.Parsed() {
		flag.Parse()
	}

	return &Args{
		ImpactOption: minimal_client_entry.GetArgs(),
		AccessArgs:   minimal_end_point_entry.GetArgs(),
	}
}

func MakeArgs(args *Args) []string {
	return append(
		minimal_client_entry.MakeArgs(args.ImpactOption),
		minimal_end_point_entry.MakeArgs(args.AccessArgs)...,
	)
}
