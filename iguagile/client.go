package iguagile

import (
	"encoding/binary"
	"fmt"
	"sync"
)

// Client is a middleman between the connection and the room.
type Client struct {
	id     int
	idByte []byte
	conn   Conn
	room   *Room
	send   chan []byte
}

// NewClient is Client constructed.
func NewClient(room *Room, conn Conn) (*Client, error) {
	id, err := room.generator.Generate()
	if err != nil {
		return nil, err
	}

	idByte := make([]byte, 2)
	binary.LittleEndian.PutUint16(idByte, uint16(id))

	client := &Client{
		id:     id,
		idByte: idByte,
		conn:   conn,
		room:   room,
		send:   make(chan []byte),
	}

	return client, nil
}

func (c *Client) readStart() {
	for {
		message, err := c.conn.Read()
		if err != nil {
			c.room.log.Println(err)
			c.room.CloseConnection(c)
			break
		}

		if err = c.room.Receive(c, message); err != nil {
			c.room.log.Println(err)
			c.room.CloseConnection(c)
			break
		}
	}
}

func (c *Client) writeStart() {
	for {
		if err := c.conn.Write(<-c.send); err != nil {
			c.room.log.Println(err)
			c.room.CloseConnection(c)
			break
		}
	}
}

// GetID is getter for id.
func (c *Client) GetID() int {
	return c.id
}

// GetIDByte is getter for idByte.
func (c *Client) GetIDByte() []byte {
	return c.idByte
}

// Send is enqueue outbound messages.
func (c *Client) Send(message []byte) {
	c.send <- message
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// ClientManager manages clients.
type ClientManager struct {
	clients map[int]*Client
	*sync.Mutex
}

// NewClientManager is ClientManager constructed.
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[int]*Client),
		Mutex:   &sync.Mutex{},
	}
}

// Get the client.
func (m *ClientManager) Get(clientID int) (*Client, error) {
	m.Lock()
	client, ok := m.clients[clientID]
	m.Unlock()
	if !ok {
		return nil, fmt.Errorf("client not exists %v", clientID)
	}

	return client, nil
}

// Add the client.
func (m *ClientManager) Add(client *Client) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client.GetID()]; ok {
		return fmt.Errorf("client already exists %v", client.GetID())
	}

	m.clients[client.GetID()] = client
	return nil
}

// Remove the client.
func (m *ClientManager) Remove(clientID int) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[clientID]; !ok {
		return
	}

	delete(m.clients, clientID)
}

// Exist checks the client exists.
func (m *ClientManager) Exist(clientID int) bool {
	m.Lock()
	_, ok := m.clients[clientID]
	m.Unlock()
	return ok
}

// GetClientMap returns all clients.
func (m *ClientManager) GetClientsMap() map[int]*Client {
	return m.clients
}

// Clear all clients.
func (m *ClientManager) Clear() {
	m.Lock()
	m.clients = make(map[int]*Client)
	m.Unlock()
}

// Count clients.
func (m *ClientManager) Count() int {
	m.Lock()
	defer m.Unlock()
	return len(m.clients)
}
