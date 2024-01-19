package main

import (
	"neo-omega-kernel/neomega/entries/minimal_client_entry"
)

func main() {
	args := minimal_client_entry.GetArgs()
	minimal_client_entry.Entry(args)
}
