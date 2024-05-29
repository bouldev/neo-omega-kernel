package mc_command

import (
	"fmt"
	"neo-omega-kernel/neomega/blocks"
	"neo-omega-kernel/utils/mc_command/token"
	"strconv"
	"strings"
)

type LegacyMCExecuteCommand struct {
	Selector              string
	Pos                   string
	SubCommand            string
	DetectPosIfAny        string
	DetectBlockNameIfAny  string
	DetectBlockValueIfAny string
}

func (c *LegacyMCExecuteCommand) String() string {
	return fmt.Sprintf("execute: <%v> <%v> [<%v><%v,%v>] %v", c.Selector, c.Pos, c.DetectPosIfAny, c.DetectBlockNameIfAny, c.DetectBlockValueIfAny, c.SubCommand)
}

func ParseLegacyMCCommand(command string) *LegacyMCExecuteCommand {
	c := &LegacyMCExecuteCommand{}
	command = strings.TrimSpace(command)
	if !strings.HasPrefix(strings.ToLower(command), "execute") {
		return nil
	}
	command = command[7:]
	reader := CleanStringAndNewSimpleTextReader(command)
	var ok bool
	var t string
	_, _ = token.ReadWhiteSpace(reader)
	ok, t = token.ReadMCSelector(reader)
	if !ok {
		return nil
	}
	c.Selector = t
	_, _ = token.ReadWhiteSpace(reader)
	ok, t = token.ReadPosition(reader)
	if !ok {
		return nil
	}
	c.Pos = t
	_, _ = token.ReadWhiteSpace(reader)
	_, subCommand := token.ReadUntilEnd(reader)
	if !strings.HasPrefix(strings.ToLower(subCommand), "detect") {
		c.SubCommand = subCommand
		return c
	}
	reader = CleanStringAndNewSimpleTextReader(subCommand[6:])
	token.ReadWhiteSpace(reader)
	ok, t = token.ReadPosition(reader)
	if !ok {
		return nil
	}
	c.DetectPosIfAny = t
	token.ReadWhiteSpace(reader)
	ok, t = token.ReadNonWhiteSpace(reader)
	if !ok {
		return nil
	}
	c.DetectBlockNameIfAny = t
	token.ReadWhiteSpace(reader)
	ok, t = token.ReadSignedInteger(reader)
	if !ok {
		return nil
	}
	c.DetectBlockValueIfAny = t
	token.ReadWhiteSpace(reader)
	ok, t = token.ReadUntilEnd(reader)
	if !ok {
		return nil
	}
	c.SubCommand = t
	return c
}

func UpdateBlockDescribe(blockName, blockValueString string) (string, error) {
	blockValueString = strings.TrimSpace(blockValueString)
	blockValue, err := strconv.ParseInt(blockValueString, 10, 64)
	if err != nil {
		return fmt.Sprintf("[ERROR BLOCK: %v %v]", blockName, blockValueString), fmt.Errorf("%v %v is not a legal legacy block description", blockName, blockValueString)
	}
	if blockValue == -1 {
		return fmt.Sprintf("%v []", blockName), nil
	}
	rtid, found := blocks.LegacyBlockToRuntimeID(blockName, uint16(blockValue))
	if !found {
		return fmt.Sprintf("[ERROR BLOCK: %v %v]", blockName, blockValueString), fmt.Errorf("cannot find new block represent for block %v %v", blockName, blockValueString)
	}
	newName, newState, found := blocks.RuntimeIDToBlockNameAndStateStr(rtid)
	if !found {
		return fmt.Sprintf("[ERROR BLOCK: %v %v]", blockName, blockValueString), fmt.Errorf("cannot find new block represent for block %v %v", blockName, blockValueString)
	}
	if newState == "" {
		newState = "[]"
	}
	newState = strings.ReplaceAll(strings.ReplaceAll(newState, " ", ""), ":", "=")
	return newName + " " + newState, nil
}

func UpdateLegacyExecuteCommand(command string) string {
	c := ParseLegacyMCCommand(command)
	if c == nil {
		return command
	}
	newCommand := fmt.Sprintf("execute as %v at @s positioned %v", c.Selector, c.Pos)
	if c.DetectPosIfAny != "" {
		updateBlock, err := UpdateBlockDescribe(c.DetectBlockNameIfAny, c.DetectBlockValueIfAny)
		if err != nil {
			fmt.Println(err)
			return command
		}
		newCommand += fmt.Sprintf(" if block %v %v", c.DetectPosIfAny, updateBlock)
	}
	newCommand += fmt.Sprintf(" run %v", UpdateLegacyExecuteCommand(c.SubCommand))
	return newCommand
}
