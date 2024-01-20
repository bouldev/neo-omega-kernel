package chunk

// blockEntry represents a block as found in a disk save of a world.
type blockEntry struct {
	Name    string                 `nbt:"name"`
	State   map[string]interface{} `nbt:"states"`
	Version int32                  `nbt:"version"`
	ID      int32                  `nbt:"oldid,omitempty"` // PM writes this field, so we allow it anyway to avoid issues loading PM worlds.
	Meta    int16                  `nbt:"val,omitempty"`
}
