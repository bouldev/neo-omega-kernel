package access_helper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/minecraft"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/bundle"
	"neo-omega-kernel/neomega/fbauth"
	"neo-omega-kernel/neomega/rental_server_impact/challenges"
	"neo-omega-kernel/nodes"
	"neo-omega-kernel/py_rpc"
	"time"
)

// Copied from phoenixbuilder/core/core
func initializeMinecraftConnection(ctx context.Context, authenticator minecraft.Authenticator) (conn *minecraft.Conn, err error) {
	dialer := minecraft.Dialer{
		Authenticator: authenticator,
	}
	conn, err = dialer.DialContext(ctx, "raknet")
	if err != nil {
		return
	}
	conn.WritePacket(&packet.ClientCacheStatus{
		Enabled: false,
	})
	runtimeid := fmt.Sprintf("%d", conn.GameData().EntityUniqueID)
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"SyncUsingMod",
			[]interface{}{},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"SyncVipSkinUuid",
			[]interface{}{nil},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"ClientLoadAddonsFinishedFromGac",
			[]interface{}{},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"ModEventC2S",
			[]interface{}{
				"Minecraft",
				"preset",
				"GetLoadedInstances",
				map[string]interface{}{
					"playerId": runtimeid,
				},
			},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"arenaGamePlayerFinishLoad",
			[]interface{}{},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"ModEventC2S",
			[]interface{}{
				"Minecraft",
				"vipEventSystem",
				"PlayerUiInit",
				runtimeid,
			},
			nil,
		}),
	})
	return
}

func makeAuthenticatorAndChallengeSolver(options *ImpactOption, writeBackFBToken bool) (authenticator minecraft.Authenticator, challengeSolver challenges.CanSolveChallenge, err error) {
	clientOptions := fbauth.MakeDefaultClientOptions()
	clientOptions.AuthServer = options.AuthServer
	fmt.Printf(i18n.T(i18n.S_connecting_to_auth_server)+": %v\n", options.AuthServer)
	fbClient, err := fbauth.CreateClient(clientOptions)
	if err != nil {
		return nil, nil, err
	}
	challengeSolver = fbClient
	fmt.Println(i18n.T(i18n.S_done_connecting_to_auth_server))
	hashedPassword := ""
	if options.UserToken == "" {
		psw_sum := sha256.Sum256([]byte(options.UserPassword))
		hashedPassword = hex.EncodeToString(psw_sum[:])
	}
	authenticator = fbauth.NewAccessWrapper(fbClient, options.ServerCode, options.ServerPassword, options.UserToken, options.UserName, hashedPassword, writeBackFBToken)
	return authenticator, challengeSolver, nil
}

func connectMCServer(ctx context.Context, authenticator minecraft.Authenticator) (conn *minecraft.Conn, err error) {
	conn, err = initializeMinecraftConnection(ctx, authenticator)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ErrRentalServerConnectionTimeOut
		}
		return nil, fmt.Errorf("%v: %v", ErrFailToConnectRentalServer, err)
	}
	return conn, nil
}

func connectMCServerWithRetry(ctx context.Context, authenticator minecraft.Authenticator, retryTimesRemains int) (conn *minecraft.Conn, err error) {
	retryTimes := 0
	for {
		conn, err = connectMCServer(ctx, authenticator)
		if err == nil {
			break
		} else {
			fmt.Println(err)
		}
		if retryTimesRemains <= 0 {
			break
		}
		retryTimes++
		fmt.Printf(i18n.T(i18n.S_fail_connecting_to_mc_server_retrying), retryTimes)
		// wait for 1s
		time.Sleep(time.Second)
		retryTimesRemains--
	}
	if err != nil {
		return nil, err
	}
	fmt.Println(i18n.T(i18n.S_done_connecting_to_mc_server))
	return conn, nil
}

func makeNodeOmegaCoreFromConn(node nodes.Node, conn *minecraft.Conn) (neomega.UnReadyMicroOmega, error) {
	omegaCore := bundle.NewAccessPointMicroOmega(node, conn)
	return omegaCore, nil

}

func copeWithRentalServerChallenge(ctx context.Context, omegaCore neomega.MicroOmega, canSolveChallenge challenges.CanSolveChallenge) (err error) {
	fmt.Println(i18n.T(i18n.S_coping_with_rental_server_challenges))
	challengeSolver := challenges.NewPyRPCResponder(omegaCore, canSolveChallenge)

	err = challengeSolver.ChallengeCompete(ctx)
	if err != nil {
		return ErrFBChallengeSolvingTimeout
	}
	fmt.Println(i18n.T(i18n.S_done_coping_with_rental_server_challenges))
	return nil
}

func reasonWithPrivilegeStuff(ctx context.Context, deadReason chan error, omegaCore neomega.MicroOmega, options *PrivilegeStuffOptions) (err error) {
	fmt.Println(i18n.T(i18n.S_checking_bot_op_permission_and_game_cheat_mode))
	helper := challenges.NewOperatorChallenge(omegaCore, func() {
		if options.OpPrivilegeRemovedCallBack != nil {
			options.OpPrivilegeRemovedCallBack()
		}
		if options.DieOnLosingOpPrivilege {
			deadReason <- ErrBotOpPrivilegeRemoved
		}
	})
	waitErr := make(chan error)
	go func() {
		waitErr <- helper.WaitForPrivilege(ctx)
	}()
	select {
	case err = <-waitErr:
	case err = <-deadReason:
	}
	if err != nil {
		return err
	}
	fmt.Println(i18n.T(i18n.S_done_checking_bot_op_permission_and_game_cheat_mode))
	return nil
}

func makeBotCreative(omegaCoreCtrl neomega.GameCtrl) {
	waitor := make(chan struct{})
	fmt.Println(i18n.T(i18n.S_switching_bot_to_creative_mode))
	omegaCoreCtrl.SendWebSocketCmdNeedResponse("gamemode c @s").SetTimeout(time.Second * 3).AsyncGetResult(func(output *packet.CommandOutput) {
		fmt.Println(i18n.T(i18n.S_done_setting_bot_to_creative_mode))
		close(waitor)
	})
	<-waitor
}

func disableCommandBlock(omegaCoreCtrl neomega.GameCtrl) {
	omegaCoreCtrl.SendWOCmd("gamerule commandblocksenabled false")
	//	waitor := make(chan struct{})
	//	omegaCoreCtrl.SendPlayerCmdNeedResponse("gamerule commandblocksenabled false").AsyncGetResult(func(output *packet.CommandOutput) {
	fmt.Println(i18n.T(i18n.S_done_setting_commandblocksenabled_false))
	//		close(waitor)
	//	})
	//
	// <-waitor
}

func waitDead(omegaCore neomega.MicroOmega, deadReason chan error) {
	// SetTime packet will be sent by server every 256 ticks, even dodaylightcycle gamerule disabled
	period := time.Second * 15
	startTime := time.Now()
	lastReceivePacket := time.Now()
	omegaCore.GetGameListener().SetAnyPacketCallBack(func(p packet.Packet) {
		lastReceivePacket = time.Now()
	}, false)
	for {
		time.Sleep(time.Second)
		nowTime := time.Now()
		if lastReceivePacket.Add(time.Second * 5).Before(nowTime) {
			omegaCore.GetGameControl().SendWebSocketCmdOmitResponse("testforblock ~~~ air")
		}
		if lastReceivePacket.Add(time.Second * 15).Before(nowTime) {
			deadReason <- fmt.Errorf(i18n.T(i18n.S_no_response_after_a_long_time_bot_is_down), period*2, time.Since(startTime))
			break
		}
	}
}
