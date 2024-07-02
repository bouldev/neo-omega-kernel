package bot_action

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"

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

var containerNameContainerIDMapping = map[string]uint8{
	"blast_furnace":      protocol.ContainerBlastFurnaceIngredient,
	"lit_blast_furnace":  protocol.ContainerBlastFurnaceIngredient,
	"smoker":             protocol.ContainerSmokerIngredient,
	"lit_smoker":         protocol.ContainerSmokerIngredient,
	"furnace":            protocol.ContainerFurnaceIngredient,
	"lit_furnace":        protocol.ContainerFurnaceIngredient,
	"chest":              protocol.ContainerLevelEntity,
	"trapped_chest":      protocol.ContainerLevelEntity,
	"lectern":            protocol.ContainerTypeLectern,
	"hopper":             protocol.ContainerLevelEntity,
	"dispenser":          protocol.ContainerLevelEntity,
	"dropper":            protocol.ContainerLevelEntity,
	"jukebox":            protocol.ContainerTypeJukebox,
	"brewing_stand":      protocol.ContainerBrewingStandInput,
	"undyed_shulker_box": protocol.ContainerShulkerBox,
	"shulker_box":        protocol.ContainerShulkerBox,
	"barrel":             protocol.ContainerBarrel,
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
