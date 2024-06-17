package login_and_spawn_core

import (
	"neo-omega-kernel/minecraft/resource"
	"neo-omega-kernel/minecraft_neo/game_data"
)

// Authenticated returns true if the connection was authenticated through XBOX Live services.
func (core *Core) Authenticated() bool {
	return core.IdentityData.XUID != ""
}

// GameData returns specific game data set to the connection for the player to be initialised with. If the
// Conn is obtained using Listen, this game data may be set to the Listener. If obtained using Dial, the data
// is obtained from the server.
func (core *Core) GameData() game_data.GameData {
	return core.gameData
}

// ResourcePacks returns a slice of all resource packs the connection holds. For a Conn obtained using a
// Listener, this holds all resource packs set to the Listener. For a Conn obtained using Dial, the resource
// packs include all packs sent by the server connected to.
func (core *Core) ResourcePacks() []*resource.Pack {
	return core.resourcePacks
}

// Close closes the Conn and its underlying connection. Before closing, it also calls Flush() so that any
// packets currently pending are sent out.
func (core *Core) Close() error {
	var err error
	core.once.Do(func() {
		close(core.close)
	})
	return err
}

// ClientCacheEnabled checks if the connection has the client blob cache enabled. If true, the server may send
// blobs to the client to reduce network transmission, but if false, the client does not support it, and the
// server must send chunks as usual.
func (core *Core) ClientCacheEnabled() bool {
	return core.EnableClientCache
}

// ChunkRadius returns the initial chunk radius of the connection. For connections obtained through a
// Listener, this is the radius that the client requested. For connections obtained through a Dialer, this
// is the radius that the server approved upon.
func (core *Core) ChunkRadius() int {
	return int(core.gameData.ChunkRadius)
}
