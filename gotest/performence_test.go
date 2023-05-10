package gotest

import (
	"sync"
	"testing"
	"time"
)

type RW interface {
	Write()
	Read()
}

const cost = time.Microsecond

type Lock struct {
	count int
	mu    sync.Mutex
}

func (l *Lock) Write() {
	l.mu.Lock()
	l.count++
	time.Sleep(cost)
	l.mu.Unlock()
}

func (l *Lock) Read() {
	l.mu.Lock()
	time.Sleep(cost)
	_ = l.count
	l.mu.Unlock()
}

type RWLock struct {
	count int
	mu    sync.RWMutex
}

func (l *RWLock) Write() {
	l.mu.Lock()
	l.count++
	time.Sleep(cost)
	l.mu.Unlock()
}

func (l *RWLock) Read() {
	l.mu.RLock()
	_ = l.count
	time.Sleep(cost)
	l.mu.RUnlock()
}

func benchmark(b *testing.B, rw RW, read, write int) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for k := 0; k < read*100; k++ {
			wg.Add(1)
			go func() {
				rw.Read()
				wg.Done()
			}()
		}
		for k := 0; k < write*100; k++ {
			wg.Add(1)
			go func() {
				rw.Write()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkLock(b *testing.B) {

	benchmark(b, &Lock{}, 9, 1)

	benchmark(b, &RWLock{}, 9, 1)

	benchmark(b, &Lock{}, 1, 9)

	benchmark(b, &RWLock{}, 1, 9)

	benchmark(b, &Lock{}, 5, 5)

	benchmark(b, &RWLock{}, 5, 5)

}
