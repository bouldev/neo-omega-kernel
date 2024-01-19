package player_interact

import (
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/uqholder"
	"strings"
)

func (i *PlayerInteract) onTextPacket(pk *packet.Text) {
	if pk.Message == "" {
		return
	}
	if pk.SourceName == i.botBasicUQ.GetBotName() {
		return
	}
	splitMessage := strings.Split(pk.Message, " ")
	cleanedMessage := make([]string, 0, len(splitMessage))
	for _, v := range splitMessage {
		if v != "" {
			cleanedMessage = append(cleanedMessage, v)
		}
	}
	chat := &neomega.GameChat{
		Name:          uqholder.ToPlainName(pk.SourceName),
		Msg:           cleanedMessage,
		Type:          pk.TextType,
		RawMsg:        pk.Message,
		RawName:       pk.SourceName,
		RawParameters: pk.Parameters,
		Aux:           nil,
	}
	i.onChat(chat)
}

func (i *PlayerInteract) SetOnChatCallBack(cb func(chat *neomega.GameChat)) {
	i.chatCbs = append(i.chatCbs, cb)
}

func (i *PlayerInteract) SetOnSpecificCommandBlockTellCallBack(commandBlockName string, cb func(chat *neomega.GameChat)) {
	commandBlockName = strings.TrimSuffix(commandBlockName, "§r")
	commandBlockName += "§r"
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.commandBlockTellCbs[commandBlockName]; !ok {
		i.commandBlockTellCbs[commandBlockName] = make([]func(*neomega.GameChat), 0)
	}
	i.commandBlockTellCbs[commandBlockName] = append(i.commandBlockTellCbs[commandBlockName], cb)
}

func (i *PlayerInteract) SetOnSpecificItemMsgCallBack(itemName string, cb func(chat *neomega.GameChat)) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.specificItemMsgCbs[itemName]; !ok {
		i.specificItemMsgCbs[itemName] = make([]func(*neomega.GameChat), 0)
	}
	i.specificItemMsgCbs[itemName] = append(i.specificItemMsgCbs[itemName], cb)
}

func (i *PlayerInteract) InterceptJustNextInput(playerName string, cb func(chat *neomega.GameChat)) {
	i.mu.Lock()
	defer i.mu.Unlock()
	var cbs []func(*neomega.GameChat)
	var ok bool
	if cbs, ok = i.nextMsgCbs[playerName]; !ok {
		cbs = make([]func(*neomega.GameChat), 0)
	}
	cbs = append(cbs, cb)
	i.nextMsgCbs[playerName] = cbs
}

func (i *PlayerInteract) onChat(chat *neomega.GameChat) {
	i.mu.Lock()
	// specific item msg
	if cbs, ok := i.specificItemMsgCbs[chat.RawName]; ok {
		for _, cb := range cbs {
			go cb(chat)
		}
		i.mu.Unlock()
		return
	}
	_, isPlayer := i.playersUQ.GetPlayerByName(chat.Name)
	// command block tell
	if strings.HasSuffix(chat.RawName, "§r") && !isPlayer {
		if chat.Type == packet.TextTypeWhisper {
			if cbs, ok := i.commandBlockTellCbs[chat.RawName]; ok {
				for _, cb := range cbs {
					go cb(chat)
				}
			}
		}
		i.mu.Unlock()
		return
	}
	// player, query/intercepts are responded first-in-last-out
	if cbs, ok := i.nextMsgCbs[chat.Name]; ok {
		if len(cbs) > 0 {
			lastIntercept := cbs[len(cbs)-1]
			go lastIntercept(chat)
			i.nextMsgCbs[chat.Name] = cbs[:len(cbs)-1]
		}
		if len(i.nextMsgCbs[chat.Name]) == 0 {
			delete(i.nextMsgCbs, chat.Name)
		}
		i.mu.Unlock()
		return
	}
	i.mu.Unlock()
	for _, cb := range i.chatCbs {
		go cb(chat)
	}
}
