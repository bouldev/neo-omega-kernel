package minimal_client_entry

import (
	"flag"
	"fmt"

	"github.com/OmineDev/neomega-core/neomega/rental_server_impact/access_helper"
)

type ArgsPlaceHolder struct {
	AuthServer     *string
	UserName       *string
	UserPassword   *string
	UserToken      *string
	ServerCode     *string
	ServerPassword *string
	WriteBackToken *bool
}

var argsPlaceHolder *ArgsPlaceHolder

func RegArgs() {
	if argsPlaceHolder != nil {
		return
	}
	argsPlaceHolder = &ArgsPlaceHolder{}
	argsPlaceHolder.AuthServer = flag.String("auth-server", "", "auth server")
	flag.StringVar(argsPlaceHolder.AuthServer, "A", "", "auth server")
	argsPlaceHolder.UserName = flag.String("user-name", "", "user name")
	argsPlaceHolder.UserPassword = flag.String("user-password", "", "user password")
	argsPlaceHolder.UserToken = flag.String("user-token", "", "user token")
	flag.StringVar(argsPlaceHolder.UserToken, "T", "", "user token")
	flag.StringVar(argsPlaceHolder.UserToken, "token", "", "user token")
	argsPlaceHolder.ServerCode = flag.String("server", "", "rental server code")
	argsPlaceHolder.ServerPassword = flag.String("server-password", "", "rental server password")
	argsPlaceHolder.WriteBackToken = flag.Bool("save-token", false, "save token to ~/config/fastbuilder/token after login")
}

func GetArgs() *access_helper.ImpactOption {
	RegArgs()
	if !flag.Parsed() {
		flag.Parse()
	}

	return &access_helper.ImpactOption{
		AuthServer:     *argsPlaceHolder.AuthServer,
		UserName:       *argsPlaceHolder.UserName,
		UserPassword:   *argsPlaceHolder.UserPassword,
		UserToken:      *argsPlaceHolder.UserToken,
		ServerCode:     *argsPlaceHolder.ServerCode,
		ServerPassword: *argsPlaceHolder.ServerPassword,
	}
}

func MakeArgs(args *access_helper.ImpactOption) []string {
	return []string{
		fmt.Sprintf("--%v=%v", "auth-server", args.AuthServer),
		fmt.Sprintf("--%v=%v", "user-token", args.UserToken),
		fmt.Sprintf("--%v=%v", "user-name", args.UserName),
		fmt.Sprintf("--%v=%v", "user-password", args.UserPassword),
		fmt.Sprintf("--%v=%v", "server", args.ServerCode),
		fmt.Sprintf("--%v=%v", "server-password", args.ServerPassword),
	}
}
