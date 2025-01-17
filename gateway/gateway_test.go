package gateway

import (
	"fmt"
	"testing"
	"time"
)

func TestSelect(t *testing.T) {
	eventChan := make(chan struct{})

	go func() {
		time.Sleep(5 * time.Second)
		eventChan <- struct{}{}
	}()

	for {
		select {
		case <-eventChan:
			return
		default:
			fmt.Printf("failed to epoll wait")
		}
	}
}
