package main

import (
	"github.com/OmineDev/neomega-core/neomega/entries/minimal_client_entry"
)

func main() {
	args := minimal_client_entry.GetArgs()
	minimal_client_entry.Entry(args)
}
