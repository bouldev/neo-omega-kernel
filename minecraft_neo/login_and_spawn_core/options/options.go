package options

import (
	"crypto/ecdsa"
	"log"
	"neo-omega-kernel/minecraft/protocol/login"
)

// Dialer allows specifying specific settings for connection to a Minecraft server.
// The zero value of Dialer is used for the package level Dial function.
type Options struct {
	// ChainData  string
	// Address    string
	PrivateKey *ecdsa.PrivateKey
	Request    []byte
	Salt       []byte
	// ErrorLog is a log.Logger that errors that occur during packet handling of servers are written to. By
	// default, ErrorLog is set to one equal to the global logger.
	ErrorLog *log.Logger

	// ClientData is the client data used to login to the server with. It includes fields such as the skin,
	// locale and UUIDs unique to the client. If empty, a default is sent produced using defaultClientData().
	ClientData login.ClientData
	// IdentityData is the identity data used to login to the server with. It includes the username, UUID and
	// XUID of the player.
	// The IdentityData object is obtained using Minecraft auth if Email and Password are set. If not, the
	// object provided here is used, or a default one if left empty.
	IdentityData login.IdentityData

	// EnableClientCache, if set to true, enables the client blob cache for the client. This means that the
	// server will send chunks as blobs, which may be saved by the client so that chunks don't have to be
	// transmitted every time, resulting in less network transmission.
	EnableClientCache bool
}
