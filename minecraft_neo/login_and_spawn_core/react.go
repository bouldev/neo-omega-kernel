package login_and_spawn_core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/OmineDev/neomega-core/minecraft/nbt"
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/login"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/minecraft/resource"
	"github.com/OmineDev/neomega-core/minecraft_neo/game_data"

	"github.com/google/uuid"
	"gopkg.in/square/go-jose.v2/jwt"
)

func (core *Core) Receive(pk packet.Packet) {
	loggedInBefore, readyToLoginBefore := core.loggedIn, core.readyToLogin
	core.receive(pk)
	if !readyToLoginBefore && core.readyToLogin {
		// This is the signal that the connection is ready to login, so we put a value in the channel so that
		// it may be detected.
		select {
		case core.readyToLoginSignal <- struct{}{}:
		default:
			fmt.Println("resend ready signal")
		}
	}
	if !loggedInBefore && core.loggedIn {
		// This is the signal that the connection was considered logged in, so we put a value in the channel so
		// that it may be detected.
		select {
		case core.loggedInSignal <- struct{}{}:
		default:
			fmt.Println("resend logged signal")
		}
	}
}

func (core *Core) receive(pk packet.Packet) {
	if pk.ID() == packet.IDDisconnect {
		core.CloseWithError(fmt.Errorf(pk.(*packet.Disconnect).Message))
		return
	}
	if core.loggedIn && !core.waitingForSpawn.Load() {
		return
	}
	for _, id := range core.expectedIDs.Load().([]uint32) {
		if id == pk.ID() {
			// If the packet was expected, so we handle it right now.
			err := core.handlePacket(pk)
			if err != nil {
				panic(err)
			}

		}
	}
}

// handlePacket handles an incoming packet. It returns an error if any of the data found in the packet was not
// valid or if handling failed for any other reason.
func (core *Core) handlePacket(pk packet.Packet) error {
	// fmt.Println(pk.ID())
	defer func() {
		_ = core.packetConn.Flush()
	}()
	switch pk := pk.(type) {
	// Internal packets destined for the server.
	case *packet.RequestNetworkSettings:
		return core.handleRequestNetworkSettings(pk)
	// case *packet.Login:
	// 	panic("Login!")
	// 	return core.handleLogin(pk)
	case *packet.ClientToServerHandshake:
		return core.handleClientToServerHandshake()
	case *packet.ClientCacheStatus:
		return core.handleClientCacheStatus(pk)
	case *packet.ResourcePackClientResponse:
		return core.handleResourcePackClientResponse(pk)
	case *packet.ResourcePackChunkRequest:
		return core.handleResourcePackChunkRequest(pk)
	case *packet.RequestChunkRadius:
		return core.handleRequestChunkRadius(pk)
	case *packet.SetLocalPlayerAsInitialised:
		return core.handleSetLocalPlayerAsInitialised(pk)

	// Internal packets destined for the client.
	case *packet.NetworkSettings:
		return core.handleNetworkSettings(pk)
	case *packet.ServerToClientHandshake:
		return core.handleServerToClientHandshake(pk)
	case *packet.PlayStatus:
		return core.handlePlayStatus(pk)
	case *packet.ResourcePacksInfo:
		return core.handleResourcePacksInfo(pk)
	case *packet.ResourcePackDataInfo:
		return core.handleResourcePackDataInfo(pk)
	case *packet.ResourcePackChunkData:
		return core.handleResourcePackChunkData(pk)
	case *packet.ResourcePackStack:
		return core.handleResourcePackStack(pk)
	case *packet.StartGame:
		return core.handleStartGame(pk)
	case *packet.ChunkRadiusUpdated:
		return core.handleChunkRadiusUpdated(pk)
	}
	return nil
}

// handleRequestNetworkSettings handles an incoming RequestNetworkSettings packet. It returns an error if the protocol
// version is not supported, otherwise sending back a NetworkSettings packet.
func (core *Core) handleRequestNetworkSettings(pk *packet.RequestNetworkSettings) error {
	found := false
	if !found {
		status := packet.PlayStatusLoginFailedClient
		if pk.ClientProtocol > protocol.CurrentProtocol {
			// The server is outdated in this case, so we have to change the status we send.
			status = packet.PlayStatusLoginFailedServer
		}
		_ = core.packetConn.WritePacket(&packet.PlayStatus{Status: status})
		return fmt.Errorf("%v connected with an incompatible protocol: expected protocol = %v, client protocol = %v", core.IdentityData.DisplayName, protocol.CurrentProtocol, pk.ClientProtocol)
	}

	core.expect(packet.IDLogin)
	if err := core.packetConn.WritePacket(&packet.NetworkSettings{
		CompressionThreshold: 512,
		CompressionAlgorithm: core.compression.EncodeCompression(),
	}); err != nil {
		return fmt.Errorf("error sending network settings: %v", err)
	}
	_ = core.packetConn.Flush()
	// core.EnableCompression(core.compression)
	core.packetConn.EnableCompression(core.compression)
	return nil
}

// handleNetworkSettings handles an incoming NetworkSettings packet, enabling compression for future packets.
func (core *Core) handleNetworkSettings(pk *packet.NetworkSettings) error {
	alg, ok := packet.CompressionByID(pk.CompressionAlgorithm)
	if !ok {
		return fmt.Errorf("unknown compression algorithm: %v", pk.CompressionAlgorithm)
	}
	// core.enc.EnableCompression(alg)
	core.packetConn.EnableCompression(alg)
	core.readyToLogin = true
	return nil
}

// // handleLogin handles an incoming login packet. It verifies and decodes the login request found in the packet
// // and returns an error if it couldn't be done successfully.
// func (core *Core) handleLogin(pk *packet.Login) error {
// 	// The next expected packet is a response from the client to the handshake.
// 	core.expect(packet.IDClientToServerHandshake)
// 	var (
// 		err        error
// 		authResult login.AuthResult
// 	)
// 	core.identityData, core.clientData, authResult, err = login.Parse(pk.ConnectionRequest)
// 	if err != nil {
// 		return fmt.Errorf("parse login request: %w", err)
// 	}

