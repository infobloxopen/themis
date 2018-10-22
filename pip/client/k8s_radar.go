package client

import (
	"sync"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type k8sRadar struct {
	sync.Mutex
	started bool
	done    chan struct{}

	name k8sName
	port string
	d    time.Duration

	sii cache.SharedIndexInformer
}

func newK8sRadar(addr string, ki kubernetes.Interface, d time.Duration) (*k8sRadar, error) {
	h, p, err := splitHostPort(addr)
	if err != nil {
		return nil, err
	}

	n, err := makeK8sName(h)
	if err != nil {
		return nil, err
	}

	factory := informers.NewSharedInformerFactoryWithOptions(ki, d,
		informers.WithNamespace(n.namespace),
	)

	return &k8sRadar{
		done: make(chan struct{}),
		name: n,
		port: p,
		d:    d,
		sii:  factory.Core().V1().Pods().Informer(),
	}, nil
}

func (r *k8sRadar) start(addrs []string) <-chan addrUpdate {
	r.Lock()
	defer r.Unlock()

	if r.started || r.done == nil {
		return nil
	}
	r.started = true

	done := r.done
	ch := make(chan addrUpdate)
	sii := r.sii
	sii.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    r.OnAdd(ch),
			UpdateFunc: r.OnUpdate(ch),
			DeleteFunc: r.OnDelete(ch),
		},
	)

	go func(sii cache.SharedIndexInformer, ch chan addrUpdate) {
		defer close(ch)

		sii.Run(done)
	}(sii, ch)

	return ch
}

func (r *k8sRadar) stop() {
	r.Lock()
	defer r.Unlock()

	if !r.started || r.done == nil {
		return
	}
	r.started = false

	close(r.done)
	r.done = nil
}

func (r *k8sRadar) OnAdd(ch chan addrUpdate) func(obj interface{}) {
	port := r.port
	s := r.name.selector

	return func(obj interface{}) {
		if pod, ok := obj.(*core.Pod); ok && matchK8sPod(pod, s) && isK8sPodReady(pod) {
			ch <- addrUpdate{
				op:   addrUpdateOpAdd,
				addr: joinAddrPort(pod.Status.PodIP, port),
			}
		}
	}
}

func (r *k8sRadar) OnUpdate(ch chan addrUpdate) func(oldObj, newObj interface{}) {
	port := r.port
	s := r.name.selector

	return func(oldObj, newObj interface{}) {
		if pod, ok := newObj.(*core.Pod); ok && matchK8sPod(pod, s) {
			if isK8sPodReady(pod) {
				ch <- addrUpdate{
					op:   addrUpdateOpAdd,
					addr: joinAddrPort(pod.Status.PodIP, port),
				}
			} else {
				ch <- addrUpdate{
					op:   addrUpdateOpDel,
					addr: joinAddrPort(pod.Status.PodIP, port),
				}
			}
		} else if pod, ok := oldObj.(*core.Pod); ok && matchK8sPod(pod, s) {
			ch <- addrUpdate{
				op:   addrUpdateOpDel,
				addr: joinAddrPort(pod.Status.PodIP, port),
			}
		}
	}
}

func (r *k8sRadar) OnDelete(ch chan addrUpdate) func(obj interface{}) {
	port := r.port
	s := r.name.selector

	return func(obj interface{}) {
		if pod, ok := obj.(*core.Pod); ok && matchK8sPod(pod, s) {
			ch <- addrUpdate{
				op:   addrUpdateOpDel,
				addr: joinAddrPort(pod.Status.PodIP, port),
			}
		}
	}
}
