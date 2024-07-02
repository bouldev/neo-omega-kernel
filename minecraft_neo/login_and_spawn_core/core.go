package login_and_spawn_core

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/minecraft/resource"
	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	"github.com/OmineDev/neomega-core/minecraft_neo/cascade_conn/defines"
	"github.com/OmineDev/neomega-core/minecraft_neo/game_data"
	"github.com/OmineDev/neomega-core/minecraft_neo/login_and_spawn_core/options"
)

type Core struct {
	*options.Options
	can_close.CanCloseWithError
	packetConn defines.PacketConn
	// pool packet.Pool
	// enc         *packet.Encoder
	// dec         *packet.Decoder
	compression packet.Compression

	gameData         game_data.GameData
	gameDataReceived atomic.Bool

	// readyToLogin is a bool indicating if the connection is ready to login. This is used to ensure that the client
	// has received the relevant network settings before the login sequence starts.
	readyToLogin       bool
	readyToLoginSignal chan struct{} //l
	// loggedIn is a bool indicating if the connection was logged in. It is set to true after the entire login
	// sequence is completed.
	loggedIn       bool
	loggedInSignal chan struct{} //c
	// spawn is a bool channel indicating if the connection is currently waiting for its spawning in
	// the world: It is completing a sequence that will result in the spawning.
	spawn           chan struct{}
	waitingForSpawn atomic.Bool

	// expectedIDs is a slice of packet identifiers that are next expected to arrive, until the connection is
	// logged in.
	expectedIDs atomic.Value

	packMu sync.Mutex
	// resourcePacks is a slice of resource packs that the listener may hold. Each client will be asked to
	// download these resource packs upon joining.
	resourcePacks []*resource.Pack
	// biomes is a map of biome definitions that the listener may hold. Each client will be sent these biome
	// definitions upon joining.
	biomes map[string]any
	// texturePacksRequired specifies if clients that join must accept the texture pack in order for them to
	// be able to join the server. If they don't accept, they can only leave the server.
	texturePacksRequired bool
	packQueue            *resourcePackQueue
	// downloadResourcePack is an optional function passed to a Dial() call. If set, each resource pack received
	// from the server will call this function to see if it should be downloaded or not.

	// ignoredResourcePacks is a slice of resource packs that are not being downloaded due to the downloadResourcePack
	// func returning false for the specific pack.
	ignoredResourcePacks []exemptedResourcePack
}

func NewLoginAndSpawnCore(packetConn defines.PacketConn, opt *options.Options) *Core {
	core := &Core{
		Options: opt,
		// close underlay conn on err
		CanCloseWithError:  can_close.NewClose(packetConn.Close),
		spawn:              make(chan struct{}),
		readyToLoginSignal: make(chan struct{}),
		loggedInSignal:     make(chan struct{}),
		packetConn:         packetConn,
	}
	go func() {
		// close when underlay err
		core.CloseWithError(<-packetConn.WaitClosed())
	}()
	return core
}

func (core *Core) Login(ctx context.Context) (err error) {
	// core.expectedIDs.Store([]uint32{packet.IDRequestNetworkSettings})
	core.expect(packet.IDNetworkSettings, packet.IDPlayStatus)
	if err := core.packetConn.WritePacket(&packet.RequestNetworkSettings{ClientProtocol: protocol.CurrentProtocol}); err != nil {
		return err
	}
	_ = core.packetConn.Flush()

	select {
	case <-core.WaitClosed():
		return core.CloseError()
	case <-ctx.Done():
		return ctx.Err()
	case <-core.readyToLoginSignal:
		// We've received our network settings, so we can now send our login request.
		core.expect(packet.IDServerToClientHandshake, packet.IDPlayStatus)
		if err := core.packetConn.WritePacket(&packet.Login{ConnectionRequest: core.Request, ClientProtocol: protocol.CurrentProtocol}); err != nil {
			return err
		}
		_ = core.packetConn.Flush()

		select {
		case err := <-core.WaitClosed():
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-core.loggedInSignal:
			// We've connected successfully. We return the connection and no error.
			return nil
		}
	}
}