// 	// Make sure the player is logged in with XBOX Live when necessary.
// 	// if !authResult.XBOXLiveAuthenticated && core.authEnabled {
// 	// 	_ = core.packetConn.WritePacket(&packet.Disconnect{Message: text.Colourf("<red>You must be logged in with XBOX Live to join.</red>")})
// 	// 	return fmt.Errorf("connection %v was not authenticated to XBOX Live", core.RemoteAddr())
// 	// }
// 	if err := core.enableEncryption(authResult.PublicKey); err != nil {
// 		return fmt.Errorf("error enabling encryption: %v", err)
// 	}
// 	return nil
// }

// handleClientToServerHandshake handles an incoming ClientToServerHandshake packet.
func (core *Core) handleClientToServerHandshake() error {
	// The next expected packet is a resource pack client response.
	core.expect(packet.IDResourcePackClientResponse, packet.IDClientCacheStatus)
	if err := core.packetConn.WritePacket(&packet.PlayStatus{Status: packet.PlayStatusLoginSuccess}); err != nil {
		return fmt.Errorf("error sending play status login success: %v", err)
	}
	pk := &packet.ResourcePacksInfo{TexturePackRequired: core.texturePacksRequired}
	for _, pack := range core.resourcePacks {
		// If it has behaviours, add it to the behaviour pack list. If not, we add it to the texture packs
		// list.
		if pack.HasBehaviours() {
			behaviourPack := protocol.BehaviourPackInfo{UUID: pack.UUID(), Version: pack.Version(), Size: uint64(pack.Len())}
			if pack.HasScripts() {
				// One of the resource packs has scripts, so we set HasScripts in the packet to true.
				pk.HasScripts = true
				behaviourPack.HasScripts = true
			}
			pk.BehaviourPacks = append(pk.BehaviourPacks, behaviourPack)
			continue
		}
		texturePack := protocol.TexturePackInfo{UUID: pack.UUID(), Version: pack.Version(), Size: uint64(pack.Len())}
		if pack.Encrypted() {
			texturePack.ContentKey = pack.ContentKey()
			texturePack.ContentIdentity = pack.Manifest().Header.UUID
		}
		pk.TexturePacks = append(pk.TexturePacks, texturePack)
	}
	// Finally we send the packet after the play status.
	if err := core.packetConn.WritePacket(pk); err != nil {
		return fmt.Errorf("error sending resource packs info: %v", err)
	}
	return nil
}

// saltClaims holds the claims for the salt sent by the server in the ServerToClientHandshake packet.
type saltClaims struct {
	Salt string `json:"salt"`
}

// handleServerToClientHandshake handles an incoming ServerToClientHandshake packet. It initialises encryption
// on the client side of the connection, using the hash and the public key from the server exposed in the
// packet.
func (core *Core) handleServerToClientHandshake(pk *packet.ServerToClientHandshake) error {
	tok, err := jwt.ParseSigned(string(pk.JWT))
	if err != nil {
		return fmt.Errorf("parse server token: %w", err)
	}
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	raw, _ := tok.Headers[0].ExtraHeaders["x5u"]
	kStr, _ := raw.(string)

	pub := new(ecdsa.PublicKey)
	if err := login.ParsePublicKey(kStr, pub); err != nil {
		return fmt.Errorf("parse server public key: %w", err)
	}

	var c saltClaims
	if err := tok.Claims(pub, &c); err != nil {
		return fmt.Errorf("verify claims: %w", err)
	}
	c.Salt = strings.TrimRight(c.Salt, "=")
	salt, err := base64.RawStdEncoding.DecodeString(c.Salt)
	if err != nil {
		return fmt.Errorf("error base64 decoding ServerToClientHandshake salt: %v", err)
	}

	x, _ := pub.Curve.ScalarMult(pub.X, pub.Y, core.PrivateKey.D.Bytes())
	// Make sure to pad the shared secret up to 96 bytes.
	sharedSecret := append(bytes.Repeat([]byte{0}, 48-len(x.Bytes())), x.Bytes()...)

	keyBytes := sha256.Sum256(append(salt, sharedSecret...))

	// Finally we enable encryption for the enc and dec using the secret pubKey bytes we produced.
	core.packetConn.EnableEncryption(keyBytes)
	// core.dec.EnableEncryption(keyBytes)

	// We write a ClientToServerHandshake packet (which has no payload) as a response.
	_ = core.packetConn.WritePacket(&packet.ClientToServerHandshake{})
	return nil
}

// handleClientCacheStatus handles a ClientCacheStatus packet sent by the client. It specifies if the client
// has support for the client blob cache.
func (core *Core) handleClientCacheStatus(pk *packet.ClientCacheStatus) error {
	core.EnableClientCache = pk.Enabled
	return nil
}

// handleResourcePacksInfo handles a ResourcePacksInfo packet sent by the server. The client responds by
// sending the packs it needs downloaded.
func (core *Core) handleResourcePacksInfo(pk *packet.ResourcePacksInfo) error {
	// First create a new resource pack queue with the information in the packet so we can download them
	// properly later.
	totalPacks := len(pk.TexturePacks) + len(pk.BehaviourPacks)
	core.packQueue = &resourcePackQueue{
		packAmount:       totalPacks,
		downloadingPacks: make(map[string]downloadingPack),
		awaitingPacks:    make(map[string]*downloadingPack),
	}
	packsToDownload := make([]string, 0, totalPacks)

	for _, pack := range pk.TexturePacks {
		if _, ok := core.packQueue.downloadingPacks[pack.UUID]; ok {
			core.ErrorLog.Printf("duplicate texture pack entry %v in resource pack info\n", pack.UUID)
			core.packQueue.packAmount--
			continue
		}
		// This UUID_Version is a hack Mojang put in place.
		packsToDownload = append(packsToDownload, pack.UUID+"_"+pack.Version)
		core.packQueue.downloadingPacks[pack.UUID] = downloadingPack{
			size:       pack.Size,
			buf:        bytes.NewBuffer(make([]byte, 0, pack.Size)),
			newFrag:    make(chan []byte),
			contentKey: pack.ContentKey,
		}
	}
	for _, pack := range pk.BehaviourPacks {
		if _, ok := core.packQueue.downloadingPacks[pack.UUID]; ok {
			core.ErrorLog.Printf("duplicate behaviour pack entry %v in resource pack info\n", pack.UUID)
			core.packQueue.packAmount--
			continue
		}
		// This UUID_Version is a hack Mojang put in place.
		packsToDownload = append(packsToDownload, pack.UUID+"_"+pack.Version)
		core.packQueue.downloadingPacks[pack.UUID] = downloadingPack{
			size:       pack.Size,
			buf:        bytes.NewBuffer(make([]byte, 0, pack.Size)),
			newFrag:    make(chan []byte),
			contentKey: pack.ContentKey,
		}
	}

	if len(packsToDownload) != 0 {
		core.expect(packet.IDResourcePackDataInfo, packet.IDResourcePackChunkData)
		_ = core.packetConn.WritePacket(&packet.ResourcePackClientResponse{
			Response:        packet.PackResponseSendPacks,
			PacksToDownload: packsToDownload,
		})
		return nil
	}
	core.expect(packet.IDResourcePackStack)

	_ = core.packetConn.WritePacket(&packet.ResourcePackClientResponse{Response: packet.PackResponseAllPacksDownloaded})
	return nil
}

