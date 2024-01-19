package core

import (
	"neo-omega-kernel/minecraft/protocol/packet"
	"reflect"
	"strings"
)

var mcPacketNameIDMapping map[string]uint32
var pool packet.Pool

func init() {
	pool = packet.NewPool()
	mcPacketNameIDMapping = map[string]uint32{}
	for id, pkMaker := range pool {
		pk := pkMaker()
		pkName := reflect.TypeOf(pk).Elem().Name()
		mcPacketNameIDMapping[pkName] = id
	}
}

// stringWantsToIDSet 将字符串形式的协议包名字列表转换为对应的协议包 ID 集合
func stringWantsToIDSet(want []string) map[uint32]bool {
	s := map[uint32]bool{}
	for _, w := range want {
		// 如果字符串为 "any" 或 "all"，则将 s 中添加所有协议包的 ID
		if w == "any" || w == "all" {
			for _, id := range mcPacketNameIDMapping {
				s[id] = true
			}
			continue
		}
		add := true
		if strings.HasPrefix(w, "!") {
			add = false
			w = w[1:]
		}
		if strings.HasPrefix(w, "no") {
			add = false
			w = w[2:]
		}
		// 如果字符串以 "ID" 开头，则去掉 "ID" 前缀
		w = strings.TrimPrefix(w, "ID")
		if id, found := mcPacketNameIDMapping[w]; found {
			if add {
				s[id] = true
			} else {
				delete(s, id)
			}
		}
	}
	return s
}

func (r *ReactCore) GetMCPacketNameIDMapping() map[string]uint32 {
	return mcPacketNameIDMapping
}

func (r *ReactCore) TranslateStringWantsToIDSet(want []string) map[uint32]bool {
	return stringWantsToIDSet(want)
}
