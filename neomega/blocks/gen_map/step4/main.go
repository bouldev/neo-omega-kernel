package main

import (
	"fmt"
	"neo-omega-kernel/neomega/blocks"
)

func main() {
	total := 0
	for k, v := range blocks.DefaultAnyToNemcConvertor.BaseNames {
		fmt.Printf("%v: %v\n", k, len(v.StatesWithRtidQuickMatch))
		total += len(v.StatesWithRtidQuickMatch)
	}
	fmt.Printf("total: %v\n", total)
}
