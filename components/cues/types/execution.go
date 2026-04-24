package types

import "sync"

type ExecutionModel struct {
	mu sync.RWMutex
}
