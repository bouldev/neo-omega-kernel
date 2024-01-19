package main

import access_point "neo-omega-kernel/neomega/entries/access_point_entry"

func main() {
	args := access_point.GetArgs()
	access_point.Entry(args)
}
