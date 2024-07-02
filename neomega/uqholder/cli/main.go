package main

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega/uqholder"
)

type packetData struct {
	h       *packet.Header
	full    []byte
	payload *bytes.Buffer
}

// parseData parses the packet data slice passed into a packetData struct.
func parseData(data []byte) (*packetData, error) {
	buf := bytes.NewBuffer(data)
	header := &packet.Header{}
	if err := header.Read(buf); err != nil {
		// We don't return this as an error as it's not in the hand of the user to control this. Instead,
		// we return to reading a new packet.
		return nil, fmt.Errorf("error reading packet header: %v", err)
	}
	return &packetData{h: header, full: data, payload: buf}, nil
}

func LoadPackets() []packet.Packet {
	i := 0
	pool := packet.NewPool()
	pkts := make([]packet.Packet, 0)
	for {
		i += 1
		data, err := os.ReadFile(fmt.Sprintf("dumped_packets/%06d.bin", i))
		if err != nil {
			break
		}
		pkData, err := parseData(data)
		if err != nil {
			fmt.Println(err)
			continue
		}
		pkFunc, ok := pool[pkData.h.PacketID]
		if !ok {
			fmt.Printf("packet decode err: unknown packet %v\n", pkData.h.PacketID)
			continue
		}
		pk := pkFunc()
		r := protocol.NewReader(pkData.payload, 0, false)
		pk.Marshal(r)
		if (pkData.payload.Len()) != 0 {
			panic(pkData.payload.Len())
		}
		pkts = append(pkts, pk)
	}
	return pkts
}

func main() {
	pkts := LoadPackets()
	fmt.Printf("%v packets loaded/n", len(pkts))
	players := uqholder.NewPlayers()
	for _, pkt := range pkts {
		packetName := reflect.TypeOf(pkt).String()
		if packetName == "*packet.MoveActorDelta" {
			continue
		}
		if packetName == "*packet.SetActorMotion" {
			continue
		}
		if packetName == "*packet.NetworkChunkPublisherUpdate" {
			continue
		}
		if packetName == "*packet.BlockActorData" {
			continue
		}
		if packetName == "*packet.LevelChunk" {
			// TODO analysis later
			continue
		}
		if packetName == "*packet.UpdateSubChunkBlocks" {
			// TODO analysis later
			continue
		}
		if packetName == "*packet.SetActorData" {
			continue
		}
		if packetName == "*packet.LevelSoundEvent" {
			continue
		}
		if packetName == "*packet.AvailableCommands" {
			// TODO analysis later
			continue
			// fmt.Println(pkt)
		}
		// if packetName == "*packet.UpdateAttributes" {
		// 	fmt.Println(pkt)
		// }
		if packetName == "*packet.CommandOutput" {
			continue
		}
		if packetName == "*packet.AddActor" {
			continue
		}
		if packetName == "*packet.RemoveActor" {
			continue
		}
		if packetName == "*packet.MovePlayer" {
			continue
		}
		if packetName == "*packet.Text" {
			continue
		}
		if packetName == "*packet.TickSync" {
			continue
		}
		if packetName == "*packet.PyRpc" {
			continue
		}
		if packetName == "*packet.ActorEvent" {
			continue
		}
		if packetName == "*packet.LevelEventGeneric" {
			continue
		}
		// if packetName == "*packet.UpdateAbilities" {
		// 	bs, _ := json.Marshal(pkt)
		// 	fmt.Println(string(bs))
		// }
		// if packetName == "*packet.UpdateAdventureSettings" {
		// 	bs, _ := json.Marshal(pkt)
		// 	fmt.Println(string(bs))
		// }

		// fmt.Println(pkt)
		players.UpdateFromPacket(pkt)
		// fmt.Println(packetName)
		// {"EntityUniqueID":-4294967295,"PlayerPermissions":2,"CommandPermissions":3,"Layers":[{"Type":1,"Abilities":524287,"Values":4095,"FlySpeed":0.05,"WalkSpeed":0.1}]}
		// 放置方块 {"EntityUniqueID":-4294967295,"PlayerPermissions":3,"CommandPermissions":3,"Layers":[{"Type":1,"Abilities":524287,"Values":4094,"FlySpeed":0.05,"WalkSpeed":0.1}]}
		// 3552
		// 操作员 2,3
		// 成员 1,0
		// 访客 0,0
		// 自定义 3,0
		// CommandPermissions 只和是否可以使用命令有关（和使用传送无关）
	}
}
