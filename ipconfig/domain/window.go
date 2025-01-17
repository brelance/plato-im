package domain

const windowSize = 5

type StateWindow struct {
	// maintain all states in the window
	stateQueue []*Stat
	sumStat    *Stat
	// the sender is dispatcher
	statChan chan *Stat
	idx      int64
}

func NewStatWindow() *StateWindow {
	return &StateWindow{
		stateQueue: make([]*Stat, windowSize),
		statChan:   make(chan *Stat),
		sumStat:    &Stat{},
	}
}

// Return a stat including the sum of all stat info in the window
func (sw *StateWindow) getStat() *Stat {
	// why we should cloen here?
	ns := sw.sumStat.Clone()
	ns.Avg(windowSize)
	return ns
}

func (sw *StateWindow) appendStat(s *Stat) {
	sw.sumStat.Sub(sw.stateQueue[sw.idx%windowSize])
	sw.stateQueue[sw.idx%windowSize] = s
	sw.sumStat.Add(s)
	sw.idx++
}
