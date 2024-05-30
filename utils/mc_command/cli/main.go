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
	command := `execute@a  [name="2401PT", ... ??!! , score={...}]~ ^-123.456~789detect-234.1~^bamboo 0 execute@a~ ^-123.456~789detect-234.1~^stone -1tp @s ~~~`
	fmt.Println(mc_command.UpdateLegacyExecuteCommand(command))
}
