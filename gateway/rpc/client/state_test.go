package client

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// using ctx as a done chan
func someLongRunningOperation(ctx context.Context) (string, error) {
	// Simulate a long-running operation
	select {
	case <-time.After(3 * time.Second): // Simulate work taking 3 seconds
		return "Operation completed", nil
	case <-ctx.Done(): // Handle context cancellation or timeout
		return "", ctx.Err()
	}
}

func TestContext(t *testing.T) {
	// Create a context with a 2-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Perform the operation
	result, err := someLongRunningOperation(ctx)
	if err != nil {
		fmt.Println("Operation failed:", err)
		return
	}

	fmt.Println("Operation succeeded:", result)
}
