package uqholder

import (
	"bytes"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/encoding/binary_read_write"
	"neo-omega-kernel/neomega/encoding/little_endian"

	"github.com/go-gl/mathgl/mgl32"
)

func (e *ExtendInfoHolder) Marshal() (data []byte, err error) {
	basicWriter := bytes.NewBuffer(nil)
	writer := binary_read_write.WrapBinaryWriter(basicWriter)
	writeBool := func(b bool) error {
		if b {
			return writer.WriteByte(1)
		} else {
			return writer.WriteByte(0)
		}
	}
	// err = little_endian.WriteString(writer, e.WorldName)
	// if err != nil {
	// 	return nil, err
	// }
	// err = writeBool(e.knownWorldName)
	// if err != nil {
	// 	return nil, err
	// }

	err = little_endian.WriteUint16(writer, e.CompressThreshold)
	if err != nil {
		return nil, err
	}
	err = writeBool(e.knownCompressThreshold)
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteInt64(writer, e.CurrentTick)
	if err != nil {
		return nil, err
	}
	err = writeBool(e.knownCurrentTick)
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteFloat32(writer, e.syncRatio)
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteInt32(writer, e.WorldGameMode)
	if err != nil {
		return nil, err
	}
	err = writeBool(e.knownWorldGameMode)
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteUint32(writer, e.WorldDifficulty)
	if err != nil {
		return nil, err
	}
	err = writeBool(e.knownWorldDifficulty)
	if err != nil {
		return nil, err
	}
	// err = little_endian.WriteUint32(writer, e.InventorySlotCount)
	// if err != nil {
	// 	return nil, err
	// }
	// err = writeBool(e.knownInventorySlotCount)
	// if err != nil {
	// 	return nil, err
	// }
	err = little_endian.WriteInt32(writer, e.Time)
	if err != nil {
		return nil, err
	}
	err = writeBool(e.knownTime)
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteInt32(writer, e.DayTime)
	if err != nil {
		return nil, err
	}
	err = writeBool(e.knownDayTime)
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteFloat32(writer, e.DayTimePercent)
	if err != nil {
		return nil, err
	}
	err = writeBool(e.knownDayTimePercent)
	if err != nil {
		return nil, err
	}
	{
		err = little_endian.WriteUint32(writer, uint32(len(e.GameRules)))
		if err != nil {
			return nil, err
		}
		for k, v := range e.GameRules {
			err = little_endian.WriteString(writer, k)
			if err != nil {
				return nil, err
			}
			err = writeBool(v.CanBeModifiedByPlayer)
			if err != nil {
				return nil, err
			}
			err = little_endian.WriteString(writer, v.Value)
			if err != nil {
				return nil, err
			}
		}
	}
	err = writeBool(e.knownGameRules)
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteInt32(writer, e.Dimension)
	if err != nil {
		return nil, err
	}
	err = writeBool(e.knownDimension)
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteUint64(writer, e.botRuntimeIDDup)
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteInt64(writer, e.PositionUpdateTick)
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteFloat32(writer, e.Position[0])
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteFloat32(writer, e.Position[1])
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteFloat32(writer, e.Position[2])
	if err != nil {
		return nil, err
	}
	// err = writeBool(e.knownPosition)
	// if err != nil {
	// 	return nil, err
	// }
	err = writeBool(e.currentContainerOpened)
	if err != nil {
		return nil, err
	}
	if e.currentContainerOpened {
		err = writer.WriteByte(e.currentOpenedContainer.WindowID)
		if err != nil {
			return nil, err
		}
		err = writer.WriteByte(e.currentOpenedContainer.ContainerType)
		if err != nil {
			return nil, err
		}
		err = little_endian.WriteInt32(writer, e.currentOpenedContainer.ContainerPosition[0])
		if err != nil {
			return nil, err
		}
		err = little_endian.WriteInt32(writer, e.currentOpenedContainer.ContainerPosition[1])
		if err != nil {
			return nil, err
		}
		err = little_endian.WriteInt32(writer, e.currentOpenedContainer.ContainerPosition[2])
		if err != nil {
			return nil, err
		}
		err = little_endian.WriteInt64(writer, e.currentOpenedContainer.ContainerEntityUniqueID)
		if err != nil {
			return nil, err
		}
	}
	return basicWriter.Bytes(), err
}

