package domain

import (
	"sync/atomic"
	"unsafe"
)

type Endport struct {
	IP          string       `json:"ip"`
	Port        string       `json:"port"`
	ActiveScore float64      `json:"-"`
	StaticScore float64      `json:"-"`
	Stats       *Stat        `json:"-"`
	window      *StateWindow `json:"-"`
}

func NewEndPort(ip, port string) *Endport {
	ed := &Endport{
		IP:   ip,
		Port: port,
	}
	ed.window = NewStatWindow()
	ed.Stats = ed.window.getStat()
	// note: statChan receiver
	go func() {
		for stat := range ed.window.statChan {
			ed.window.appendStat(stat)
			newStat := ed.window.getStat()
			// why we use atomic operation here? Can we use non atomic operation to improve the performance?
			// the go routine is writer. So where are readers
			atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(ed.Stats)), unsafe.Pointer(newStat))

		}
	}()

	return ed
}

// stat Chan sender. Call by dispatcher
func (ed *Endport) UpdateStat(stat *Stat) {
	ed.window.statChan <- stat
}

func (ed *Endport) CalculateScore(ctx *IpConfContext) {
	if ed.Stats != nil {
		ed.ActiveScore = ed.Stats.CalculateActiveScore()
		ed.StaticScore = ed.Stats.CalculateStaticScore()
	}
}
