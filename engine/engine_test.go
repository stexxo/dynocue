package engine

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockTask struct {
	iterations    int
	maxIterations int
	once          sync.Once
	wg            sync.WaitGroup
}

func (m *mockTask) Execute(t time.Duration) bool {
	m.iterations++
	if m.iterations == m.maxIterations {
		m.once.Do(func() { m.wg.Done() })
	}
	return m.iterations >= m.maxIterations
}

func TestNewEngine(t *testing.T) {
	e := NewEngine(60)
	e.Start()

	task := &mockTask{maxIterations: 10}
	task.wg.Add(1)
	e.AddTask(task)
	task.wg.Wait()
	assert.Equal(t, 10, task.iterations)
}

func BenchmarkEngine(b *testing.B) {
	e := NewEngine(60)

	for range 100000 {
		go func() {
			task := &mockTask{maxIterations: rand.Intn(100000)}
			task.wg.Add(1)
			e.AddTask(task)
		}()
	}

	e.Start()

	last := time.Now()
	for b.Loop() {
		e.tick(time.Since(last))
	}
}