// handleResourcePackStack handles a ResourcePackStack packet sent by the server. The stack defines the order
// that resource packs are applied in.
func (core *Core) handleResourcePackStack(pk *packet.ResourcePackStack) error {
	// We currently don't apply resource packs in any way, so instead we just check if all resource packs in
	// the stacks are also downloaded.
	for _, pack := range pk.TexturePacks {
		for i, behaviourPack := range pk.BehaviourPacks {
			if pack.UUID == behaviourPack.UUID {
				// We had a behaviour pack with the same UUID as the texture pack, so we drop the texture
				// pack and log it.
				core.ErrorLog.Printf("dropping behaviour pack with UUID %v due to a texture pack with the same UUID\n", pack.UUID)
				pk.BehaviourPacks = append(pk.BehaviourPacks[:i], pk.BehaviourPacks[i+1:]...)
			}
		}
		if !core.hasPack(pack.UUID, pack.Version, false) {
			return fmt.Errorf("texture pack {uuid=%v, version=%v} not downloaded", pack.UUID, pack.Version)
		}
	}
	for _, pack := range pk.BehaviourPacks {
		if !core.hasPack(pack.UUID, pack.Version, true) {
			return fmt.Errorf("behaviour pack {uuid=%v, version=%v} not downloaded", pack.UUID, pack.Version)
		}
	}
	core.expect(packet.IDStartGame)
	_ = core.packetConn.WritePacket(&packet.ResourcePackClientResponse{Response: packet.PackResponseCompleted})
	return nil
}

// hasPack checks if the connection has a resource pack downloaded with the UUID and version passed, provided
// the pack either has or does not have behaviours in it.
func (core *Core) hasPack(uuid string, version string, hasBehaviours bool) bool {
	for _, exempted := range exemptedPacks {
		if exempted.uuid == uuid && exempted.version == version {
			// The server may send this resource pack on the stack without sending it in the info, as the client
			// always has it downloaded.
			return true
		}
	}
	core.packMu.Lock()
	defer core.packMu.Unlock()

	for _, ignored := range core.ignoredResourcePacks {
		if ignored.uuid == uuid && ignored.version == version {
			return true
		}
	}
	for _, pack := range core.resourcePacks {
		if pack.UUID() == uuid && pack.Version() == version && pack.HasBehaviours() == hasBehaviours {
			return true
		}
	}
	return false
}

// packChunkSize is the size of a single chunk of data from a resource pack: 512 kB or 0.5 MB
const packChunkSize = 1024 * 128

// handleResourcePackClientResponse handles an incoming resource pack client response packet. The packet is
// handled differently depending on the response.
func (core *Core) handleResourcePackClientResponse(pk *packet.ResourcePackClientResponse) error {
	switch pk.Response {
	case packet.PackResponseRefused:
		// Even though this response is never sent, we handle it appropriately in case it is changed to work
		// correctly again.
		core.CloseWithError(fmt.Errorf("packet response refused"))
		return nil
	case packet.PackResponseSendPacks:
		packs := pk.PacksToDownload
		core.packQueue = &resourcePackQueue{packs: core.resourcePacks}
		if err := core.packQueue.Request(packs); err != nil {
			return fmt.Errorf("error looking up resource packs to download: %v", err)
		}
		// Proceed with the first resource pack download. We run all downloads in sequence rather than in
		// parallel, as it's less prone to packet loss.
		if err := core.nextResourcePackDownload(); err != nil {
			return err
		}
	case packet.PackResponseAllPacksDownloaded:
		pk := &packet.ResourcePackStack{BaseGameVersion: protocol.CurrentVersion, Experiments: []protocol.ExperimentData{{Name: "cameras", Enabled: true}}}
		for _, pack := range core.resourcePacks {
			resourcePack := protocol.StackResourcePack{UUID: pack.UUID(), Version: pack.Version()}
			// If it has behaviours, add it to the behaviour pack list. If not, we add it to the texture packs
			// list.
			if pack.HasBehaviours() {
				pk.BehaviourPacks = append(pk.BehaviourPacks, resourcePack)
				continue
			}
			pk.TexturePacks = append(pk.TexturePacks, resourcePack)
		}
		for _, exempted := range exemptedPacks {
			pk.TexturePacks = append(pk.TexturePacks, protocol.StackResourcePack{
				UUID:    exempted.uuid,
				Version: exempted.version,
			})
		}
		if err := core.packetConn.WritePacket(pk); err != nil {
			return fmt.Errorf("error writing resource pack stack packet: %v", err)
		}
	case packet.PackResponseCompleted:
		core.loggedIn = true
	default:
		return fmt.Errorf("unknown resource pack client response: %v", pk.Response)
	}
	return nil
}

