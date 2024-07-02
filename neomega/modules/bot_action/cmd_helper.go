package bot_action

import (
	"fmt"
	"strings"

	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/utils/string_wrapper"

	"github.com/google/uuid"
)

type CommandHelper struct {
	neomega.CmdSender
	uq neomega.MicroUQHolder
	// botExecutePrefix string
}

func NewCommandHelper(sender neomega.CmdSender, uq neomega.MicroUQHolder) neomega.CommandHelper {
	return &CommandHelper{
		CmdSender: sender,
		uq:        uq,
		// botExecutePrefix: fmt.Sprintf("execute as @a[name=\"%v\"] run", uq.GetBotName()),
	}
}

type woCmd struct {
	cmd string
	neomega.CmdSender
}

func (c *woCmd) Send() {
	// fmt.Println(c.cmd)
	c.CmdSender.SendWOCmd(c.cmd)
}

type wsCmd struct {
	cmd string
	neomega.CmdSender
}

func (c *wsCmd) Send() {
	c.CmdSender.SendWebSocketCmdOmitResponse(c.cmd)
}

func (c *wsCmd) SendAndGetResponse() neomega.ResponseHandle {
	return c.CmdSender.SendWebSocketCmdNeedResponse(c.cmd)
}

type playerCmd struct {
	cmd string
	neomega.CmdSender
}

func (c *playerCmd) Send() {
	c.CmdSender.SendPlayerCmdOmitResponse(c.cmd)
}

func (c *playerCmd) SendAndGetResponse() neomega.ResponseHandle {
	return c.CmdSender.SendPlayerCmdNeedResponse(c.cmd)
}

type generalCmd struct {
	cmd string
	neomega.CmdSender
}

func (c *generalCmd) Send() {
	(&woCmd{c.cmd, c.CmdSender}).Send()
}

func (c *generalCmd) AsWebSocket() neomega.CmdCanGetResponse {
	return &wsCmd{c.cmd, c.CmdSender}
}

func (c *generalCmd) AsPlayer() neomega.CmdCanGetResponse {
	return &playerCmd{c.cmd, c.CmdSender}
}

func (c *CommandHelper) constructWOCommand(cmd string) neomega.CmdCannotGetResponse {
	return &woCmd{cmd, c.CmdSender}
}

func (c *CommandHelper) ConstructDimensionLimitedWOCommand(cmd string) neomega.CmdCannotGetResponse {
	dimension, _ := c.uq.GetBotDimension()
	if dimension == 0 {
		return &woCmd{cmd, c.CmdSender}
	} else {
		return &wsCmd{cmd, c.CmdSender}
	}
}

func (c *CommandHelper) constructWebSocketCommand(cmd string) neomega.CmdCanGetResponse {
	return &wsCmd{cmd, c.CmdSender}
}

func (c *CommandHelper) constructPlayerCommand(cmd string) neomega.CmdCanGetResponse {
	return &playerCmd{cmd, c.CmdSender}
}

func (c *CommandHelper) ConstructDimensionLimitedGeneralCommand(cmd string) neomega.GeneralCommand {
	dimension, _ := c.uq.GetBotDimension()
	if dimension != 0 {
		if dimension == 1 {
			cmd = "execute in nether run " + cmd
		} else if dimension == 2 {
			cmd = "execute in the_end run " + cmd
		} else {
			cmd = fmt.Sprintf("execute in dm%v run ", dimension) + cmd
		}
	}
	return &generalCmd{cmd, c.CmdSender}
}

func (c *CommandHelper) ConstructGeneralCommand(cmd string) neomega.GeneralCommand {
	return &generalCmd{cmd, c.CmdSender}
}

func (c *CommandHelper) ReplaceHotBarItemCmd(slotID int32, item string) neomega.CmdCanGetResponse {
	return c.constructWebSocketCommand(fmt.Sprintf("replaceitem entity @s slot.hotbar %v %v", slotID, item))
}

