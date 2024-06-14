package core_logic

// exemptedResourcePack is a resource pack that is exempted from being downloaded. These packs may be directly
// applied by sending them in the ResourcePackStack packet.
type exemptedResourcePack struct {
	uuid    string
	version string
}

// exemptedPacks is a list of all resource packs that do not need to be downloaded, but may always be applied
// in the ResourcePackStack packet.
var exemptedPacks = []exemptedResourcePack{
	{
		uuid:    "0fba4063-dba1-4281-9b89-ff9390653530",
		version: "1.0.0",
	},
}
