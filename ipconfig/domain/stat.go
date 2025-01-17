package domain

// computing logic for sorting
import (
	"math"
)

// the available resources of endpoint
type Stat struct {
	ConnNum      float64
	MessageBytes float64
}

// what is active score
func (stat *Stat) CalculateActiveScore() float64 {
	return getGB(stat.MessageBytes)
}

func (stat *Stat) CalculateStaticScore() float64 {
	return stat.ConnNum
}

func (stat *Stat) Avg(num float64) {
	stat.ConnNum /= num
	stat.MessageBytes /= num
}

func (stat *Stat) Clone() *Stat {
	return &Stat{
		ConnNum:      stat.ConnNum,
		MessageBytes: stat.MessageBytes,
	}
}

func (stat *Stat) Add(st *Stat) {
	if st == nil {
		return
	}
	stat.ConnNum += st.ConnNum
	stat.MessageBytes += st.MessageBytes
}

func (stat *Stat) Sub(st *Stat) {
	if st == nil {
		return
	}
	stat.ConnNum -= st.ConnNum
	stat.MessageBytes -= st.MessageBytes
}

func getGB(m float64) float64 {
	return decimal(m / (1 << 30))
}

func decimal(value float64) float64 {
	return math.Trunc(value*1e2+0.5) * 1e-2
}
