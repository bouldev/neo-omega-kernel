package player_interact

import (
	"fmt"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"sync"

	"github.com/google/uuid"
)

func init() {
	if false {
		func(playerKit neomega.PlayerInteract) {}(&PlayerInteract{})
	}
}

type PlayerInteract struct {
	playersUQ             neomega.PlayersInfoHolder
	botBasicUQ            neomega.BotBasicInfoHolder
	cmdSender             neomega.CmdSender
	info                  neomega.InfoSender
	gameIntractable       neomega.GameIntractable
	chatCbs               []func(chat *neomega.GameChat)
	commandBlockTellCbs   map[string][]func(*neomega.GameChat)
	nextMsgCbs            map[string][]func(*neomega.GameChat)
	specificItemMsgCbs    map[string][]func(*neomega.GameChat)
	playerChangeListeners []func(neomega.PlayerKit, string)
	cachedPlayers         map[uuid.UUID]neomega.PlayerKit
	mu                    sync.Mutex
}

func NewPlayerInteract(
	reactCore neomega.ReactCore,
	playersUQ neomega.PlayersInfoHolder,
	botBasicUQ neomega.BotBasicInfoHolder,
	cmdSender neomega.CmdSender,
	info neomega.InfoSender,
	gameIntractable neomega.GameIntractable,
) neomega.PlayerInteract {
	i := &PlayerInteract{
		playersUQ:             playersUQ,
		botBasicUQ:            botBasicUQ,
		cmdSender:             cmdSender,
		info:                  info,
		gameIntractable:       gameIntractable,
		chatCbs:               make([]func(chat *neomega.GameChat), 0),
		commandBlockTellCbs:   make(map[string][]func(*neomega.GameChat)),
		nextMsgCbs:            make(map[string][]func(*neomega.GameChat)),
		specificItemMsgCbs:    make(map[string][]func(*neomega.GameChat)),
		cachedPlayers:         make(map[uuid.UUID]neomega.PlayerKit),
		playerChangeListeners: make([]func(neomega.PlayerKit, string), 0),
		mu:                    sync.Mutex{},
	}
	reactCore.SetTypedPacketCallBack(packet.IDText, func(p packet.Packet) {
		i.onTextPacket(p.(*packet.Text))
	}, true)
	reactCore.SetTypedPacketCallBack(packet.IDPlayerList, func(p packet.Packet) {
		i.onPlayerList(p.(*packet.PlayerList))
	}, true)
	for _, player := range playersUQ.GetAllOnlinePlayers() {
		uuid, _ := player.GetUUID()
		name, _ := player.GetUsername()
		i.cachedPlayers[uuid] = &PlayerKit{player, name, i}
	}
	return i
}

func (i *PlayerInteract) onPlayerList(pk *packet.PlayerList) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if pk.ActionType == packet.PlayerListActionAdd {
		for _, entry := range pk.Entries {
			i.onAddPlayer(entry.UUID)
		}
	} else {
		for _, entry := range pk.Entries {
			i.onRemovePlayer(entry.UUID)
		}
	}

}

func (i *PlayerInteract) onAddPlayer(uid uuid.UUID) {
	player, found := i.playersUQ.GetPlayerByUUID(uid)
	if !found {
		fmt.Printf("player not found: %s", uid.String())
	}
	name, _ := player.GetUsername()
	i.cachedPlayers[uid] = &PlayerKit{player, name, i}
	for _, cb := range i.playerChangeListeners {
		go cb(i.cachedPlayers[uid], "online")
	}
}

func (i *PlayerInteract) onRemovePlayer(uid uuid.UUID) {
	player, found := i.cachedPlayers[uid]
	if !found {
		return
	}
	name, found := player.GetUsername()
	if found {
		delete(i.nextMsgCbs, name)
	}
	for _, cb := range i.playerChangeListeners {
		go cb(player, "offline")
	}
	delete(i.cachedPlayers, uid)
}

func (i *PlayerInteract) ListenPlayerChange(cb func(player neomega.PlayerKit, action string)) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.playerChangeListeners = append(i.playerChangeListeners, cb)
	func() {
		tmp := make([]neomega.PlayerKit, 0, len(i.cachedPlayers))
		for _, player := range i.cachedPlayers {
			tmp = append(tmp, player)
		}
		for _, player := range tmp {
			go cb(player, "exist")
		}
	}()
}

func (i *PlayerInteract) ListAllPlayers() []neomega.PlayerKit {
	players := make([]neomega.PlayerKit, 0)
	for _, player := range i.playersUQ.GetAllOnlinePlayers() {
		name, _ := player.GetUsername()
		players = append(players, &PlayerKit{player, name, i})
	}
	return players
}

func (i *PlayerInteract) GetPlayerKit(name string) (playerKit neomega.PlayerKit, found bool) {
	player, found := i.playersUQ.GetPlayerByName(name)
	if !found {
		return nil, false
	}
	return &PlayerKit{player, name, i}, true
}

func (i *PlayerInteract) GetPlayerKitByUUID(uuid uuid.UUID) (playerKit neomega.PlayerKit, found bool) {
	player, found := i.playersUQ.GetPlayerByUUID(uuid)
	if !found {
		return nil, false
	}
	name, _ := player.GetUsername()
	return &PlayerKit{player, name, i}, true
}

func (i *PlayerInteract) GetPlayerKitByUUIDString(uuidStr string) (playerKit neomega.PlayerKit, found bool) {
	player, found := i.playersUQ.GetPlayerByUUIDString(uuidStr)
	if !found {
		return nil, false
	}
	name, _ := player.GetUsername()
	return &PlayerKit{player, name, i}, true
}

func (i *PlayerInteract) GetPlayerKitByUniqueID(uniqueID int64) (playerKit neomega.PlayerKit, found bool) {
	player, found := i.playersUQ.GetPlayerByUniqueID(uniqueID)
	if !found {
		return nil, false
	}
	name, _ := player.GetUsername()
	return &PlayerKit{player, name, i}, true
}
