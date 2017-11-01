package pep

import "google.golang.org/grpc/naming"

type staticWatcher struct {
	Addrs []string
	stop  chan bool
	sent  bool
}

func (w *staticWatcher) Next() ([]*naming.Update, error) {
	if w.sent {
		stop := <-w.stop
		if stop {
			return nil, nil
		}
	}
	w.stop = make(chan bool)
	w.sent = true
	u := make([]*naming.Update, len(w.Addrs))
	for i, a := range w.Addrs {
		u[i] = &naming.Update{Op: naming.Add, Addr: a}
	}
	return u, nil
}

func (w *staticWatcher) Close() {
	w.stop <- true
}