func (e *ExtendInfoHolder) Unmarshal(data []byte) (err error) {
	basicReader := bytes.NewReader(data)
	reader := binary_read_write.WrapBinaryReader(basicReader)
	readBool := func() (bool, error) {
		b, err := reader.ReadByte()
		if err != nil {
			return false, err
		}
		return byteToBool(b)
	}
	// e.WorldName, err = little_endian.String(reader)
	// if err != nil {
	// 	return err
	// }
	// e.knownWorldName, err = readBool()
	// if err != nil {
	// 	return err
	// }
	e.CompressThreshold, err = little_endian.Uint16(reader)
	if err != nil {
		return err
	}
	e.knownCompressThreshold, err = readBool()
	if err != nil {
		return err
	}
	e.CurrentTick, err = little_endian.Int64(reader)
	if err != nil {
		return err
	}
	e.knownCurrentTick, err = readBool()
	if err != nil {
		return err
	}
	e.syncRatio, err = little_endian.Float32(reader)
	if err != nil {
		return err
	}
	e.WorldGameMode, err = little_endian.Int32(reader)
	if err != nil {
		return err
	}
	e.knownWorldGameMode, err = readBool()
	if err != nil {
		return err
	}
	e.WorldDifficulty, err = little_endian.Uint32(reader)
	if err != nil {
		return err
	}
	e.knownWorldDifficulty, err = readBool()
	if err != nil {
		return err
	}
	// e.InventorySlotCount, err = little_endian.Uint32(reader)
	// if err != nil {
	// 	return err
	// }
	// e.knownInventorySlotCount, err = readBool()
	// if err != nil {
	// 	return err
	// }
	e.Time, err = little_endian.Int32(reader)
	if err != nil {
		return err
	}
	e.knownTime, err = readBool()
	if err != nil {
		return err
	}
	e.DayTime, err = little_endian.Int32(reader)
	if err != nil {
		return err
	}
	e.knownDayTime, err = readBool()
	if err != nil {
		return err
	}
	e.DayTimePercent, err = little_endian.Float32(reader)
	if err != nil {
		return err
	}
	e.knownDayTimePercent, err = readBool()
	if err != nil {
		return err
	}
	{
		var length uint32
		if length, err = little_endian.Uint32(reader); err != nil {
			return err
		}
		e.GameRules = make(map[string]*neomega.GameRule, length)
		for i := uint32(0); i < length; i++ {
			key, err := little_endian.String(reader)
			if err != nil {
				return err
			}
			value := &neomega.GameRule{}
			value.CanBeModifiedByPlayer, err = readBool()
			if err != nil {
				return err
			}
			value.Value, err = little_endian.String(reader)
			if err != nil {
				return err
			}
			e.GameRules[key] = value
		}
	}
	e.knownGameRules, err = readBool()
	if err != nil {
		return err
	}
	e.Dimension, err = little_endian.Int32(reader)
	if err != nil {
		return nil
	}
	e.knownDimension, err = readBool()
	if err != nil {
		return err
	}
	e.botRuntimeIDDup, err = little_endian.Uint64(reader)
	if err != nil {
		return err
	}
	e.PositionUpdateTick, err = little_endian.Int64(reader)
	if err != nil {
		return err
	}
	pos := mgl32.Vec3{}
	pos[0], err = little_endian.Float32(reader)
	if err != nil {
		return err
	}
	pos[1], err = little_endian.Float32(reader)
	if err != nil {
		return err
	}
	pos[2], err = little_endian.Float32(reader)
	if err != nil {
		return err
	}
	e.Position = pos
	e.currentContainerOpened, err = readBool()
	if err != nil {
		return err
	}
	if e.currentContainerOpened {
		e.currentOpenedContainer = &packet.ContainerOpen{}
		e.currentOpenedContainer.WindowID, err = reader.ReadByte()
		if err != nil {
			return err
		}
		e.currentOpenedContainer.ContainerType, err = reader.ReadByte()
		if err != nil {
			return err
		}
		e.currentOpenedContainer.ContainerPosition[0], err = little_endian.Int32(reader)
		if err != nil {
			return err
		}
		e.currentOpenedContainer.ContainerPosition[1], err = little_endian.Int32(reader)
		if err != nil {
			return err
		}
		e.currentOpenedContainer.ContainerPosition[2], err = little_endian.Int32(reader)
		if err != nil {
			return err
		}
		e.currentOpenedContainer.ContainerEntityUniqueID, err = little_endian.Int64(reader)
		if err != nil {
			return err
		}
	}
	return nil
}
