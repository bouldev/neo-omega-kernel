package bot_action

import (
	"encoding/json"
	"fmt"
	"math"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/mirror/define"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

func RefreshPosAndDimensionInfo(e *neomega.PosAndDimensionInfo, omega neomega.CmdSender) error {
	if resp := omega.SendWebSocketCmdNeedResponse("querytarget @s").BlockGetResult(); resp == nil {
		return fmt.Errorf("cannot get bot dimension info")
	} else {
		var QueryResults []neomega.QueryResult
		if resp.SuccessCount > 0 {
			for _, v := range resp.OutputMessages {
				for _, j := range v.Parameters {
					err := json.Unmarshal([]byte(j), &QueryResults)
					if err != nil {
						return fmt.Errorf("cannot parse querytarget result from output")
					}
					if len(QueryResults) == 0 {
						return fmt.Errorf("querytarget output contains 0 result")
					}
					e.Dimension = QueryResults[0].Dimension
					e.InOverWorld = e.Dimension == 0
					e.InNether = e.Dimension == 1
					e.InEnd = e.Dimension == 2
					e.HeadPosPrecise = mgl32.Vec3{float32(QueryResults[0].Position.X), float32(QueryResults[0].Position.Y), float32(QueryResults[0].Position.Z)}
					e.FeetPosPrecise = e.HeadPosPrecise.Sub(mgl32.Vec3{0, 1.62, 0})
					e.FeetBlockPos = define.CubePos{
						int(math.Floor(float64(e.FeetPosPrecise.X()))),
						int(math.Floor(float64(e.FeetPosPrecise.Y()))),
						int(math.Floor(float64(e.FeetPosPrecise.Z()))),
					}
					e.HeadBlockPos = e.FeetBlockPos.Add(define.CubePos{0, 1, 0})
					e.YRot = QueryResults[0].YRot
					return nil
				}
			}
			return fmt.Errorf("querytarget output contains 0 result")
		} else {
			return fmt.Errorf("querytarget output fail")
		}
	}
}

const (
	ContainerIDDefault = byte(7)

	ContainerIDInventory = byte(12)
	ContainerIDHotBar    = byte(28)

	ContainerIDFurnace      = byte(25)
	ContainerIDSmoker       = byte(28) // 它和 ContainerIDHotBar 是一样的吗？
	ContainerIDShulkerBox   = byte(30)
	ContainerIDBlastFurnace = byte(45)
	ContainerIDBarrel       = byte(58)
	ContainerIDBrewingStand = byte(59)

	ContainerIDUnknown = byte(255)
)

var containerNameContainerIDMapping = map[string]uint8{
	"blast_furnace":      ContainerIDBlastFurnace,
	"lit_blast_furnace":  ContainerIDBlastFurnace,
	"smoker":             ContainerIDSmoker,
	"lit_smoker":         ContainerIDSmoker,
	"furnace":            ContainerIDFurnace,
	"lit_furnace":        ContainerIDFurnace,
	"chest":              ContainerIDDefault,
	"trapped_chest":      ContainerIDDefault,
	"lectern":            ContainerIDUnknown,
	"hopper":             ContainerIDDefault,
	"dispenser":          ContainerIDDefault,
	"dropper":            ContainerIDDefault,
	"jukebox":            ContainerIDUnknown,
	"brewing_stand":      ContainerIDBrewingStand,
	"undyed_shulker_box": ContainerIDShulkerBox,
	"shulker_box":        ContainerIDShulkerBox,
	"barrel":             ContainerIDBarrel,
}

func getContainerIDMappingByBlockBaseName(blockName string) (id uint8, found bool) {
	blockName = strings.TrimPrefix(blockName, "minecraft:")
	id, found = containerNameContainerIDMapping[blockName]
	if found {
		return id, found
	}
	for name, id := range containerNameContainerIDMapping {
		if strings.Contains(blockName, name) {
			return id, true
		}
	}
	return uint8(255), false
}
