package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/minecraft_neo/cascade_conn/base_net"
	"neo-omega-kernel/minecraft_neo/cascade_conn/byte_frame_conn"
	"neo-omega-kernel/minecraft_neo/cascade_conn/packet_conn"
	"neo-omega-kernel/minecraft_neo/login_and_spawn_core"
	"neo-omega-kernel/minecraft_neo/login_and_spawn_core/options"
	"neo-omega-kernel/neomega/fbauth"
)

func main() {
	AuthServer := "https://liliya233.uk/"
	ServerCode := ""
	ServerPassword := ""
	Token := ""
	UserName := "2401PT"
	UserPassword := ""
	{
		psw_sum := sha256.Sum256([]byte(UserPassword))
		UserPassword = hex.EncodeToString(psw_sum[:])
	}

	authClient, err := fbauth.CreateClient(&fbauth.ClientOptions{
		AuthServer:          AuthServer,
		RespondUserOverride: "",
	})
	if err != nil {
		panic(err)
	}
	authenticator := fbauth.NewAccessWrapper(
		authClient, ServerCode, ServerPassword, Token, UserName, UserPassword,
		false,
	)

	privateKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	publicKey, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)

	ctx := context.Background()

	address, chainData, err := authenticator.GetAccess(ctx, publicKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("Auth Pass")
	rakNetConn, err := base_net.RakNet.DialContext(ctx, address)
	if err != nil {
		panic(err)
	}

	opt := options.NewDefaultOptions(address, chainData, privateKey)
	byteFrameConn := byte_frame_conn.NewConnectionFromNet(rakNetConn)
	packetConn := packet_conn.NewPacketConn(byteFrameConn, false)
	loginAndSpawnCore := login_and_spawn_core.NewLoginAndSpawnCore(packetConn, opt)
	go packetConn.ListenRoutine(func(pk packet.Packet) {
		fmt.Println("read:", pk.ID())
		loginAndSpawnCore.Receive(pk)
	})
	err = loginAndSpawnCore.Login(ctx)
	if err != nil {
		panic(err)
	}
	panic(<-packetConn.WaitClosed())
	// for {
	// 	pk, err := conn.ReadPacket()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// }
	// frameConn := byte_frame_conn.NewConnectionFromNet(rakNetConn)
	// packetConn := packet_conn.NewClientFromConn(frameConn)
	// option := core_logic.NewDefaultOptions(address, chainData, privateKey)
	// core := core_logic.NewCore(packetConn, option)
	// go core.StartReactRoutine()
	// core.StartLoginSequence()
	// err = <-core.WaitClosed()
	// fmt.Println(err)
}
