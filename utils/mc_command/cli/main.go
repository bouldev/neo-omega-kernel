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

	// command = `execute@a  [name="2401PT", ... ??!! , score={...}]~ ^-123.456~789detect-234.1~^bamboo 0 execute@a~ ^-123.456~789detect-234.1~^stone -1setblock ^^^ wool 1 keep`
	command = `summon armor_stand -5025 304 -4974 草方包`
	// fmt.Println(command)
	// fmt.Println(mc_command.UpdateLegacyCommand(command))
	command = `title @a[b] title abc`
	command = `title @a[b] title abc @a[c=e] 第一行结尾
	 第二行
	 第3行 @s@a@p[m=1,score={mm=1..9}] @s @a @p [m=1,score={mm=1..9}]
	 第4行
	 第5行 @不是选择器 @@ @@@@ @9 @?
	 第6行
	 \"`
	command = `fill ~~~1 ~~~1 chain_command_block ["conditional_bit"=false,"facing_direction"=3]`
	command = `clone 0 0 0 ^10 100 ~20 ~~~ filtered normal stone 1`
	command = `execute as @a[scores={撤离=0}] at @s positioned ~ ~ ~ if block ~ ~0.1 ~ lime_carpet [] run scoreboard players add @s 撤离 10`
	fmt.Println(command)
	fmt.Println(mc_command.UpdateLegacyCommand(command))
}
