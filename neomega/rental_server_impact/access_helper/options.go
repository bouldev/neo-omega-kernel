package access_helper

import (
	"time"
)

const DefaultAuthServer = "https://api.fastbuilder.pro"

type PrivilegeStuffOptions struct {
	DieOnLosingOpPrivilege     bool
	OpPrivilegeRemovedCallBack func()
}

type ImpactOption struct {
	AuthServer     string
	UserName       string
	UserPassword   string
	UserToken      string
	ServerCode     string
	ServerPassword string
}

type Options struct {
	ServerConnectionTimeout time.Duration
	ChallengeSolvingTimeout time.Duration
	WriteBackToken          bool
	ServerConnectRetryTimes int
	ImpactOption            *ImpactOption
	PrintUQHolderDebugInfo  bool
	MakeBotCreative         bool
	DisableCommandBlock     bool
	MaximumWaitTime         time.Duration
	*PrivilegeStuffOptions
	ReasonWithPrivilegeStuff bool
}

func DefaultImpactOption() *ImpactOption {
	return &ImpactOption{
		AuthServer:     DefaultAuthServer,
		UserName:       "",
		UserPassword:   "",
		UserToken:      "",
		ServerCode:     "",
		ServerPassword: "",
	}
}

func DefaultOptions() *Options {
	return &Options{
		ServerConnectionTimeout: time.Minute,
		ChallengeSolvingTimeout: time.Minute,
		WriteBackToken:          false,
		ServerConnectRetryTimes: 1,
		ImpactOption:            DefaultImpactOption(),
		PrintUQHolderDebugInfo:  false,
		MakeBotCreative:         true,
		DisableCommandBlock:     true,
		MaximumWaitTime:         time.Minute * 3,
		PrivilegeStuffOptions: &PrivilegeStuffOptions{
			DieOnLosingOpPrivilege:     true,
			OpPrivilegeRemovedCallBack: nil,
		},
		ReasonWithPrivilegeStuff: true,
	}
}