// startGame sends a StartGame packet using the game data of the connection.
func (core *Core) startGame() {
	data := core.gameData
	_ = core.packetConn.WritePacket(&packet.StartGame{
		Difficulty:                   data.Difficulty,
		EntityUniqueID:               data.EntityUniqueID,
		EntityRuntimeID:              data.EntityRuntimeID,
		PlayerGameMode:               data.PlayerGameMode,
		PlayerPosition:               data.PlayerPosition,
		Pitch:                        data.Pitch,
		Yaw:                          data.Yaw,
		WorldSeed:                    data.WorldSeed,
		Dimension:                    data.Dimension,
		WorldSpawn:                   data.WorldSpawn,
		EditorWorld:                  data.EditorWorld,
		CreatedInEditor:              data.CreatedInEditor,
		ExportedFromEditor:           data.ExportedFromEditor,
		PersonaDisabled:              data.PersonaDisabled,
		CustomSkinsDisabled:          data.CustomSkinsDisabled,
		GameRules:                    data.GameRules,
		Time:                         data.Time,
		Blocks:                       data.CustomBlocks,
		Items:                        data.Items,
		AchievementsDisabled:         true,
		Generator:                    1,
		EducationFeaturesEnabled:     true,
		MultiPlayerGame:              true,
		MultiPlayerCorrelationID:     uuid.Must(uuid.NewRandom()).String(),
		CommandsEnabled:              true,
		WorldName:                    data.WorldName,
		LANBroadcastEnabled:          true,
		PlayerMovementSettings:       data.PlayerMovementSettings,
		WorldGameMode:                data.WorldGameMode,
		ServerAuthoritativeInventory: data.ServerAuthoritativeInventory,
		PlayerPermissions:            data.PlayerPermissions,
		Experiments:                  data.Experiments,
		ClientSideGeneration:         data.ClientSideGeneration,
		ChatRestrictionLevel:         data.ChatRestrictionLevel,
		DisablePlayerInteractions:    data.DisablePlayerInteractions,
		BaseGameVersion:              data.BaseGameVersion,
		GameVersion:                  protocol.CurrentVersion,
		UseBlockNetworkIDHashes:      data.UseBlockNetworkIDHashes,
	})
	_ = core.packetConn.Flush()
	core.expect(packet.IDRequestChunkRadius, packet.IDSetLocalPlayerAsInitialised)
}

// nextResourcePackDownload moves to the next resource pack to download and sends a resource pack data info
// packet with information about it.
func (core *Core) nextResourcePackDownload() error {
	pk, ok := core.packQueue.NextPack()
	if !ok {
		return fmt.Errorf("no resource packs to download")
	}
	if err := core.packetConn.WritePacket(pk); err != nil {
		return fmt.Errorf("error sending resource pack data info packet: %v", err)
	}
	// Set the next expected packet to ResourcePackChunkRequest packets.
	core.expect(packet.IDResourcePackChunkRequest)
	return nil
}

// handleResourcePackDataInfo handles a resource pack data info packet, which initiates the downloading of the
// pack by the client.
func (core *Core) handleResourcePackDataInfo(pk *packet.ResourcePackDataInfo) error {
	id := strings.Split(pk.UUID, "_")[0]

	pack, ok := core.packQueue.downloadingPacks[id]
	if !ok {
		// We either already downloaded the pack or we got sent an invalid UUID, that did not match any pack
		// sent in the ResourcePacksInfo packet.
		return fmt.Errorf("unknown pack to download with UUID %v", id)
	}
	if pack.size != pk.Size {
		// Size mismatch: The ResourcePacksInfo packet had a size for the pack that did not match with the
		// size sent here.
		// Netease: disable error log as they're encrypted
		// core.ErrorLog.Printf("pack %v had a different size in the ResourcePacksInfo packet than the ResourcePackDataInfo packet\n", id)
		pack.size = pk.Size
	}

	// Remove the resource pack from the downloading packs and add it to the awaiting packets.
	delete(core.packQueue.downloadingPacks, id)
	core.packQueue.awaitingPacks[id] = &pack

	pack.chunkSize = pk.DataChunkSize

	// The client calculates the chunk count by itself: You could in theory send a chunk count of 0 even
	// though there's data, and the client will still download normally.
	chunkCount := uint32(pk.Size / uint64(pk.DataChunkSize))
	if pk.Size%uint64(pk.DataChunkSize) != 0 {
		chunkCount++
	}

	idCopy := pk.UUID
	go func() {
		for i := uint32(0); i < chunkCount; i++ {
			_ = core.packetConn.WritePacket(&packet.ResourcePackChunkRequest{
				UUID:       idCopy,
				ChunkIndex: i,
			})
			select {
			case <-core.WaitClosed():
				return
			case frag := <-pack.newFrag:
				// Write the fragment to the full buffer of the downloading resource pack.
				_, _ = pack.buf.Write(frag)
			}
		}
		core.packMu.Lock()
		defer core.packMu.Unlock()

		if pack.buf.Len() != int(pack.size) {
			core.ErrorLog.Printf("incorrect resource pack size: expected %v, but got %v\n", pack.size, pack.buf.Len())
			return
		}
		// First parse the resource pack from the total byte buffer we obtained.
		newPack, err := resource.FromBytes(pack.buf.Bytes())
		if err != nil {
			core.ErrorLog.Printf("invalid full resource pack data for UUID %v: %v\n", id, err)
			return
		}
		core.packQueue.packAmount--
		// Finally we add the resource to the resource packs slice.
		core.resourcePacks = append(core.resourcePacks, newPack.WithContentKey(pack.contentKey))
		if core.packQueue.packAmount == 0 {
			core.expect(packet.IDResourcePackStack)
			_ = core.packetConn.WritePacket(&packet.ResourcePackClientResponse{Response: packet.PackResponseAllPacksDownloaded})
		}
	}()
	return nil
}

