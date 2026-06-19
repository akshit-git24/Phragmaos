package phragmaos

import (
	"sync"
	"time"
)

type Store struct {
	buckets sync.Map  // key is string of IP/API and value is the pointer of the bucket in memory
}


func (s *Store) GetOrCreate(inp string) *token_bucket {

	newBucket := &token_bucket{
		capacity:  15,
		tokens:  15,
		refillRate: 1,
		lastRefillTime: time.Now(),
	}

	actual, _ := s.buckets.LoadOrStore(inp, newBucket)
	return actual.(*token_bucket)

}