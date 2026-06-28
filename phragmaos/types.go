package phragmaos

import (
	"sync"
	"time"
)


type token_bucket struct {
	capacity        float64
	tokens          float64
	refillRate      float64
	lastRefillTime  time.Time 
	mu              sync.Mutex
}

type Result struct {
    Allowed bool
	Remaining int
    RetryAfter int
}


type Store struct {
	buckets sync.Map  // key is string of IP/API and value is the pointer of the bucket in memory
}

//config file structs
type EndpointConfig struct {
    Path       string  `json:"path"`
    Limit      float64 `json:"limit"`
    RefillRate float64 `json:"refillRate"`
}

type Config struct {
    Endpoints []EndpointConfig `json:"endpoints"`
}