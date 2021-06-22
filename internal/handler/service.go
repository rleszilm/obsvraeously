package handler

import (
	"context"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rleszilm/genms/log"
	"github.com/rleszilm/genms/service"
	rest_service "github.com/rleszilm/genms/service/rest"
	"github.com/rleszilm/obsvraeously/internal/avrae"
)

var (
	logs = log.NewChannel("handler")
)

type Service struct {
	service.Dependencies

	mux      sync.Mutex
	avdb     *avrae.Service
	server   *rest_service.Server
	upgrader *websocket.Upgrader
	streams  map[chan *avrae.Roll]struct{}
	avatars  map[string]string
}

func NewService(avdb *avrae.Service, server *rest_service.Server) *Service {
	s := &Service{
		avdb:    avdb,
		server:  server,
		streams: map[chan *avrae.Roll]struct{}{},
		avatars: map[string]string{},
	}
	s.WithDependencies(avdb)

	s.server.WithRouteFunc("/ws", s.socket)
	s.server.WithDependencies(s)

	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	s.upgrader = upgrader

	return s
}

func (s *Service) Initialize(ctx context.Context) error {
	go s.rollWorker()
	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Service) NameOf() string {
	return "handler"
}

func (s *Service) String() string {
	return s.NameOf()
}

func (s *Service) socket(resp http.ResponseWriter, req *http.Request) {
	ws, err := s.upgrader.Upgrade(resp, req, nil)
	if err != nil {
		logs.Error(err)
		return
	}
	defer ws.Close()

	ch := s.addstream()
	defer s.delstream(ch)

	for roll := range ch {
		if err := ws.WriteJSON(roll); err != nil {
			logs.Error(err)
			return
		}
	}
}

func (s *Service) addstream() chan *avrae.Roll {
	s.mux.Lock()
	defer s.mux.Unlock()

	ch := make(chan *avrae.Roll)
	s.streams[ch] = struct{}{}

	return ch
}

func (s *Service) delstream(ch chan *avrae.Roll) {
	s.mux.Lock()
	defer s.mux.Unlock()

	delete(s.streams, ch)
	close(ch)
}

func (s *Service) rollWorker() {
	for roll := range s.avdb.Rolls() {
		s.handleRoll(roll)
	}
}

func (s *Service) handleRoll(r *avrae.Roll) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if r.Avatar != "" {
		s.avatars[r.Player] = r.Avatar
	} else {
		r.Avatar = s.avatars[r.Player]
	}

	for ch := range s.streams {
		ch <- r
	}
}
