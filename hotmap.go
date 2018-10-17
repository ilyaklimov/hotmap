package hotmap

import (
	"sync"
	"time"
	"context"
)

const durationDefault = time.Duration(30 * time.Second)

type Hotmap struct {
	mu sync.RWMutex

	cache map[string]string
	duration time.Duration
	cleaners map[string](chan struct{})
}

func (hm *Hotmap) get(key string) (string, bool) {
	v, ok := hm.cache[key]
	if ok {
		hm.stopCleaner(key)
		hm.delete(key)
	}
	return v, ok
}

func (hm *Hotmap) set(key, value string) {
	hm.cache[key] = value
}

func (hm *Hotmap) delete(key string) {
	hm.deleteCleaner(key)
	delete(hm.cache, key)
}

func (hm *Hotmap) stopCleaner(key string) {
	hm.cleaners[key] <- struct{}{}
}

func (hm *Hotmap) deleteCleaner(key string) {
	close(hm.cleaners[key])
	delete(hm.cleaners, key)
}

func (hm *Hotmap) cleaner(ctx context.Context, key string, stop <-chan struct{}) {
	select {
	case <- stop:
		return
	case <- ctx.Done():
		hm.Delete(key)
		return
	}
}

func (hm *Hotmap) Get(key string) (string, bool) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	return hm.get(key)
}

func (hm *Hotmap) Set(key, value string) {
	hm.mu.Lock()
	if _, ok := hm.cache[key]; ok {
		hm.delete(key)
	}
	hm.set(key, value)
	hm.cleaners[key] = make(chan (struct{}))
	stop := hm.cleaners[key]
	hm.mu.Unlock()

	ctx, _ := context.WithTimeout(context.Background(), hm.duration)
	go hm.cleaner(ctx, key, stop)
}

func (hm *Hotmap) Delete(key string) {
	hm.mu.Lock()
	hm.delete(key)
	hm.mu.Unlock()
}

func (hm *Hotmap) Len() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return len(hm.cache)
}

func (hm *Hotmap) SetDuration(duration time.Duration) {
	hm.duration = duration
}

func (hm *Hotmap) Close() {
	if len(hm.cleaners) != 0 {
		for key := range hm.cleaners {
			hm.stopCleaner(key)
			hm.deleteCleaner(key)
		}
	}
	hm.cache = nil
}

func New() *Hotmap {
	return &Hotmap{
		cache: make(map[string]string),
		duration: durationDefault,
		cleaners: make(map[string](chan struct{})),
	}
}