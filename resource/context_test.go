package resource

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestResourceCacheContext(t *testing.T) {
	const testRunCount = 2
	const testCacheKey = 999
	specs := map[string]struct {
		srcCtx         func() context.Context
		srcKey         int
		srcFactoryFunc *creatorMock
		expResp        interface{}
		expCalls       int
	}{
		"with cache in ctx ": {
			srcCtx: func() context.Context {
				return WithResourceCache(context.TODO())
			},
			srcKey:         testCacheKey,
			srcFactoryFunc: &creatorMock{respValue: "myValue"},
			expCalls:       1,
			expResp:        "myValue",
		},
		"with cache filled": {
			srcCtx: func() context.Context {
				ctx := WithResourceCache(context.TODO())
				GetResourceCache(ctx).getOrSet(testCacheKey, func() interface{} {
					return "myValue"
				})
				return ctx
			},
			srcKey:         testCacheKey,
			srcFactoryFunc: &creatorMock{respValue: "otherValue"},
			expCalls:       0,
			expResp:        "myValue",
		},
		"with empty ctx": {
			srcCtx:         context.TODO,
			srcKey:         testCacheKey,
			srcFactoryFunc: &creatorMock{respValue: "foo"},
			expCalls:       testRunCount,
			expResp:        "foo",
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			ctx := spec.srcCtx()
			mock := spec.srcFactoryFunc
			for i := 0; i < testRunCount; i++ {
				resp := GetResourceCache(ctx).getOrSet(spec.srcKey, mock.CountCall)
				if exp, got := spec.expResp, resp; !reflect.DeepEqual(exp, got) {
					t.Errorf("expected %v but got %v", exp, got)
				}
			}
			if exp, got := spec.expCalls, mock.called; exp != got {
				t.Errorf("expected %d but got %d", exp, got)
			}
		})
	}
}

func TestCacheWithConcurrentAccess(t *testing.T) {
	const testCacheKey = 999
	const actorCount = 10

	var awaitStart sync.WaitGroup
	awaitStart.Add(actorCount)
	var awaitCompleted sync.WaitGroup
	awaitCompleted.Add(actorCount)

	actors := make([]*creatorMock, actorCount)
	rsp := make(chan interface{}, actorCount)

	c := &cache{m: make(map[int]interface{}, 1)}
	for i := 0; i < actorCount; i++ {
		actors[i] = &creatorMock{respValue: i}
		go func(i int) {
			awaitStart.Done()
			awaitStart.Wait() // wait for all actors to start sync
			rsp <- c.getOrSet(testCacheKey, actors[i].CountCall)
			awaitCompleted.Done()
		}(i)
	}
	awaitCompleted.Wait()

	// then only 1 create function should be called
	var active *creatorMock
	var expResult int
	for i, a := range actors {
		if a.called != 0 {
			if active != nil {
				t.Fatal("more than 1 create function called")
			}
			active = a
			expResult = i
		}
	}
	// and all should see the same result
	for i := 0; i < actorCount; i++ {
		select {
		case r := <-rsp:
			if exp, got := expResult, r; exp != got {
				t.Errorf("expected %v but got %v", exp, got)
			}
		case <-time.After(time.Millisecond):
			t.Fatal("test timeout")
		}
	}
}

type creatorMock struct {
	l         sync.Mutex
	called    int
	respValue interface{}
}

func (m *creatorMock) CountCall() interface{} {
	m.l.Lock()
	m.called++
	m.l.Unlock()
	return m.respValue
}
