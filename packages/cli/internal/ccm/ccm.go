package ccm

import "sync"

type ConcurrencyManager struct {
	wg      sync.WaitGroup
	queueCh chan struct{}
}

func New(concurrency int) ConcurrencyManager {
	return ConcurrencyManager{
		wg:      sync.WaitGroup{},
		queueCh: make(chan struct{}, concurrency),
	}
}

func (cm *ConcurrencyManager) Add() {
	cm.wg.Add(1)
	cm.queueCh <- struct{}{}
}

func (cm *ConcurrencyManager) Done() {
	cm.wg.Done()
	<-cm.queueCh
}

func (cm *ConcurrencyManager) Wait() {
	cm.wg.Wait()
}

func (cm *ConcurrencyManager) NumRunning() int {
	return len(cm.queueCh)
}