// handleResourcePackChunkData handles a resource pack chunk data packet, which holds a fragment of a resource
// pack that is being downloaded.
func (core *Core) handleResourcePackChunkData(pk *packet.ResourcePackChunkData) error {
	pk.UUID = strings.Split(pk.UUID, "_")[0]
	pack, ok := core.packQueue.awaitingPacks[pk.UUID]
	if !ok {
		// We haven't received a ResourcePackDataInfo packet from the server, so we can't use this data to
		// download a resource pack.
		return fmt.Errorf("resource pack chunk data for resource pack that was not being downloaded")
	}
	lastData := pack.buf.Len()+int(pack.chunkSize) >= int(pack.size)
	if !lastData && uint32(len(pk.Data)) != pack.chunkSize {
		// The chunk data didn't have the full size and wasn't the last data to be sent for the resource pack,
		// meaning we got too little data.
		return fmt.Errorf("resource pack chunk data had a length of %v, but expected %v", len(pk.Data), pack.chunkSize)
	}
	if pk.ChunkIndex != pack.expectedIndex {
		return fmt.Errorf("resource pack chunk data had chunk index %v, but expected %v", pk.ChunkIndex, pack.expectedIndex)
	}
	pack.expectedIndex++
	pack.newFrag <- pk.Data
	return nil
}

// handleResourcePackChunkRequest handles a resource pack chunk request, which requests a part of the resource
// pack to be downloaded.
func (core *Core) handleResourcePackChunkRequest(pk *packet.ResourcePackChunkRequest) error {
	current := core.packQueue.currentPack
	if current.UUID() != pk.UUID {
		return fmt.Errorf("resource pack chunk request had unexpected UUID: expected %v, but got %v", current.UUID(), pk.UUID)
	}
	if core.packQueue.currentOffset != uint64(pk.ChunkIndex)*packChunkSize {
		return fmt.Errorf("resource pack chunk request had unexpected chunk index: expected %v, but got %v", core.packQueue.currentOffset/packChunkSize, pk.ChunkIndex)
	}
	response := &packet.ResourcePackChunkData{
		UUID:       pk.UUID,
		ChunkIndex: pk.ChunkIndex,
		DataOffset: core.packQueue.currentOffset,
		Data:       make([]byte, packChunkSize),
	}
	core.packQueue.currentOffset += packChunkSize
	// We read the data directly into the response's data.
	if n, err := current.ReadAt(response.Data, int64(response.DataOffset)); err != nil {
		// If we hit an EOF, we don't need to return an error, as we've simply reached the end of the content
		// AKA the last chunk.
		if err != io.EOF {
			return fmt.Errorf("error reading resource pack chunk: %v", err)
		}
		response.Data = response.Data[:n]

		defer func() {
			if !core.packQueue.AllDownloaded() {
				_ = core.nextResourcePackDownload()
			} else {
				core.expect(packet.IDResourcePackClientResponse)
			}
		}()
	}
	if err := core.packetConn.WritePacket(response); err != nil {
		return fmt.Errorf("error writing resource pack chunk data packet: %v", err)
	}

	return nil
}

// handleStartGame handles an incoming StartGame packet. It is the signal that the player has been added to a
// world, and it obtains most of its dedicated properties.
func (core *Core) handleStartGame(pk *packet.StartGame) error {
	core.gameData = game_data.GameData{
		Difficulty:                   pk.Difficulty,
		WorldName:                    pk.WorldName,
		WorldSeed:                    pk.WorldSeed,
		EntityUniqueID:               pk.EntityUniqueID,
		EntityRuntimeID:              pk.EntityRuntimeID,
		PlayerGameMode:               pk.PlayerGameMode,
		BaseGameVersion:              pk.BaseGameVersion,
		PlayerPosition:               pk.PlayerPosition,
		Pitch:                        pk.Pitch,
		Yaw:                          pk.Yaw,
		Dimension:                    pk.Dimension,
		WorldSpawn:                   pk.WorldSpawn,
		EditorWorld:                  pk.EditorWorld,
		CreatedInEditor:              pk.CreatedInEditor,
		ExportedFromEditor:           pk.ExportedFromEditor,
		PersonaDisabled:              pk.PersonaDisabled,
		CustomSkinsDisabled:          pk.CustomSkinsDisabled,
		GameRules:                    pk.GameRules,
		Time:                         pk.Time,
		ServerBlockStateChecksum:     pk.ServerBlockStateChecksum,
		CustomBlocks:                 pk.Blocks,
		Items:                        pk.Items,
		PlayerMovementSettings:       pk.PlayerMovementSettings,
		WorldGameMode:                pk.WorldGameMode,
		ServerAuthoritativeInventory: pk.ServerAuthoritativeInventory,
		PlayerPermissions:            pk.PlayerPermissions,
		ChatRestrictionLevel:         pk.ChatRestrictionLevel,
		DisablePlayerInteractions:    pk.DisablePlayerInteractions,
		ClientSideGeneration:         pk.ClientSideGeneration,
		Experiments:                  pk.Experiments,
		UseBlockNetworkIDHashes:      pk.UseBlockNetworkIDHashes,
	}
	for _, item := range pk.Items {
		if item.Name == "minecraft:shield" {
			core.packetConn.SetShieldID(int32(item.RuntimeID))
		}
	}

	_ = core.packetConn.WritePacket(&packet.RequestChunkRadius{ChunkRadius: 16})
	core.expect(packet.IDChunkRadiusUpdated, packet.IDPlayStatus)
	return nil
}

