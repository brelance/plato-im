package domain

import (
	"sort"
	"sync"

	"github.com/brelance/plato/common/logger"
	"github.com/brelance/plato/ipconfig/source"
)

// Share by go routines
type Dispatcher struct {
	candidateTable map[string]*Endport
	sync.RWMutex
}

var dp *Dispatcher

func Init() {
	dp = &Dispatcher{}
	dp.candidateTable = make(map[string]*Endport)
	go func() {
		for event := range source.EventChan() {
			logger.Logger.Debug().Msgf("Dispatcher received the event %v", event)
			switch event.Type {
			case source.AddNodeEvent:
				dp.addNode(*event)
			case source.DelNodeEvent:
				dp.delNode(*event)
			}
		}
	}()
}

func Dispatch(ctx *IpConfContext) []*Endport {
	eds := dp.getCandiateList()
	for _, ed := range eds {
		ed.CalculateScore(ctx)
	}

	sort.Slice(eds, func(i, j int) bool {
		if eds[i].ActiveScore > eds[j].ActiveScore {
			return true
		}

		if eds[i].ActiveScore == eds[j].ActiveScore {
			return eds[i].StaticScore > eds[j].StaticScore
		}
		return false
	})
	return eds
}

func (dp *Dispatcher) getCandiateList() []*Endport {
	dp.RLock()
	defer dp.RUnlock()
	candidateList := make([]*Endport, 0, len(dp.candidateTable))
	for _, ed := range dp.candidateTable {
		candidateList = append(candidateList, ed)
	}
	return candidateList
}

func (dp *Dispatcher) addNode(event source.Event) {
	dp.Lock()
	defer dp.Unlock()
	var ed *Endport
	var ok bool

	newStat := &Stat{
		ConnNum:      event.ConnNum,
		MessageBytes: event.MessageBytes,
	}

	if ed, ok = dp.candidateTable[event.Key()]; !ok {
		ed = NewEndPort(event.IP, event.Port)
		logger.Logger.Debug().Msgf("Create Endport %s:%s", ed.IP, ed.Port)
		dp.candidateTable[event.Key()] = ed
	}

	ed.UpdateStat(newStat)
	logger.Logger.Debug().Msgf("Update stat window of Endport %s:%s", ed.IP, ed.Port)
}

func (dp *Dispatcher) delNode(event source.Event) {
	dp.Lock()
	defer dp.Unlock()
	delete(dp.candidateTable, event.Key())
}
