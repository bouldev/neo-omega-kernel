package global

import (
	"neo-omega-kernel/neomega/mirror"
)

type ChunkWriteFn func(chunk *mirror.ChunkData)
