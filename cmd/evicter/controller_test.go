package main

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestCanEvict(t *testing.T) {
	now := int(time.Now().Unix())
	specs := map[string]struct {
		srcAnn    map[string]string
		expResult bool
	}{
		"with timestamp after incubation period": {
			srcAnn: map[string]string{
				"k-rail/tainted-timestamp": strconv.Itoa(now - 1),
				"k-rail/tainted-reason":    "test",
			},
			expResult: true,
		},
		"with timestamp in incubation period": {
			srcAnn: map[string]string{
				"k-rail/tainted-timestamp": strconv.Itoa(now),
				"k-rail/tainted-reason":    "test",
			},
			expResult: false,
		},
		"without timestamp annotation": {
			srcAnn: map[string]string{
				"k-rail/tainted-reason": "test",
			},
			expResult: true,
		},
		"with timestamp containing non timestamp string": {
			srcAnn: map[string]string{
				"k-rail/tainted-timestamp": "",
				"k-rail/tainted-reason":    "test",
			},
			expResult: true,
		},
		"with preventEviction annotation": {
			srcAnn: map[string]string{
				"k-rail/tainted-timestamp":        strconv.Itoa(now - 1),
				"k-rail/tainted-reason":           "test",
				"k-rail/tainted-prevent-eviction": "true",
			},
			expResult: false,
		},
		"with preventEviction annotation - uppercase": {
			srcAnn: map[string]string{
				"k-rail/tainted-timestamp":        strconv.Itoa(now - 1),
				"k-rail/tainted-reason":           "test",
				"k-rail/tainted-prevent-eviction": "TRUE",
			},
			expResult: false,
		},
		"with preventEviction annotation - yes": {
			srcAnn: map[string]string{
				"k-rail/tainted-timestamp":        strconv.Itoa(now - 1),
				"k-rail/tainted-reason":           "test",
				"k-rail/tainted-prevent-eviction": "yes",
			},
			expResult: false,
		},
		"with preventEviction annotation - non bool": {
			srcAnn: map[string]string{
				"k-rail/tainted-timestamp":        strconv.Itoa(now - 1),
				"k-rail/tainted-reason":           "test",
				"k-rail/tainted-prevent-eviction": "",
			},
			expResult: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			pod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: spec.srcAnn}}
			if got := canEvict(&pod, time.Second); spec.expResult != got {
				t.Errorf("expected %v but got %v", spec.expResult, got)
			}
		})
	}
}

func TestEvictPod(t *testing.T) {
	now := int(time.Now().Unix())

	specs := map[string]struct {
		srcAnn        map[string]string
		expReason     string
		expMsg        string
		expNoEviction bool
	}{
		"evicted with custom reason": {
			srcAnn: map[string]string{
				"k-rail/tainted-timestamp": strconv.Itoa(now - 1),
				"k-rail/tainted-reason":    "test",
			},
			expReason: "Tainted",
			expMsg:    "test",
		},
		"evicted with default reason": {
			srcAnn: map[string]string{
				"k-rail/tainted-timestamp": strconv.Itoa(now - 1),
			},
			expReason: "Tainted",
			expMsg:    noEvictionNote,
		},
		"not evicted with annotation": {
			srcAnn: map[string]string{
				"k-rail/tainted-timestamp":        strconv.Itoa(now - 1),
				"k-rail/tainted-prevent-eviction": "yes",
			},
			expNoEviction: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			store := cache.NewStore(cache.MetaNamespaceKeyFunc)
			pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "myPod", Annotations: spec.srcAnn}}
			store.Add(pod)
			prov := &recordingPodProvisioner{}
			c := NewController(nil, store, nil, prov, 1)
			// when
			err := c.evictPod("myPod")
			// then
			if err != nil {
				t.Fatalf("unexpected error: %+v", err)
			}
			if spec.expNoEviction {
				if prov.evictedPods != nil {
					t.Fatalf("expected no call but got %v", prov.evictedPods)
				}
				return
			}
			// there should be 1 call
			if exp, got := []*v1.Pod{pod}, prov.evictedPods; !reflect.DeepEqual(exp, got) {
				t.Errorf("expected %v but got %v", exp, got)
			}
			if exp, got := []string{spec.expReason}, prov.reasons; !reflect.DeepEqual(exp, got) {
				t.Errorf("expected %v but got %v", exp, got)
			}
			if exp, got := []string{spec.expMsg}, prov.msgs; !reflect.DeepEqual(exp, got) {
				t.Errorf("expected %v but got %v", exp, got)
			}
		})
	}
}

type recordingPodProvisioner struct {
	evictedPods []*v1.Pod
	reasons     []string
	msgs        []string
	result      error
}

func (r *recordingPodProvisioner) Evict(pod *v1.Pod, reason, msg string) error {
	r.evictedPods = append(r.evictedPods, pod)
	r.reasons = append(r.reasons, reason)
	r.msgs = append(r.msgs, msg)
	return r.result
}
