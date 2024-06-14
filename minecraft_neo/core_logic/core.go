package core_logic

import (
	"fmt"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/minecraft/resource"
	"neo-omega-kernel/minecraft_neo/cascade_conn/can_close"
	"neo-omega-kernel/minecraft_neo/cascade_conn/defines"
	"sync/atomic"

	"github.com/google/uuid"
)

type Core struct {
	*Options
	can_close.CanCloseWithError
	underlayConn defines.PacketConn
	shieldID     atomic.Int32
	expectedIDs  atomic.Value

	compression packet.Compression

	gameData         GameData
	gameDataReceived atomic.Bool
	// readyToLogin is a bool indicating if the connection is ready to login. This is used to ensure that the client
	// has received the relevant network settings before the login sequence starts.
	readyToLogin bool
	//  signal that the connection is ready to login
	readyToLoginSignal chan struct{}
	// loggedIn is a bool indicating if the connection was logged in. It is set to true after the entire login
	// sequence is completed.
	loggedIn bool
	// signal that the connection was considered logged in
	loggedInSignal chan struct{}
	// spawn is a bool channel indicating if the connection is currently waiting for its spawning in
	// the world: It is completing a sequence that will result in the spawning.
	spawn           chan struct{}
	waitingForSpawn atomic.Bool
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
	downloadResourcePack func(id uuid.UUID, version string, currentPack, totalPacks int) bool
	// ignoredResourcePacks is a slice of resource packs that are not being downloaded due to the downloadResourcePack
	// func returning false for the specific pack.
	ignoredResourcePacks []exemptedResourcePack

	cacheEnabled bool
}

func NewCore(conn defines.PacketConn, opt *Options) *Core {
	c := &Core{
		Options:           opt,
		CanCloseWithError: can_close.NewClose(conn.Close),
		underlayConn:      conn,

		spawn:              make(chan struct{}, 1),
		readyToLoginSignal: make(chan struct{}, 1),
		loggedInSignal:     make(chan struct{}, 1),
	}
	go func() {
		err := <-conn.WaitClosed()
		c.CloseWithError(err)
	}()
	return c
}

func (c *Core) ShieldID() int32 {
	return c.shieldID.Load()
}

func (c *Core) HandlePacket(pk *defines.RawPacketAndByte) {
	fmt.Println(pk.Pk.ID())
	if c.readyToLogin {
		// This is the signal that the connection is ready to login, so we put a value in the channel so that
		// it may be detected.
		c.readyToLoginSignal <- struct{}{}
	}
	if c.loggedIn {
		// This is the signal that the connection was considered logged in, so we put a value in the channel so
		// that it may be detected.
		c.loggedInSignal <- struct{}{}
	}
	for _, id := range c.expectedIDs.Load().([]uint32) {
		if id == pk.Pk.ID() {
			c.react(pk.Pk)
		}
	}
}

func (c *Core) StartReactRoutine() {
	c.underlayConn.ReadRoutine(c)
}

func (c *Core) expect(packetIDs ...uint32) {
	c.expectedIDs.Store(packetIDs)
}

func (c *Core) StartLoginSequence() error {
	c.expect(packet.IDNetworkSettings, packet.IDPlayStatus)
	if err := c.WritePacket(&packet.RequestNetworkSettings{ClientProtocol: protocol.CurrentProtocol}); err != nil {
		return err
	}
	_ = c.underlayConn.Flush()
	<-c.readyToLoginSignal

	c.expect(packet.IDServerToClientHandshake, packet.IDPlayStatus)
	if err := c.WritePacket(&packet.Login{ConnectionRequest: c.ConnectionRequest, ClientProtocol: protocol.CurrentProtocol}); err != nil {
		return err
	}
	_ = c.underlayConn.Flush()
	<-c.loggedInSignal
	return nil
}

func (c *Core) WritePacket(pk packet.Packet) error {
	return c.underlayConn.WritePacket(pk, c.ShieldID())
}