// handleRequestChunkRadius handles an incoming RequestChunkRadius packet. It sets the initial chunk radius
// of the connection, and spawns the player.
func (core *Core) handleRequestChunkRadius(pk *packet.RequestChunkRadius) error {
	if pk.ChunkRadius < 1 {
		return fmt.Errorf("requested chunk radius must be at least 1, got %v", pk.ChunkRadius)
	}
	core.expect(packet.IDSetLocalPlayerAsInitialised)
	radius := pk.ChunkRadius
	if r := core.gameData.ChunkRadius; r != 0 {
		radius = r
	}
	_ = core.packetConn.WritePacket(&packet.ChunkRadiusUpdated{ChunkRadius: radius})
	core.gameData.ChunkRadius = pk.ChunkRadius

	// The client crashes when not sending all biomes, due to achievements assuming all biomes are present.
	//noinspection SpellCheckingInspection
	if core.biomes == nil {
		const s = `CgAKDWJhbWJvb19qdW5nbGUFCGRvd25mYWxsZmZmPwULdGVtcGVyYXR1cmUzM3M/AAoTYmFtYm9vX2p1bmdsZV9oaWxscwUIZG93bmZhbGxmZmY/BQt0ZW1wZXJhdHVyZTMzcz8ACgViZWFjaAUIZG93bmZhbGzNzMw+BQt0ZW1wZXJhdHVyZc3MTD8ACgxiaXJjaF9mb3Jlc3QFCGRvd25mYWxsmpkZPwULdGVtcGVyYXR1cmWamRk/AAoSYmlyY2hfZm9yZXN0X2hpbGxzBQhkb3duZmFsbJqZGT8FC3RlbXBlcmF0dXJlmpkZPwAKGmJpcmNoX2ZvcmVzdF9oaWxsc19tdXRhdGVkBQhkb3duZmFsbM3MTD8FC3RlbXBlcmF0dXJlMzMzPwAKFGJpcmNoX2ZvcmVzdF9tdXRhdGVkBQhkb3duZmFsbM3MTD8FC3RlbXBlcmF0dXJlMzMzPwAKCmNvbGRfYmVhY2gFCGRvd25mYWxsmpmZPgULdGVtcGVyYXR1cmXNzEw9AAoKY29sZF9vY2VhbgUIZG93bmZhbGwAAAA/BQt0ZW1wZXJhdHVyZQAAAD8ACgpjb2xkX3RhaWdhBQhkb3duZmFsbM3MzD4FC3RlbXBlcmF0dXJlAAAAvwAKEGNvbGRfdGFpZ2FfaGlsbHMFCGRvd25mYWxszczMPgULdGVtcGVyYXR1cmUAAAC/AAoSY29sZF90YWlnYV9tdXRhdGVkBQhkb3duZmFsbM3MzD4FC3RlbXBlcmF0dXJlAAAAvwAKD2RlZXBfY29sZF9vY2VhbgUIZG93bmZhbGwAAAA/BQt0ZW1wZXJhdHVyZQAAAD8AChFkZWVwX2Zyb3plbl9vY2VhbgUIZG93bmZhbGwAAAA/BQt0ZW1wZXJhdHVyZQAAAAAAChNkZWVwX2x1a2V3YXJtX29jZWFuBQhkb3duZmFsbAAAAD8FC3RlbXBlcmF0dXJlAAAAPwAKCmRlZXBfb2NlYW4FCGRvd25mYWxsAAAAPwULdGVtcGVyYXR1cmUAAAA/AAoPZGVlcF93YXJtX29jZWFuBQhkb3duZmFsbAAAAD8FC3RlbXBlcmF0dXJlAAAAPwAKBmRlc2VydAUIZG93bmZhbGwAAAAABQt0ZW1wZXJhdHVyZQAAAEAACgxkZXNlcnRfaGlsbHMFCGRvd25mYWxsAAAAAAULdGVtcGVyYXR1cmUAAABAAAoOZGVzZXJ0X211dGF0ZWQFCGRvd25mYWxsAAAAAAULdGVtcGVyYXR1cmUAAABAAAoNZXh0cmVtZV9oaWxscwUIZG93bmZhbGyamZk+BQt0ZW1wZXJhdHVyZc3MTD4AChJleHRyZW1lX2hpbGxzX2VkZ2UFCGRvd25mYWxsmpmZPgULdGVtcGVyYXR1cmXNzEw+AAoVZXh0cmVtZV9oaWxsc19tdXRhdGVkBQhkb3duZmFsbJqZmT4FC3RlbXBlcmF0dXJlzcxMPgAKGGV4dHJlbWVfaGlsbHNfcGx1c190cmVlcwUIZG93bmZhbGyamZk+BQt0ZW1wZXJhdHVyZc3MTD4ACiBleHRyZW1lX2hpbGxzX3BsdXNfdHJlZXNfbXV0YXRlZAUIZG93bmZhbGyamZk+BQt0ZW1wZXJhdHVyZc3MTD4ACg1mbG93ZXJfZm9yZXN0BQhkb3duZmFsbM3MTD8FC3RlbXBlcmF0dXJlMzMzPwAKBmZvcmVzdAUIZG93bmZhbGzNzEw/BQt0ZW1wZXJhdHVyZTMzMz8ACgxmb3Jlc3RfaGlsbHMFCGRvd25mYWxszcxMPwULdGVtcGVyYXR1cmUzMzM/AAoMZnJvemVuX29jZWFuBQhkb3duZmFsbAAAAD8FC3RlbXBlcmF0dXJlAAAAAAAKDGZyb3plbl9yaXZlcgUIZG93bmZhbGwAAAA/BQt0ZW1wZXJhdHVyZQAAAAAACgRoZWxsBQhkb3duZmFsbAAAAAAFC3RlbXBlcmF0dXJlAAAAQAAKDWljZV9tb3VudGFpbnMFCGRvd25mYWxsAAAAPwULdGVtcGVyYXR1cmUAAAAAAAoKaWNlX3BsYWlucwUIZG93bmZhbGwAAAA/BQt0ZW1wZXJhdHVyZQAAAAAAChFpY2VfcGxhaW5zX3NwaWtlcwUIZG93bmZhbGwAAIA/BQt0ZW1wZXJhdHVyZQAAAAAACgZqdW5nbGUFCGRvd25mYWxsZmZmPwULdGVtcGVyYXR1cmUzM3M/AAoLanVuZ2xlX2VkZ2UFCGRvd25mYWxszcxMPwULdGVtcGVyYXR1cmUzM3M/AAoTanVuZ2xlX2VkZ2VfbXV0YXRlZAUIZG93bmZhbGzNzEw/BQt0ZW1wZXJhdHVyZTMzcz8ACgxqdW5nbGVfaGlsbHMFCGRvd25mYWxsZmZmPwULdGVtcGVyYXR1cmUzM3M/AAoOanVuZ2xlX211dGF0ZWQFCGRvd25mYWxsZmZmPwULdGVtcGVyYXR1cmUzM3M/AAoTbGVnYWN5X2Zyb3plbl9vY2VhbgUIZG93bmZhbGwAAAA/BQt0ZW1wZXJhdHVyZQAAAAAACg5sdWtld2FybV9vY2VhbgUIZG93bmZhbGwAAAA/BQt0ZW1wZXJhdHVyZQAAAD8ACgptZWdhX3RhaWdhBQhkb3duZmFsbM3MTD8FC3RlbXBlcmF0dXJlmpmZPgAKEG1lZ2FfdGFpZ2FfaGlsbHMFCGRvd25mYWxszcxMPwULdGVtcGVyYXR1cmWamZk+AAoEbWVzYQUIZG93bmZhbGwAAAAABQt0ZW1wZXJhdHVyZQAAAEAACgptZXNhX2JyeWNlBQhkb3duZmFsbAAAAAAFC3RlbXBlcmF0dXJlAAAAQAAKDG1lc2FfcGxhdGVhdQUIZG93bmZhbGwAAAAABQt0ZW1wZXJhdHVyZQAAAEAAChRtZXNhX3BsYXRlYXVfbXV0YXRlZAUIZG93bmZhbGwAAAAABQt0ZW1wZXJhdHVyZQAAAEAAChJtZXNhX3BsYXRlYXVfc3RvbmUFCGRvd25mYWxsAAAAAAULdGVtcGVyYXR1cmUAAABAAAoabWVzYV9wbGF0ZWF1X3N0b25lX211dGF0ZWQFCGRvd25mYWxsAAAAAAULdGVtcGVyYXR1cmUAAABAAAoPbXVzaHJvb21faXNsYW5kBQhkb3duZmFsbAAAgD8FC3RlbXBlcmF0dXJlZmZmPwAKFW11c2hyb29tX2lzbGFuZF9zaG9yZQUIZG93bmZhbGwAAIA/BQt0ZW1wZXJhdHVyZWZmZj8ACgVvY2VhbgUIZG93bmZhbGwAAAA/BQt0ZW1wZXJhdHVyZQAAAD8ACgZwbGFpbnMFCGRvd25mYWxszczMPgULdGVtcGVyYXR1cmXNzEw/AAobcmVkd29vZF90YWlnYV9oaWxsc19tdXRhdGVkBQhkb3duZmFsbM3MTD8FC3RlbXBlcmF0dXJlmpmZPgAKFXJlZHdvb2RfdGFpZ2FfbXV0YXRlZAUIZG93bmZhbGzNzEw/BQt0ZW1wZXJhdHVyZQAAgD4ACgVyaXZlcgUIZG93bmZhbGwAAAA/BQt0ZW1wZXJhdHVyZQAAAD8ACg1yb29mZWRfZm9yZXN0BQhkb3duZmFsbM3MTD8FC3RlbXBlcmF0dXJlMzMzPwAKFXJvb2ZlZF9mb3Jlc3RfbXV0YXRlZAUIZG93bmZhbGzNzEw/BQt0ZW1wZXJhdHVyZTMzMz8ACgdzYXZhbm5hBQhkb3duZmFsbAAAAAAFC3RlbXBlcmF0dXJlmpmZPwAKD3NhdmFubmFfbXV0YXRlZAUIZG93bmZhbGwAAAA/BQt0ZW1wZXJhdHVyZc3MjD8ACg9zYXZhbm5hX3BsYXRlYXUFCGRvd25mYWxsAAAAAAULdGVtcGVyYXR1cmUAAIA/AAoXc2F2YW5uYV9wbGF0ZWF1X211dGF0ZWQFCGRvd25mYWxsAAAAPwULdGVtcGVyYXR1cmUAAIA/AAoLc3RvbmVfYmVhY2gFCGRvd25mYWxsmpmZPgULdGVtcGVyYXR1cmXNzEw+AAoQc3VuZmxvd2VyX3BsYWlucwUIZG93bmZhbGzNzMw+BQt0ZW1wZXJhdHVyZc3MTD8ACglzd2FtcGxhbmQFCGRvd25mYWxsAAAAPwULdGVtcGVyYXR1cmXNzEw/AAoRc3dhbXBsYW5kX211dGF0ZWQFCGRvd25mYWxsAAAAPwULdGVtcGVyYXR1cmXNzEw/AAoFdGFpZ2EFCGRvd25mYWxszcxMPwULdGVtcGVyYXR1cmUAAIA+AAoLdGFpZ2FfaGlsbHMFCGRvd25mYWxszcxMPwULdGVtcGVyYXR1cmUAAIA+AAoNdGFpZ2FfbXV0YXRlZAUIZG93bmZhbGzNzEw/BQt0ZW1wZXJhdHVyZQAAgD4ACgd0aGVfZW5kBQhkb3duZmFsbAAAAD8FC3RlbXBlcmF0dXJlAAAAPwAKCndhcm1fb2NlYW4FCGRvd25mYWxsAAAAPwULdGVtcGVyYXR1cmUAAAA/AAA=`
		b, _ := base64.StdEncoding.DecodeString(s)
		_ = core.packetConn.WritePacket(&packet.BiomeDefinitionList{
			SerialisedBiomeDefinitions: b,
		})
	} else {
		b, _ := nbt.MarshalEncoding(core.biomes, nbt.NetworkLittleEndian)
		_ = core.packetConn.WritePacket(&packet.BiomeDefinitionList{SerialisedBiomeDefinitions: b})
	}

	_ = core.packetConn.WritePacket(&packet.PlayStatus{Status: packet.PlayStatusPlayerSpawn})
	_ = core.packetConn.WritePacket(&packet.CreativeContent{})
	return nil
}

