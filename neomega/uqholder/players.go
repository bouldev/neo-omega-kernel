package uqholder

import (
	"bytes"
	"fmt"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/encoding/binary_read_write"
	"neo-omega-kernel/neomega/encoding/little_endian"
	"neo-omega-kernel/utils/sync_wrapper"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

func init() {
	if false {
		func(neomega.PlayersInfoHolder) {}(&Players{})
	}
}

type Players struct {
	playersByUUID                *sync_wrapper.SyncKVMap[uuid.UUID, *Player]
	cachePlayersByEntityUniqueID *sync_wrapper.SyncKVMap[int64, *Player]
	// cachePlayerByRuntimeID       map[uint64]*Player
	cachePlayerByUsername          *sync_wrapper.SyncKVMap[string, *Player]
	pendingAddPlayerPacket         map[int64]*packet.AddPlayer
	pendingAdventureSettingsPacket map[int64]*packet.AdventureSettings
	pendingMu                      sync.Mutex
}

func NewPlayers() *Players {
	return &Players{
		playersByUUID:                  sync_wrapper.NewSyncKVMap[uuid.UUID, *Player](),
		cachePlayersByEntityUniqueID:   sync_wrapper.NewSyncKVMap[int64, *Player](),
		cachePlayerByUsername:          sync_wrapper.NewSyncKVMap[string, *Player](),
		pendingAddPlayerPacket:         make(map[int64]*packet.AddPlayer),
		pendingAdventureSettingsPacket: make(map[int64]*packet.AdventureSettings),
		pendingMu:                      sync.Mutex{},
	}
}

func (p *Players) GetPlayerByUUID(uuid uuid.UUID) (player neomega.PlayerUQReader, found bool) {
	return p.playersByUUID.Get(uuid)
}

func (p *Players) GetPlayerByUUIDString(uuidStr string) (player neomega.PlayerUQReader, found bool) {
	uuid := uuid.UUID{}
	err := uuid.UnmarshalText([]byte(uuidStr))
	if err != nil {
		return nil, false
	}
	return p.GetPlayerByUUID(uuid)
}

func (p *Players) GetPlayerByUniqueID(uniqueID int64) (player neomega.PlayerUQReader, found bool) {
	player, found = p.cachePlayersByEntityUniqueID.Get(uniqueID)
	if found {
		uid, hasUID := player.GetUUID()
		euid, hasEuid := player.GetEntityUniqueID()
		if hasEuid && hasUID {
			if player, found := p.GetPlayerByUUID(uid); found && euid == uniqueID {
				return player, found
			}
		}
		p.cachePlayersByEntityUniqueID.Delete(uniqueID)
	}
	p.playersByUUID.Iter(func(_uid uuid.UUID, _player *Player) bool {
		if _player.EntityUniqueID == uniqueID {
			p.cachePlayersByEntityUniqueID.Set(uniqueID, _player)
			player = _player
			found = true
			return false
		}
		return true
	})
	return
}

func (p *Players) GetPlayerByName(username string) (player neomega.PlayerUQReader, found bool) {
	player, found = p.cachePlayerByUsername.Get(username)
	if found {
		uid, hasUID := player.GetUUID()
		uname, hasUname := player.GetUsername()
		if hasUname && hasUID {
			if player, found := p.GetPlayerByUUID(uid); found && uname == username {
				return player, found
			}
		}
		p.cachePlayerByUsername.Delete(username)
	}
	p.playersByUUID.Iter(func(_uid uuid.UUID, _player *Player) bool {
		if _player.Username == username {
			p.cachePlayerByUsername.Set(username, _player)
			player = _player
			found = true
			return false
		}
		return true
	})
	return
}

func (players *Players) Marshal() (data []byte, err error) {
	basicWriter := bytes.NewBuffer(nil)
	writer := binary_read_write.WrapBinaryWriter(basicWriter)
	playersByUUID := map[uuid.UUID]*Player{}
	players.playersByUUID.Iter(func(_uuid uuid.UUID, _player *Player) bool {
		playersByUUID[_uuid] = _player
		return true
	})
	err = little_endian.WriteUint32(writer, uint32(len(playersByUUID)))
	for _, player := range playersByUUID {
		playerData, err := player.Marshal()
		if err != nil {
			return nil, err
		}
		dataLen := len(playerData)
		err = little_endian.WriteUint32(writer, uint32(dataLen))
		if err != nil {
			return nil, err
		}
		err = writer.Write(playerData)
		if err != nil {
			return nil, err
		}
	}
	return basicWriter.Bytes(), nil
}

func (players *Players) Unmarshal(data []byte) (err error) {
	basicReader := bytes.NewReader(data)
	reader := binary_read_write.WrapBinaryReader(basicReader)

	players.playersByUUID = sync_wrapper.NewSyncKVMap[uuid.UUID, *Player]()
	players.cachePlayersByEntityUniqueID = sync_wrapper.NewSyncKVMap[int64, *Player]()
	// players.cachePlayerByRuntimeID = make(map[uint64]*Player)
	players.cachePlayerByUsername = sync_wrapper.NewSyncKVMap[string, *Player]()
	playersLen, err := little_endian.Uint32(reader)
	if err != nil {
		return err
	}
	for i := uint32(0); i < playersLen; i++ {
		player := NewPlayerUQHolder()
		playerDataLen, err := little_endian.Uint32(reader)
		if err != nil {
			return err
		}
		playerData, err := reader.ReadOut(int(playerDataLen))
		if err != nil {
			return err
		}
		err = player.Unmarshal(playerData)
		if err != nil {
			return err
		}
		players.playersByUUID.Set(player.UUID, player)
		players.cachePlayersByEntityUniqueID.Set(player.EntityUniqueID, player)
		players.cachePlayerByUsername.Set(player.Username, player)
	}
	if players.pendingAddPlayerPacket == nil {
		players.pendingAddPlayerPacket = make(map[int64]*packet.AddPlayer)
	}
	if players.pendingAdventureSettingsPacket == nil {
		players.pendingAdventureSettingsPacket = make(map[int64]*packet.AdventureSettings)
	}
	players.pendingMu = sync.Mutex{}

	return nil
}

func GetStringContents(s string) []string {
	_s := strings.Split(s, " ")
	for i, c := range _s {
		_s[i] = strings.TrimSpace(c)
	}
	ss := make([]string, 0, len(_s))
	for _, c := range _s {
		if c != "" {
			ss = append(ss, c)
		}
	}
	return ss
}

const (
	automataStateBeginningOfWord = 0
	automataStateTakeRune        = 1
)

func ToPlainName(name string) string {
	if !strings.ContainsAny(name, "<>§") {
		return name
	}
	flip := false
	cleanedNameRunes := []rune{}
	for _, r := range name {
		if flip {
			flip = false
			continue
		} else if r == '§' {
			flip = true
			continue
		}
		cleanedNameRunes = append(cleanedNameRunes, r)
	}
	cleanedName := string(cleanedNameRunes)
	if !strings.ContainsAny(cleanedName, "<>") {
		return cleanedName
	}

	automataState := automataStateBeginningOfWord
	lastWord := []rune{}
	for _, r := range cleanedName {
		leftBracket := r == '<'
		rightBracket := r == '>'
		switch automataState {
		case automataStateBeginningOfWord:
			if leftBracket || rightBracket {
				continue
			} else {
				lastWord = []rune{r}
				automataState = automataStateTakeRune
			}
		case automataStateTakeRune:
			if leftBracket || rightBracket {
				automataState = automataStateBeginningOfWord
			} else {
				lastWord = append(lastWord, r)
			}
		}
	}
	return string(lastWord)
}

func (p *Players) AddPlayer(e protocol.PlayerListEntry) *Player {
	player := NewPlayerUQHolder()
	player.setUUID(e.UUID)
	player.setEntityUniqueID(e.EntityUniqueID)
	player.setUsername(ToPlainName(e.Username))
	player.setPlatformChatID(e.PlatformChatID)
	player.setBuildPlatform(e.BuildPlatform)
	player.setSkinID(e.Skin.SkinID)
	player.setLoginTime(time.Now())

	p.playersByUUID.Set(e.UUID, player)
	p.cachePlayersByEntityUniqueID.Set(e.EntityUniqueID, player)
	p.cachePlayerByUsername.Set(ToPlainName(e.Username), player)
	return player
}

func (p *Players) RemovePlayer(e protocol.PlayerListEntry) {
	if player, ok := p.playersByUUID.Get(e.UUID); ok {
		player.Online = false
		p.cachePlayersByEntityUniqueID.Delete(player.EntityUniqueID)
		p.cachePlayerByUsername.Delete(player.Username)
		p.playersByUUID.Delete(e.UUID)
	} else {
		fmt.Println("player not found: ", e.UUID.String())
	}
}

func (uq *Players) UpdateFromPacket(pk packet.Packet) {
	// if pk.ID() == packet.IDAdventureSettings || pk.ID() == packet.IDPlayerList {
	// 	bs, _ := json.Marshal(pk)
	// 	fmt.Println(string(bs))
	// }

	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("UQHolder Update Error: ", r)
			debug.PrintStack()
		}
	}()
	switch p := pk.(type) {
	case *packet.PlayerList:
		if p.ActionType == packet.PlayerListActionAdd {
			for _, e := range p.Entries {
				player := uq.AddPlayer(e)
				uq.pendingMu.Lock()
				if pk, found := uq.pendingAdventureSettingsPacket[e.EntityUniqueID]; found {
					player.setPropertiesFlag(pk.Flags)
					player.setCommandPermissionLevel(pk.CommandPermissionLevel)
					player.setActionPermissions(pk.ActionPermissions)
					player.setOPPermissionLevel(pk.PermissionLevel)
					player.setCustomStoredPermissions(pk.CustomStoredPermissions)
					delete(uq.pendingAdventureSettingsPacket, e.EntityUniqueID)
				}
				if pk, found := uq.pendingAddPlayerPacket[e.EntityUniqueID]; found {
					player.setDeviceID(pk.DeviceID)
					player.setEntityRuntimeID(pk.EntityRuntimeID)
					player.setEntityMetadata(pk.EntityMetadata)
					delete(uq.pendingAddPlayerPacket, e.EntityUniqueID)
				}
				uq.pendingMu.Unlock()
			}
		} else {
			for _, e := range p.Entries {
				uq.RemovePlayer(e)
			}
		}
	case *packet.AdventureSettings:
		playerReader, found := uq.GetPlayerByUniqueID(p.PlayerUniqueID)
		if !found {
			uq.pendingMu.Lock()
			uq.pendingAdventureSettingsPacket[p.PlayerUniqueID] = p
			uq.pendingMu.Unlock()
			return
		}
		player := playerReader.(*Player)
		player.setPropertiesFlag(p.Flags)
		player.setCommandPermissionLevel(p.CommandPermissionLevel)
		player.setActionPermissions(p.ActionPermissions)
		player.setOPPermissionLevel(p.PermissionLevel)
		player.setCustomStoredPermissions(p.CustomStoredPermissions)
	case *packet.AddPlayer:
		playerReader, found := uq.GetPlayerByUniqueID(p.EntityUniqueID)
		if !found {
			uq.pendingMu.Lock()
			uq.pendingAddPlayerPacket[p.PlayerUniqueID] = p
			uq.pendingMu.Unlock()
			return
		}
		player := playerReader.(*Player)
		player.setDeviceID(p.DeviceID)
		player.setEntityRuntimeID(p.EntityRuntimeID)
		player.setEntityMetadata(p.EntityMetadata)
	}
}

func (players *Players) GetAllOnlinePlayers() []neomega.PlayerUQReader {
	playersOut := make([]neomega.PlayerUQReader, 0)
	players.playersByUUID.Iter(func(_uuid uuid.UUID, _player *Player) bool {
		playersOut = append(playersOut, _player)
		return true
	})
	return playersOut
}
