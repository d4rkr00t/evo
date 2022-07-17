package ccm_test

import (
	"evo/internal/ccm"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CcmSeq(t *testing.T) {
	var cm = ccm.New(1)
	var mmap sync.Map
	var counter int32 = 0

	for i := 0; i < 10; i++ {
		cm.Add()

		go func(ii int) {
			mmap.Store(ii, atomic.LoadInt32(&counter))
			atomic.AddInt32(&counter, 1)
			cm.Done()
		}(i)
	}

	cm.Wait()
	assert.Equal(t, cm.NumRunning(), 0)

	var val0, _ = mmap.Load(0)
	assert.Equal(t, val0, int32(0))

	var val1, _ = mmap.Load(1)
	assert.Equal(t, val1, int32(1))

	var val9, _ = mmap.Load(9)
	assert.Equal(t, val9, int32(9))
}

func Test_CcmConcurrent(t *testing.T) {
	var concurrency = 3
	var cm = ccm.New(concurrency)
	var mmap sync.Map
	var counter int32 = 0

	for i := 0; i < 10; i++ {
		cm.Add()
		assert.LessOrEqual(t, cm.NumRunning(), concurrency)

		go func(ii int) {
			mmap.Store(ii, atomic.LoadInt32(&counter))
			atomic.AddInt32(&counter, 1)
			cm.Done()
		}(i)
	}

	cm.Wait()
	assert.Equal(t, cm.NumRunning(), 0)
}
