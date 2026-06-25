package phragmaos

import (
	"sync"
	"testing"
	"time"
)

// newBucket is a helper to create a token_bucket directly for unit tests.
func newBucket(capacity, tokens, refillRate float64) *token_bucket {
	return &token_bucket{
		capacity:       capacity,
		tokens:         tokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// token_bucket.Allow() tests
// ---------------------------------------------------------------------------

func TestAllow_FullBucket_AllowsRequest(t *testing.T) {
	b := newBucket(10, 10, 1)
	result := b.Allow()

	if !result.Allowed {
		t.Fatal("expected request to be allowed on a full bucket")
	}
	if result.Remaining != 9 {
		t.Errorf("expected 9 remaining tokens, got %d", result.Remaining)
	}
	if result.RetryAfter != 0 {
		t.Errorf("expected RetryAfter 0, got %d", result.RetryAfter)
	}
}

func TestAllow_EmptyBucket_DeniesRequest(t *testing.T) {
	b := newBucket(10, 0, 1)
	result := b.Allow()

	if result.Allowed {
		t.Fatal("expected request to be denied on an empty bucket")
	}
	if result.Remaining != 0 {
		t.Errorf("expected 0 remaining tokens, got %d", result.Remaining)
	}
	// With 0 tokens and refillRate=1, RetryAfter should be ceil((1-0)/1) = 1
	if result.RetryAfter != 1 {
		t.Errorf("expected RetryAfter 1, got %d", result.RetryAfter)
	}
}

func TestAllow_ExactlyOneToken_AllowsThenDenies(t *testing.T) {
	b := newBucket(10, 1, 1)

	first := b.Allow()
	if !first.Allowed {
		t.Fatal("expected first request to be allowed")
	}

	second := b.Allow()
	if second.Allowed {
		t.Fatal("expected second request to be denied after consuming the last token")
	}
}

func TestAllow_DrainsBucket_ToZero(t *testing.T) {
	capacity := 5.0
	b := newBucket(capacity, capacity, 1)

	for i := 0; i < int(capacity); i++ {
		r := b.Allow()
		if !r.Allowed {
			t.Fatalf("request %d should have been allowed, bucket had enough tokens", i+1)
		}
	}

	// Next request should be denied
	r := b.Allow()
	if r.Allowed {
		t.Fatal("bucket should be empty; request should be denied")
	}
}

func TestAllow_RetryAfter_ReflectsRefillRate(t *testing.T) {
	// refillRate = 2 tokens/sec, tokens = 0
	// RetryAfter = ceil((1 - 0) / 2) = 1
	b := newBucket(10, 0, 2)
	result := b.Allow()

	if result.Allowed {
		t.Fatal("expected denial on empty bucket")
	}
	if result.RetryAfter != 1 {
		t.Errorf("expected RetryAfter 1 with refillRate 2, got %d", result.RetryAfter)
	}
}

func TestAllow_TokensRefilled_AfterDelay(t *testing.T) {
	// Start with 0 tokens, refill at 5 tokens/sec.
	// After ~300 ms at least 1 token should have been added.
	b := newBucket(10, 0, 5)
	b.lastRefillTime = time.Now().Add(-300 * time.Millisecond)

	result := b.Allow()
	if !result.Allowed {
		t.Error("expected request to be allowed after refill delay")
	}
}

func TestAllow_TokensDoNotExceedCapacity(t *testing.T) {
	// Set lastRefillTime far in the past so many tokens would be added.
	b := newBucket(10, 0, 100)
	b.lastRefillTime = time.Now().Add(-1 * time.Hour)

	result := b.Allow()
	if !result.Allowed {
		t.Fatal("expected request to be allowed")
	}
	// After consuming 1 token, remaining should be capacity-1 = 9, not > 9.
	if result.Remaining > 9 {
		t.Errorf("remaining tokens %d exceed capacity-1 (9)", result.Remaining)
	}
}

func TestAllow_Concurrent_NoDuplicateGrants(t *testing.T) {
	// Bucket has exactly 5 tokens; 20 goroutines race to consume.
	// Exactly 5 should succeed.
	const capacity = 5
	b := newBucket(capacity, capacity, 0) // refillRate=0 so no new tokens arrive

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allowed int
	)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := b.Allow()
			if r.Allowed {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowed != capacity {
		t.Errorf("expected exactly %d allowed requests, got %d", capacity, allowed)
	}
}

// ---------------------------------------------------------------------------
// Store tests
// ---------------------------------------------------------------------------

func TestStore_GetOrCreate_ReturnsSameBucket(t *testing.T) {
	var s Store

	b1 := s.GetOrCreate("client-a")
	b2 := s.GetOrCreate("client-a")

	if b1 != b2 {
		t.Error("expected the same bucket pointer for the same key")
	}
}

func TestStore_GetOrCreate_ReturnsDifferentBucketsForDifferentKeys(t *testing.T) {
	var s Store

	b1 := s.GetOrCreate("client-a")
	b2 := s.GetOrCreate("client-b")

	if b1 == b2 {
		t.Error("expected different bucket pointers for different keys")
	}
}

func TestStore_NewBucket_HasFullTokens(t *testing.T) {
	var s Store
	b := s.GetOrCreate("fresh-client")

	if b.tokens != b.capacity {
		t.Errorf("new bucket should start full: tokens=%v capacity=%v", b.tokens, b.capacity)
	}
}

func TestStore_GetOrCreate_Concurrent_ReturnsSameBucket(t *testing.T) {
	var (
		s       Store
		wg      sync.WaitGroup
		mu      sync.Mutex
		buckets []*token_bucket
	)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b := s.GetOrCreate("shared-key")
			mu.Lock()
			buckets = append(buckets, b)
			mu.Unlock()
		}()
	}
	wg.Wait()

	first := buckets[0]
	for i, b := range buckets {
		if b != first {
			t.Errorf("goroutine %d got a different bucket pointer", i)
		}
	}
}
