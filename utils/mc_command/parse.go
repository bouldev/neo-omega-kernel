package mc_command

import (
	"fmt"
	"neo-omega-kernel/utils/mc_command/token"
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

func ParseLegacyMCExecuteCommand(command string) *LegacyMCExecuteCommand {
	c := &LegacyMCExecuteCommand{}
	command = strings.TrimSpace(command)
	reader := CleanStringAndNewSimpleTextReader(command)
	var ok bool
	var t string
	token.ReadSpecific(reader, "/", true)
	token.ReadWhiteSpace(reader)
	ok, _ = token.ReadSpecific(reader, "execute", true)
	if !ok {
		return nil
	}
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
	back := reader.Snapshot()
	ok, _ = token.ReadSpecific(reader, "detect", true)
	if !ok {
		back()
		_, subCommand := token.ReadUntilEnd(reader)
		c.SubCommand = subCommand
		return c
	}
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

type LegacySetBlockCommand struct {
	Pos             string
	BlockName       string
	BlockValueIfAny string
	OtherOptions    string
}

func ParseLegacySetBlockCommand(command string) *LegacySetBlockCommand {
	origCommand := command
	command = strings.TrimSpace(origCommand)
	reader := CleanStringAndNewSimpleTextReader(command)
	var ok bool
	var t string
	token.ReadSpecific(reader, "/", true)
	token.ReadWhiteSpace(reader)
	ok, _ = token.ReadSpecific(reader, "setblock", true)
	if !ok {
		return nil
	}
	_, _ = token.ReadWhiteSpace(reader)
	c := &LegacySetBlockCommand{}
	// position
	ok, t = token.ReadPosition(reader)
	if !ok {
		return nil
	}
	c.Pos = t
	token.ReadWhiteSpace(reader)
	// block name
	ok, t = token.ReadNonWhiteSpace(reader)
	if !ok {
		return nil
	}
	c.BlockName = t
	token.ReadWhiteSpace(reader)
	// block Value
	ok, t = token.ReadSignedInteger(reader)
	if ok {
		c.BlockValueIfAny = t
	}
	// other options
	ok, t = token.ReadUntilEnd(reader)
	if ok {
		c.OtherOptions += t
	}
	return c
}

type LegacyFillCommand struct {
	StartPos, EndPos       string
	BlockName              string
	BlockValueIfAny        string
	OtherOptions           string
	ReplaceBlockNameIfAny  string
	ReplaceBlockValueIfAny string
}

func ParseLegacyFillCommand(command string) *LegacyFillCommand {
	origCommand := command
	command = strings.TrimSpace(origCommand)
	reader := CleanStringAndNewSimpleTextReader(command)
	var ok bool
	var t string
	token.ReadSpecific(reader, "/", true)
	token.ReadWhiteSpace(reader)
	ok, _ = token.ReadSpecific(reader, "fill", true)
	if !ok {
		return nil
	}
	_, _ = token.ReadWhiteSpace(reader)
	c := &LegacyFillCommand{}
	// start position
	ok, t = token.ReadPosition(reader)
	if !ok {
		return nil
	}
	c.StartPos = t
	token.ReadWhiteSpace(reader)
	// end position
	ok, t = token.ReadPosition(reader)
	if !ok {
		return nil
	}
	c.EndPos = t
	token.ReadWhiteSpace(reader)
	// block name
	ok, t = token.ReadNonWhiteSpace(reader)
	if !ok {
		return nil
	}
	c.BlockName = t
	token.ReadWhiteSpace(reader)
	// block Value
	ok, t = token.ReadSignedInteger(reader)
	if ok {
		c.BlockValueIfAny = t
	}
	back := reader.Snapshot()
	token.ReadWhiteSpace(reader)
	ok, _ = token.ReadSpecific(reader, "replace", true)
	if !ok {
		back()
		_, t = token.ReadUntilEnd(reader)
		c.OtherOptions += t
		return c
	}
	token.ReadWhiteSpace(reader)
	// replaced block name
	ok, t = token.ReadNonWhiteSpace(reader)
	if !ok {
		return nil
	}
	c.ReplaceBlockNameIfAny = t
	token.ReadWhiteSpace(reader)
	// block Value
	ok, t = token.ReadSignedInteger(reader)
	if ok {
		c.ReplaceBlockValueIfAny = t
	}
	return c
}

type LegacyCloneCommand struct {
	StartPos, EndPos, TargetPos string
	IsFiltered                  bool
	OtherOptions                string
	ModeIfFiltered              string
	BlockNameIfFiltered         string
	BlockValueIfFiltered        string
}

func ParseLegacyCloneCommand(command string) *LegacyCloneCommand {
	origCommand := command
	command = strings.TrimSpace(origCommand)
	reader := CleanStringAndNewSimpleTextReader(command)
	var ok bool
	var t string
	token.ReadSpecific(reader, "/", true)
	token.ReadWhiteSpace(reader)
	ok, _ = token.ReadSpecific(reader, "clone", true)
	if !ok {
		return nil
	}
	_, _ = token.ReadWhiteSpace(reader)
	c := &LegacyCloneCommand{}
	// start position
	ok, t = token.ReadPosition(reader)
	if !ok {
		return nil
	}
	c.StartPos = t
	token.ReadWhiteSpace(reader)
	// end position
	ok, t = token.ReadPosition(reader)
	if !ok {
		return nil
	}
	c.EndPos = t
	token.ReadWhiteSpace(reader)
	// target position
	ok, t = token.ReadPosition(reader)
	if !ok {
		return nil
	}
	c.TargetPos = t
	token.ReadWhiteSpace(reader)

	back := reader.Snapshot()
	token.ReadWhiteSpace(reader)
	ok, _ = token.ReadSpecific(reader, "filtered", true)
	if !ok {
		back()
		_, t = token.ReadUntilEnd(reader)
		c.OtherOptions += t
		return c
	}
	c.IsFiltered = true
	// fmt.Println(reader)
	token.ReadWhiteSpace(reader)
	ok, t = token.ReadNonWhiteSpace(reader)
	if !ok {
		return nil
	}
	c.ModeIfFiltered = t
	// fmt.Println(reader)

	token.ReadWhiteSpace(reader)
	// block name
	ok, t = token.ReadNonWhiteSpace(reader)
	if !ok {
		return nil
	}
	c.BlockNameIfFiltered = t
	token.ReadWhiteSpace(reader)
	// block Value
	ok, t = token.ReadSignedInteger(reader)
	if ok {
		c.BlockValueIfFiltered = t
	}
	return c
}

type LegacyTestForBlockCommand struct {
	Pos             string
	BlockName       string
	BlockValueIfAny string
}

func ParseLegacyTestForBlockCommand(command string) *LegacyTestForBlockCommand {
	origCommand := command
	command = strings.TrimSpace(origCommand)
	reader := CleanStringAndNewSimpleTextReader(command)
	var ok bool
	var t string
	token.ReadSpecific(reader, "/", true)
	token.ReadWhiteSpace(reader)
	ok, _ = token.ReadSpecific(reader, "testforblock", true)
	if !ok {
		return nil
	}
	_, _ = token.ReadWhiteSpace(reader)
	c := &LegacyTestForBlockCommand{}
	// position
	ok, t = token.ReadPosition(reader)
	if !ok {
		return nil
	}
	c.Pos = t
	token.ReadWhiteSpace(reader)
	// block name
	ok, t = token.ReadNonWhiteSpace(reader)
	if !ok {
		return nil
	}
	c.BlockName = t
	token.ReadWhiteSpace(reader)
	// block Value
	ok, t = token.ReadSignedInteger(reader)
	if ok {
		c.BlockValueIfAny = t
	}
	return c
}
