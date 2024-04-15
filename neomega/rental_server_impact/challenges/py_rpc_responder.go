package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
)

type CanSolveChallenge interface {
	GetUID() string
	TransferData(content string) (string, error)
	TransferCheckNum(data string) (string, error)
}

type PyRPCResponder struct {
	neomega.MicroOmega
	isCheckNumResponded       bool
	chanCheckNumResponded     chan struct{}
	isGetStartTypeResponded   bool
	chanGetStartTypeResponded chan struct{}
	solveFail                 chan error
	CanSolveChallenge
}

func NewPyRPCResponder(omega neomega.MicroOmega, canSolveChallenge CanSolveChallenge) *PyRPCResponder {
	responser := &PyRPCResponder{
		MicroOmega:                omega,
		chanCheckNumResponded:     make(chan struct{}),
		chanGetStartTypeResponded: make(chan struct{}),
		CanSolveChallenge:         canSolveChallenge,
		solveFail:                 make(chan error),
	}
	omega.GetGameListener().SetTypedPacketCallBack(packet.IDPyRpc, responser.onPyRPC, true)
	return responser
}

func (o *PyRPCResponder) ChallengeCompete(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-o.solveFail:
		return err
	case <-o.chanGetStartTypeResponded:
		if o.isCheckNumResponded {
			return nil
		} else {
			select {
			case <-o.chanCheckNumResponded:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	case <-o.chanCheckNumResponded:
		if o.isGetStartTypeResponded {
			return nil
		} else {
			select {
			case <-o.chanGetStartTypeResponded:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (o *PyRPCResponder) onPyRPC(pk packet.Packet) {
	pkt, ok := pk.(*packet.PyRpc)
	if !ok {
		return
	}
	goContentData := pkt.Value
	if goContentData == nil {
		return
	}
	content, ok := goContentData.([]interface{})
	if !ok || len(content) < 2 {
		return
	}
	command, ok := content[0].(string)
	if !ok {
		return
	}
	data, ok := content[1].([]interface{})
	if !ok {
		return
	}
	if command == "S2CHeartBeat" {
		o.GetGameControl().SendPacket(&packet.PyRpc{
			Value: []interface{}{
				"C2SHeartBeat",
				data,
				nil,
			},
		})
	} else if command == "GetStartType" {
		if len(data) < 1 {
			o.solveFail <- fmt.Errorf(i18n.T(i18n.S_fail_to_get_start_type_data))
		}
		response, err := o.TransferData(data[0].(string))
		if err != nil {
			o.solveFail <- err
		}
		o.GetGameControl().SendPacket(&packet.PyRpc{
			Value: []interface{}{
				"SetStartType",
				[]interface{}{response},
				nil,
			},
		})
		if !o.isGetStartTypeResponded {
			o.isGetStartTypeResponded = true
			close(o.chanGetStartTypeResponded)
		}
	} else if (command == "GetMCPCheckNum") && !o.isCheckNumResponded {
		if len(data) < 2 {
			o.solveFail <- fmt.Errorf(i18n.T(i18n.S_fail_to_get_check_num_data))
		}
		firstArg := data[0].(string)
		secondArg := (data[1].([]interface{}))[0].(string)
		arg, err := json.Marshal([]interface{}{firstArg, secondArg, o.GetMicroUQHolder().GetBotBasicInfo().GetBotUniqueID()})
		if err != nil {
			o.solveFail <- err
		}
		ret, err := o.TransferCheckNum(string(arg))
		if err != nil {
			o.solveFail <- err
		}
		ret_p := []interface{}{}
		json.Unmarshal([]byte(ret), &ret_p)

		if len(ret_p) > 7 {
			ret6, ok := ret_p[6].(float64)
			if ok {
				ret_p[6] = int64(ret6)
			}
		}

		o.GetGameControl().SendPacket(&packet.PyRpc{
			Value: []interface{}{
				"SetMCPCheckNum",
				[]interface{}{
					ret_p,
				},
				nil,
			},
		})
		o.isCheckNumResponded = true
		close(o.chanCheckNumResponded)
	}
}
