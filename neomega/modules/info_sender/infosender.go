package info_sender

import (
	"encoding/json"
	"fmt"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"strings"
)

func init() {
	if false {
		func(sender neomega.InfoSender) {}(&InfoSender{})
	}
}

type InfoSender struct {
	neomega.InteractCore
	neomega.CmdSender
	neomega.BotBasicInfoHolder
}

func NewInfoSender(interactable neomega.InteractCore, cmdSender neomega.CmdSender, info neomega.BotBasicInfoHolder) neomega.InfoSender {
	return &InfoSender{
		InteractCore:       interactable,
		CmdSender:          cmdSender,
		BotBasicInfoHolder: info,
	}
}

func (i *InfoSender) BotSay(msg string) {
	pk := &packet.Text{
		TextType:         packet.TextTypeChat,
		NeedsTranslation: false,
		SourceName:       i.GetBotName(),
		Message:          msg,
		XUID:             "",
		PlayerRuntimeID:  fmt.Sprintf("%d", i.GetBotRuntimeID()),
	}
	i.SendPacket(pk)
}

func (i *InfoSender) SayTo(target string, msg string) {
	content := toJsonRawString(msg)
	if strings.HasPrefix(target, "@") {
		i.SendWOCmd(fmt.Sprintf("tellraw %v %v", target, content))
	} else {
		i.SendWOCmd(fmt.Sprintf("tellraw \"%v\" %v", target, content))
	}
}

func (i *InfoSender) RawSayTo(target string, msg string) {
	if strings.HasPrefix(target, "@") {
		i.SendWOCmd(fmt.Sprintf("tellraw %v %v", target, msg))
	} else {
		i.SendWOCmd(fmt.Sprintf("tellraw \"%v\" %v", target, msg))
	}
}

type TellrawItem struct {
	Text string `json:"text"`
}
type TellrawStruct struct {
	RawText []TellrawItem `json:"rawtext"`
}

func toJsonRawString(line string) string {
	final := &TellrawStruct{
		RawText: []TellrawItem{{Text: line}},
	}
	content, _ := json.Marshal(final)
	return string(content)
}

func (i *InfoSender) ActionBarTo(target string, msg string) {
	content := toJsonRawString(msg)
	if strings.HasPrefix(target, "@") {
		i.SendWOCmd(fmt.Sprintf("titleraw %v actionbar %v", target, content))
	} else {
		i.SendWOCmd(fmt.Sprintf("titleraw \"%v\" actionbar %v", target, content))
	}
}

func (i *InfoSender) TitleTo(target string, msg string) {
	content := toJsonRawString(msg)
	if strings.HasPrefix(target, "@") {
		i.SendWOCmd(fmt.Sprintf("titleraw %v title %v", target, content))
	} else {
		i.SendWOCmd(fmt.Sprintf("titleraw \"%v\" title %v", target, content))
	}
}

func (i *InfoSender) SubTitleTo(target string, subTitle string, title string) {
	i.TitleTo(target, title)
	content := toJsonRawString(subTitle)
	if strings.HasPrefix(target, "@") {
		i.SendWOCmd(fmt.Sprintf("titleraw %v subtitle %v", target, content))
	} else {
		i.SendWOCmd(fmt.Sprintf("titleraw \"%v\" subtitle %v", target, content))
	}
}