// handleChunkRadiusUpdated handles an incoming ChunkRadiusUpdated packet, which updates the initial chunk
// radius of the connection.
func (core *Core) handleChunkRadiusUpdated(pk *packet.ChunkRadiusUpdated) error {
	if pk.ChunkRadius < 1 {
		return fmt.Errorf("new chunk radius must be at least 1, got %v", pk.ChunkRadius)
	}
	core.expect(packet.IDPlayStatus)

	core.gameData.ChunkRadius = pk.ChunkRadius
	core.gameDataReceived.Store(true)

	core.tryFinaliseClientConn()
	return nil
}

// handleSetLocalPlayerAsInitialised handles an incoming SetLocalPlayerAsInitialised packet. It is the final
// packet in the spawning sequence and it marks the point where a server sided connection is considered
// logged in.
func (core *Core) handleSetLocalPlayerAsInitialised(pk *packet.SetLocalPlayerAsInitialised) error {
	if pk.EntityRuntimeID != core.gameData.EntityRuntimeID {
		return fmt.Errorf("entity runtime ID mismatch: entity runtime ID in StartGame and SetLocalPlayerAsInitialised packets should be equal")
	}
	if core.waitingForSpawn.CompareAndSwap(true, false) {
		close(core.spawn)
	}
	return nil
}

