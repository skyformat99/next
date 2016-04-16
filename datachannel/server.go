package datachannel

import (
	"sync"

	"github.com/chzyer/flow"
	"github.com/chzyer/next/util"
	"gopkg.in/logex.v1"
)

type Server struct {
	flow           *flow.Flow
	delegate       SvrDelegate
	listeners      []*Listener
	mutex          sync.Mutex
	onListenerExit chan struct{}
}

func NewServer(f *flow.Flow, d SvrDelegate) *Server {
	m := &Server{
		flow:           f,
		delegate:       d,
		onListenerExit: make(chan struct{}, 1),
	}
	return m
}

func (m *Server) GetDataChannel() int {
	ports := m.GetAllDataChannel()
	if len(ports) == 0 {
		return -1
	}
	return util.RandChoiseInt(ports)
}

func (m *Server) GetAllDataChannel() []int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	ret := make([]int, 0, len(m.listeners))
	for _, ln := range m.listeners {
		// BUG(chzyer): ln can be nil
		if ln != nil {
			ret = append(ret, ln.GetPort())
		}
	}
	return ret
}

func (m *Server) Start(n int) {
	m.flow.Add(1)
	defer m.flow.DoneAndClose()

	started := 0
loop:
	for !m.flow.IsClosed() {
		if started < n {
			m.AddChannelListener()
			started++
			m.delegate.OnDChanUpdate(m.GetAllDataChannel())
		} else {
			select {
			case <-m.flow.IsClose():
				break loop
			case <-m.onListenerExit:
				started--
			}
		}
	}
}

func (m *Server) removeListener(idx int) {
	m.mutex.Lock()
	m.listeners[idx] = nil
	m.mutex.Unlock()
	select {
	case m.onListenerExit <- struct{}{}:
	default:
	}
}

func (m *Server) findNewSlot() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for idx, ln := range m.listeners {
		if ln == nil {
			return idx
		}
	}
	m.listeners = append(m.listeners, nil)
	return len(m.listeners) - 1
}

func (m *Server) AddChannelListener() error {
	idx := m.findNewSlot()
	ln, err := NewListener(m.flow, m.delegate, func() {
		m.removeListener(idx)
	})
	if err != nil {
		return logex.Trace(err)
	}

	m.mutex.Lock()
	m.listeners[idx] = ln
	m.mutex.Unlock()

	go ln.Serve()
	return nil
}
