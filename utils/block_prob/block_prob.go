package block_prob

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	happen    bool
	uid       uuid.UUID
	eventName func() string
	timeout   time.Time
}

type BlockProb struct {
	probName      string
	eventChan     chan *Event
	runningEvents map[uuid.UUID]*Event
	checkSchedule time.Duration
}

func (p *BlockProb) MarkEventStartByTimeout(event func() string, timeout time.Duration) uuid.UUID {
	return p.MarkEventStart(event, time.Now().Add(timeout))
}

func (p *BlockProb) MarkEventStart(event func() string, endTime time.Time) uuid.UUID {
	uid := uuid.New()
	select {
	case p.eventChan <- &Event{
		true, uid, event, endTime,
	}:
		break
	case <-time.NewTimer(p.checkSchedule * 2).C:
		panic("event prob blocked")
	}
	return uid
}

func (p *BlockProb) MarkEventFinished(uid uuid.UUID) {
	select {
	case p.eventChan <- &Event{
		happen: false,
		uid:    uid,
	}:
		break
	default:
		panic("event prob blocked")
	}
}

func (p *BlockProb) checkTimeout() {
	nowTime := time.Now()
	for _, ev := range p.runningEvents {
		if ev.timeout.Before(nowTime) {
			// print warning
			fmt.Printf("block prob: %v :event %v block timeout!\n", p.probName, ev.eventName())
		}
	}
}
func (p *BlockProb) startBlockCheckRoutine() {
	ticker := time.NewTicker(p.checkSchedule)
	for {
		if len(p.runningEvents) == 0 {
			ev := <-p.eventChan
			if !ev.happen {
				delete(p.runningEvents, ev.uid)
			} else {
				p.runningEvents[ev.uid] = ev
			}
		} else {
			consumed := false
			for !consumed {
				select {
				case ev := <-p.eventChan:
					if !ev.happen {
						delete(p.runningEvents, ev.uid)
					} else {
						p.runningEvents[ev.uid] = ev
					}
				default:
					consumed = true
				}
			}
			select {
			case ev := <-p.eventChan:
				if !ev.happen {
					delete(p.runningEvents, ev.uid)
				} else {
					p.runningEvents[ev.uid] = ev
				}
			case <-ticker.C:
				p.checkTimeout()
			}
		}
	}
}

// aim to replace timeout check in individual go routine:
// orig:
// flag=make(chan struct{})
//
//	go func(){
//		select{
//		case <-flag:
//		case <-time.NewTimer(time.Second).C:
//			fmt.Println("timeout!")
//		}
//	}
//
// do sth
// close(flag)
// now we don't need to create an individual go routine for each check
func NewBlockProb(probName string, checkSchedule time.Duration) *BlockProb {
	p := &BlockProb{
		probName:      probName,
		eventChan:     make(chan *Event, 128),
		runningEvents: make(map[uuid.UUID]*Event),
		checkSchedule: checkSchedule,
	}
	go p.startBlockCheckRoutine()
	return p
}