// handlePlayStatus handles an incoming PlayStatus packet. It reacts differently depending on the status
// found in the packet.
func (core *Core) handlePlayStatus(pk *packet.PlayStatus) error {
	switch pk.Status {
	case packet.PlayStatusLoginSuccess:
		if err := core.packetConn.WritePacket(&packet.ClientCacheStatus{Enabled: core.EnableClientCache}); err != nil {
			return fmt.Errorf("error sending client cache status: %v", err)
		}
		// The next packet we expect is the ResourcePacksInfo packet.
		core.expect(packet.IDResourcePacksInfo)
		return core.packetConn.Flush()
	case packet.PlayStatusLoginFailedClient:
		core.CloseWithError(fmt.Errorf("packet.PlayStatusLoginFailedClient"))
		return fmt.Errorf("client outdated")
	case packet.PlayStatusLoginFailedServer:
		core.CloseWithError(fmt.Errorf("packet.PlayStatusLoginFailedServer"))
		return fmt.Errorf("server outdated")
	case packet.PlayStatusPlayerSpawn:
		// We've spawned and can send the last packet in the spawn sequence.
		core.waitingForSpawn.Store(true)
		core.tryFinaliseClientConn()
		return nil
	case packet.PlayStatusLoginFailedInvalidTenant:
		core.CloseWithError(fmt.Errorf("packet.PlayStatusLoginFailedInvalidTenant"))
		return fmt.Errorf("invalid edu edition game owner")
	case packet.PlayStatusLoginFailedVanillaEdu:
		core.CloseWithError(fmt.Errorf("packet.PlayStatusLoginFailedVanillaEdu"))
		return fmt.Errorf("cannot join an edu edition game on vanilla")
	case packet.PlayStatusLoginFailedEduVanilla:
		core.CloseWithError(fmt.Errorf("packet.PlayStatusLoginFailedEduVanilla"))
		return fmt.Errorf("cannot join a vanilla game on edu edition")
	case packet.PlayStatusLoginFailedServerFull:
		core.CloseWithError(fmt.Errorf("packet.PlayStatusLoginFailedServerFull"))
		return fmt.Errorf("server full")
	case packet.PlayStatusLoginFailedEditorVanilla:
		core.CloseWithError(fmt.Errorf("packet.PlayStatusLoginFailedEditorVanilla"))
		return fmt.Errorf("cannot join a vanilla game on editor")
	case packet.PlayStatusLoginFailedVanillaEditor:
		core.CloseWithError(fmt.Errorf("packet.PlayStatusLoginFailedVanillaEditor"))
		return fmt.Errorf("cannot join an editor game on vanilla")
	default:
		return fmt.Errorf("unknown play status in PlayStatus packet %v", pk.Status)
	}
}

// tryFinaliseClientConn attempts to finalise the client connection by sending
// the SetLocalPlayerAsInitialised packet when if the ChunkRadiusUpdated and
// PlayStatus packets have been sent.
func (core *Core) tryFinaliseClientConn() {
	if core.waitingForSpawn.Load() && core.gameDataReceived.Load() {
		core.waitingForSpawn.Store(false)
		core.gameDataReceived.Store(false)

		close(core.spawn)
		core.loggedIn = true
		_ = core.packetConn.WritePacket(&packet.SetLocalPlayerAsInitialised{EntityRuntimeID: core.gameData.EntityRuntimeID})
	}
}

// // enableEncryption enables encryption on the server side over the connection. It sends an unencrypted
// // handshake packet to the client and enables encryption after that.
// func (core *Core) enableEncryption(clientPublicKey *ecdsa.PublicKey) error {
// 	signer, _ := jose.NewSigner(jose.SigningKey{Key: core.PrivateKey, Algorithm: jose.ES384}, &jose.SignerOptions{
// 		ExtraHeaders: map[jose.HeaderKey]any{"x5u": login.MarshalPublicKey(&core.PrivateKey.PublicKey)},
// 	})
// 	// We produce an encoded JWT using the header and payload above, then we send the JWT in a ServerToClient-
// 	// Handshake packet so that the client can initialise encryption.
// 	serverJWT, err := jwt.Signed(signer).Claims(saltClaims{Salt: base64.RawStdEncoding.EncodeToString(core.Salt)}).CompactSerialize()
// 	if err != nil {
// 		return fmt.Errorf("compact serialise server JWT: %w", err)
// 	}
// 	if err := core.packetConn.WritePacket(&packet.ServerToClientHandshake{JWT: []byte(serverJWT)}); err != nil {
// 		return fmt.Errorf("error sending ServerToClientHandshake packet: %v", err)
// 	}
// 	// Flush immediately as we'll enable encryption after this.
// 	_ = core.packetConn.Flush()

// 	// We first compute the shared secret.
// 	x, _ := clientPublicKey.Curve.ScalarMult(clientPublicKey.X, clientPublicKey.Y, core.PrivateKey.D.Bytes())

// 	sharedSecret := append(bytes.Repeat([]byte{0}, 48-len(x.Bytes())), x.Bytes()...)

// 	keyBytes := sha256.Sum256(append(core.Salt, sharedSecret...))

// 	// Finally we enable encryption for the encoder and decoder using the secret key bytes we produced.
// 	core.packetConn.EnableEncryption(keyBytes)
// 	// core.dec.EnableEncryption(keyBytes)

// 	return nil
// }

// expect sets the packet IDs that are next expected to arrive.
func (core *Core) expect(packetIDs ...uint32) {
	core.expectedIDs.Store(packetIDs)
}
