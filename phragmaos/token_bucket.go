package phragmaos

import (
	"sync"
	"time"
)

// const maxBucCapacity = 15

type token_bucket struct {
	capacity        float64
	tokens          float64
	refillRate      float64
	lastRefillTime  time.Time 
	mu              sync.Mutex
}


func (t *token_bucket) Allow() bool{
	t.mu.Lock()
	defer t.mu.Unlock()
    
	// Calculate the time elapsed
    timeElapsed := time.Since(t.lastRefillTime).Seconds()

    // Calculated the number of tokens already would be in bucket upto time.
	tokensRefilled := timeElapsed * t.refillRate

    // Capacity check and update
	if tokensRefilled >= t.capacity {
		t.tokens = t.capacity
	}else{
		t.tokens += tokensRefilled
	}
    
	//Updating the time
	t.lastRefillTime = time.Now()

    if t.tokens >= 1 {
		//reduce one token for each request allowance
		t.tokens = t.tokens - 1
		return true
	}
    
	return false
}