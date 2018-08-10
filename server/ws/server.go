package ws

import (
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"gamelib/server"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 2048,
	}
	deadline = time.Duration(30) * time.Second
)

type Server struct {
	id      uint64
	mux     sync.Mutex
	handler server.Handler
	addr    string
	maxConn int
	quit    chan bool
	config  *server.Config
	conns   map[uint64]*Conn
}

func NewServer(config *server.Config) server.GateServer {

	return &Server{
		config:  config,
		addr:    config.Addr,
		maxConn: config.MaxConn,
		quit:    make(chan bool),
		conns:   make(map[uint64]*Conn),
	}
}

func (server *Server) SetHandler(handler server.Handler) {

	server.handler = handler
}

func (server *Server) SetMaxConn(n int) {

	server.maxConn = n
}

func (server *Server) Start() {

	go func() {

		log.Println("websocket listening on", server.addr)

		http.HandleFunc("/", server.serveWs)
		http.HandleFunc("/ws", server.serveWs)

		err := http.ListenAndServe(server.addr, nil)
		if err != nil {

			panic(err.Error())
		}
	}()

	<-server.quit
}

func (server *Server) Close() {

	log.Println("websocket closing")

	close(server.quit)

	server.mux.Lock()

	conns := make(map[uint64]*Conn)
	for i, c := range server.conns {

		conns[i] = c
	}

	server.mux.Unlock()

	for _, c := range conns {

		c.Close()
	}
}

func (server *Server) Count() int {

	server.mux.Lock()
	defer server.mux.Unlock()

	return len(server.conns)
}

func (server *Server) removeConn(id uint64) {

	server.mux.Lock()
	defer server.mux.Unlock()

	if conn, ok := server.conns[id]; ok {

		if server.handler != nil {

			server.handler.Close(conn)
		}

		delete(server.conns, id)
	}
}

func (server *Server) serveWs(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {

		http.Error(w, "Method not allowed", 405)

		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {

		origin := r.Header["Origin"]
		if len(origin) == 0 {

			return true
		}

		u, err := url.Parse(origin[0])
		if err != nil {

			return false
		}

		if len(server.config.OriginAllow) == 0 {

			return true
		}

		return u.Host == server.config.OriginAllow
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {

		log.Println(err)

		return
	}

	server.mux.Lock()

	id := server.id
	server.id++
	conn := newConn(id, ws, server.removeConn)
	server.conns[id] = conn

	server.mux.Unlock()

	if server.handler != nil {

		server.handler.Open(conn)
	}
}
