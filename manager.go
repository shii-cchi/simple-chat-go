package main

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:8080":
		return true
	default:
		return false
	}
}

type Manager struct {
	clients ClientList

	sync.RWMutex

	handlers map[string]EventHandler
}

func NewManager() *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
	}

	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn, m)
	m.addClient(client)

	go client.readMessages()
	go client.writeMessages()
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}