func (c *CommandHelper) ReplaceBotHotBarItemFullCmd(slotID int32, itemName string, count uint8, value int32, components *neomega.ItemComponentsInGiveOrReplace) neomega.CmdCanGetResponse {
	componentsStr := ""
	if components != nil {
		if components.ItemLock != "" {
			fmt.Printf("warning: bot cannot drop or move item when item lock, this makes no sense, so item lock option is wiped\n")
			components.ItemLock = ""
		}
		itemName = strings.TrimSpace(itemName)
		if strings.Contains(itemName, " ") {
			itemName = strings.Split(itemName, " ")[0]
		}
		componentsStr = (components).ToString()
	}
	return c.constructWebSocketCommand(fmt.Sprintf("/replaceitem entity @s slot.hotbar %v destroy %v %v %v "+componentsStr, slotID, itemName, count, value))
}

func (c *CommandHelper) ReplaceContainerItemFullCmd(pos define.CubePos, slotID int32, itemName string, count uint8, value int32, components *neomega.ItemComponentsInGiveOrReplace) neomega.CmdCanGetResponse {
	itemName = strings.TrimSpace(itemName)
	if strings.Contains(itemName, " ") {
		itemName = strings.Split(itemName, " ")[0]
	}
	componentsStr := (components).ToString()
	return c.constructWebSocketCommand(fmt.Sprintf("/replaceitem block %v %v %v slot.container %v destroy %v %v %v "+componentsStr, pos.X(), pos.Y(), pos.Z(), slotID, itemName, count, value))
}

func (c *CommandHelper) ReplaceContainerBlockItemCmd(pos define.CubePos, slotID int32, item string) neomega.CmdCanGetResponse {
	return c.constructWebSocketCommand(fmt.Sprintf("replaceitem block %v %v %v slot.container %v %v", pos.X(), pos.Y(), pos.Z(), slotID, item))
}

func (c *CommandHelper) BackupStructureWithGivenNameCmd(start define.CubePos, size define.CubePos, name string) neomega.CmdCanGetResponse {
	end := start.Add(size).Sub(define.CubePos{1, 1, 1})
	return c.constructWebSocketCommand(fmt.Sprintf(
		"structure save \"%v\" %v %v %v %v %v %v",
		name,
		start.X(), start.Y(), start.Z(),
		end.X(), end.Y(), end.Z(),
	))
}

func (c *CommandHelper) RevertStructureWithGivenNameCmd(start define.CubePos, name string) neomega.CmdCanGetResponse {
	return c.constructWebSocketCommand(fmt.Sprintf(
		"structure load \"%v\" %v %v %v",
		name,
		start.X(), start.Y(), start.Z(),
	))
}

func (c *CommandHelper) GenAutoUnfilteredUUID() string {
	return string_wrapper.ReplaceWithUnfilteredLetter(uuid.New().String())
}

func (c *CommandHelper) BackupStructureWithAutoNameCmd(start define.CubePos, size define.CubePos) (name string, cmd neomega.CmdCanGetResponse) {
	name = c.GenAutoUnfilteredUUID()
	return name, c.BackupStructureWithGivenNameCmd(start, size, name)
}

func (c *CommandHelper) SetBlockCmd(pos define.CubePos, blockString string) neomega.GeneralCommand {
	return c.ConstructDimensionLimitedGeneralCommand(fmt.Sprintf("setblock %v %v %v %v", pos.X(), pos.Y(), pos.Z(), blockString))
}

func (c *CommandHelper) SetBlockRelativeCmd(pos define.CubePos, blockString string) neomega.GeneralCommand {
	return c.ConstructDimensionLimitedGeneralCommand(fmt.Sprintf("setblock ~%v ~%v ~%v %v", pos.X(), pos.Y(), pos.Z(), blockString))
}

func (c *CommandHelper) FillBlocksWithRangeCmd(startPos define.CubePos, endPos define.CubePos, blockString string) neomega.GeneralCommand {
	return c.ConstructDimensionLimitedGeneralCommand(fmt.Sprintf("fill %v %v %v %v %v %v %v", startPos.X(), startPos.Y(), startPos.Z(), endPos.X(), endPos.Y(), endPos.Z(), blockString))
}

func (c *CommandHelper) FillBlocksWithSizeCmd(startPos define.CubePos, size define.CubePos, blockString string) neomega.GeneralCommand {
	endPos := startPos.Add(size).Sub(define.CubePos{1, 1, 1})
	return c.ConstructDimensionLimitedGeneralCommand(fmt.Sprintf("fill %v %v %v %v %v %v %v", startPos.X(), startPos.Y(), startPos.Z(), endPos.X(), endPos.Y(), endPos.Z(), blockString))
}
