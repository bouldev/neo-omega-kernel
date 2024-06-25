package uqholder

import (
	"bytes"
	"errors"
	"fmt"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/encoding/binary_read_write"
	"neo-omega-kernel/neomega/encoding/little_endian"
	"neo-omega-kernel/neomega/minecraft_conn"
	"neo-omega-kernel/nodes/defines"
	"sync"
	"time"
)

const (
	DEBUG = false
)

func init() {
	if false {
		func(holder neomega.MicroUQHolder) {}(&MicroUQHolder{})
	}
}

type MicroUQHolder struct {
	neomega.BotBasicInfoHolder
	neomega.PlayersInfoHolder
	neomega.ExtendInfo
	mu sync.Mutex
}

func NewAccessPointMicroUQHolder(node defines.APINode, conn minecraft_conn.Conn, reactCore neomega.ReactCore) *MicroUQHolder {
	uq := &MicroUQHolder{
		NewBotInfoHolder(conn),
		NewPlayers(),
		NewExtendInfoHolder(conn),
		sync.Mutex{},
	}
	node.ExposeAPI("get-uqholder", func(args defines.Values) (result defines.Values, err error) {
		data, err := uq.Marshal()
		return defines.FromFrags(data), err
	}, false)
	reactCore.SetAnyPacketCallBack(uq.UpdateFromPacket, false)
	return uq
}

func NewEndPointMicroUQHolder(node defines.APINode, reactCore neomega.ReactCore) (uq *MicroUQHolder, err error) {
	rets, err := node.CallWithResponse("get-uqholder", defines.Empty).SetTimeout(time.Second * 3).BlockGetResponse()
	if err != nil {
		return nil, err
	}
	data, err := rets.ToBytes()
	if err != nil {
		return nil, err
	}
	uq = &MicroUQHolder{}
	err = uq.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	reactCore.SetAnyPacketCallBack(uq.UpdateFromPacket, false)
	return uq, nil
}

func (u *MicroUQHolder) GetBotBasicInfo() neomega.BotBasicInfoHolder {
	return u.BotBasicInfoHolder
}

func (u *MicroUQHolder) GetPlayersInfo() neomega.PlayersInfoHolder {
	return u.PlayersInfoHolder
}

func (u *MicroUQHolder) GetExtendInfo() neomega.ExtendInfo {
	return u.ExtendInfo
}

func (u *MicroUQHolder) Marshal() (data []byte, err error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()
	basicWriter := bytes.NewBuffer(nil)
	writer := binary_read_write.WrapBinaryWriter(basicWriter)
	for moduleName, module := range map[string]neomega.UQInfoHolderEntry{
		"BotBasicInfoHolder": u.BotBasicInfoHolder,
		"PlayersInfoHolder":  u.PlayersInfoHolder,
		"ExtendInfo":         u.ExtendInfo,
	} {
		err = little_endian.WriteString(writer, moduleName)
		if err != nil {
			return nil, err
		}
		var subData []byte
		subData, err = module.Marshal()
		if err != nil {
			return nil, err
		}
		err = little_endian.WriteInt64(writer, int64(len(subData)))
		if err != nil {
			return nil, err
		}
		err = writer.Write(subData)
		if err != nil {
			return nil, err
		}
	}
	return basicWriter.Bytes(), err
}

var ErrInvalidUQHolderEntry = errors.New("invalid uqholder entry")

func (u *MicroUQHolder) Unmarshal(data []byte) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.BotBasicInfoHolder == nil {
		u.BotBasicInfoHolder = &BotBasicInfoHolder{}
	}
	if u.PlayersInfoHolder == nil {
		u.PlayersInfoHolder = &Players{}
	}
	if u.ExtendInfo == nil {
		u.ExtendInfo = &ExtendInfoHolder{}
	}
	basicReader := bytes.NewBuffer(data)
	reader := binary_read_write.WrapBinaryReader(basicReader)
	modules := map[string]neomega.UQInfoHolderEntry{
		"BotBasicInfoHolder": u.BotBasicInfoHolder,
		"PlayersInfoHolder":  u.PlayersInfoHolder,
		"ExtendInfo":         u.ExtendInfo,
	}
	for i := 0; i < len(modules); i++ {
		var name string
		name, err := little_endian.String(reader)
		if err != nil {
			return err
		}
		module, ok := modules[name]
		if !ok {
			return ErrInvalidUQHolderEntry
		}
		var subData []byte
		var subDataLen int64
		subDataLen, err = little_endian.Int64(reader)
		if err != nil {
			return err
		}
		subData, err = reader.ReadOut(int(subDataLen))
		if err != nil {
			return err
		}
		err = module.Unmarshal(subData)
		if err != nil {
			return err
		}
		modules[name] = module
	}
	return nil
}

func (u *MicroUQHolder) UpdateFromPacket(packet packet.Packet) {
	u.mu.Lock()
	defer u.mu.Unlock()
	// fmt.Println(packet)
	u.BotBasicInfoHolder.UpdateFromPacket(packet)
	u.PlayersInfoHolder.UpdateFromPacket(packet)
	u.ExtendInfo.UpdateFromPacket(packet)
}
