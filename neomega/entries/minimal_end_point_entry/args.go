package minimal_end_point_entry

import (
	"flag"
	"fmt"
)

const DefaultAccessPointAddr = "tcp://localhost:24015" // "ipc://neomega_ctrl.ipc"    //"tcp://localhost:24015"

type ArgsPlaceHolder struct {
	AccessPointAddr *string
}

type Args struct {
	AccessPointAddr string
}

var argsPlaceHolder *ArgsPlaceHolder

func RegArgs() {
	if argsPlaceHolder != nil {
		return
	}
	argsPlaceHolder = &ArgsPlaceHolder{}
	argsPlaceHolder.AccessPointAddr = flag.String("access-point-addr", DefaultAccessPointAddr, "access point connection addr")
}

func GetArgs() *Args {
	RegArgs()
	if !flag.Parsed() {
		flag.Parse()
	}

	return &Args{
		AccessPointAddr: *argsPlaceHolder.AccessPointAddr,
	}
}

func MakeArgs(args *Args) []string {
	return []string{
		fmt.Sprintf("--%v=%v", "access-point-addr", args.AccessPointAddr),
	}
}
