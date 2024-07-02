package main

import "github.com/OmineDev/neomega-core/neomega/entries/minimal_end_point_entry"

func main() {
	args := minimal_end_point_entry.GetArgs()
	minimal_end_point_entry.Entry(args)
}
