package main

import (
	"fmt"
	"neo-omega-kernel/utils/mc_command"
)

func main() {
	// reader := mc_command.NewSimpleTextReader([]rune(`~ ^-123.456~`))
	// // ok, read := token.ReadWhiteSpace(reader)
	// // fmt.Println(ok, "'"+read+"'")
	// // fmt.Println(reader)
	// // ok, read = token.ReadNonWhiteSpace(reader)
	// // fmt.Println(ok, "'"+read+"'")
	// // fmt.Println(reader)
	// // fmt.Println(reader)
	// end := &fsm.NodeEnd{}
	// fail := &fsm.NodeBasic{}
	// node := fsm.MakePositionGroupNode()
	// node.Else = fail
	// node.Next = end
	// node.Do(reader, &fsm.Text{})
	// fmt.Println(end, "\n", reader)
	command := ""
	command = `execute@a  [name="2401PT", ... ??!! , score={...}]~ ^-123.456~789detect-234.1~^bamboo 0 execute@a~ ^-123.456~789detect-234.1~^stone -1tp @s ~~~`
	command = `setblock ^^^ wool 2`
	command = `fill 0 0 0 ^10 100 ~20 wool 1 replace stone`
	command = `clone 0 0 0 ^10 100 ~20 ~~~`
	command = `execute @a[scores={撤离=0}] ~ ~ ~ detect ~ ~0.1 ~ carpet 5 scoreboard players add @s 撤离 10`
	// command = `execute@a  [name="2401PT", ... ??!! , score={...}]~ ^-123.456~789detect-234.1~^bamboo 0 execute@a~ ^-123.456~789detect-234.1~^stone -1setblock ^^^ wool 1 keep`
	command = `summon armor_stand -5025 304 -4974 草方包`
	fmt.Println(command)
	fmt.Println(mc_command.UpdateLegacyCommand(command))
}
