package access_helper

import (
	"context"
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/bundle"
	"neo-omega-kernel/neomega/rental_server_impact/challenges"
	"neo-omega-kernel/nodes/defines"
	"time"
)

type AuthenticatorWithToken interface {
	Authenticator
	GetToken() string
}

func ImpactServer(ctx context.Context, node defines.Node, options *Options) (omegaCore neomega.MicroOmega, err error) {
	defer func() {
		if err != nil {
			err = i18n.FuzzyTransErr(err)
		}
	}()
	fmt.Println(i18n.T(i18n.S_executing_login_sequence))
	if options.MaximumWaitTime > 0 {
		ctx, _ = context.WithTimeout(ctx, options.MaximumWaitTime)
	}
	fmt.Println(i18n.T(i18n.S_updating_tag_in_omega_net))
	if node.CheckNetTag("access-point") {
		return nil, fmt.Errorf(i18n.T(i18n.S_net_position_conflict_only_one_access_point_can_exist))
	}
	node.SetTags("access-point")
	// make auth client and wrap authenticator & challenge solver
	var authenticator Authenticator
	var challengeSolver challenges.CanSolveChallenge
	authenticator, challengeSolver, err = makeAuthenticatorAndChallengeSolver(options.ImpactOption, options.WriteBackToken)
	if err != nil {
		return nil, err
	}

	// connect to minecraft solver
	var unReadyOmega neomega.UnReadyMicroOmega
	{
		mcServerConnectCtx := ctx
		if options.ServerConnectionTimeout != 0 {
			mcServerConnectCtx, _ = context.WithTimeout(ctx, options.ServerConnectionTimeout)
		}
		password := "No"
		if options.ImpactOption.ServerPassword != "" {
			password = "Yes"
		}
		fmt.Printf(i18n.T(i18n.S_connecting_to_mc_server)+"\n", options.ImpactOption.ServerCode, password)
		conn, err := loginMCServerWithRetry(mcServerConnectCtx, authenticator, options.ServerConnectRetryTimes)
		if err != nil {
			return nil, err
		}
		unReadyOmega = bundle.NewAccessPointMicroOmega(node, conn)
	}
	// unReadyOmega, err = makeNodeOmegaCoreFromConn(node, conn)
	omegaCore = unReadyOmega

	// POST PROCESSES
	{
		challengeSolvingCtx := ctx
		if options.ChallengeSolvingTimeout != 0 {
			challengeSolvingCtx, _ = context.WithTimeout(ctx, options.ChallengeSolvingTimeout)
		}
		if err := copeWithRentalServerChallenge(challengeSolvingCtx, omegaCore, challengeSolver); err != nil {
			return nil, err
		}
	}
	if options.ReasonWithPrivilegeStuff {
		err := reasonWithPrivilegeStuff(ctx, omegaCore.Dead(), omegaCore, options.PrivilegeStuffOptions)
		if err != nil {
			return nil, err
		}
	}
	if options.MakeBotCreative {
		unReadyOmega.PostponeActionsAfterChallengePassed("make bot creative", func() {
			makeBotCreative(omegaCore.GetGameControl())
		})
	}
	if options.DisableCommandBlock {
		unReadyOmega.PostponeActionsAfterChallengePassed("make bot creative", func() {
			disableCommandBlock(omegaCore.GetGameControl())
		})
	}
	unReadyOmega.NotifyChallengePassed()
	go waitDead(omegaCore, omegaCore.Dead())
	// wait everything stable
	// we need to wait until some packets received before using some api
	time.Sleep(time.Second * 3)
	return omegaCore, nil
}
