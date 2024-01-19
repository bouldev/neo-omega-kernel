package main

import "neo-omega-kernel/neomega/entries/minimal_end_point_entry"

func main() {
	args := minimal_end_point_entry.GetArgs()
	minimal_end_point_entry.Entry(args)
}
