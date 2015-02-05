package tcp_server

import (
	"bufio"
	"log"
	"net"
)

// Client holds info about connection
type Client struct {
	conn     net.Conn
	Server   *Server
	incoming chan string // Channel for incoming data from client
}

// TCP server
type Server struct {
	clients                  []*Client
	address                  string        // Address to open connection: localhost:9999
	joins                    chan net.Conn // Channel for new connections
	onNewClientCallback      func(c *Client)
	onClientConnectionClosed func(c *Client, err error)
	onNewMessage             func(c *Client, message string)
}

func (c *Client) Addr() string {
	return c.conn.RemoteAddr().String()
}

// Read client data from channel
func (c *Client) listen() {
	reader := bufio.NewReader(c.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			c.conn.Close()
			c.Server.onClientConnectionClosed(c, err)
			return
		}
		c.Server.onNewMessage(c, message)
	}
}

// Called right after server starts listening new client
func (s *Server) OnNewClient(callback func(c *Client)) {
	s.onNewClientCallback = callback
}

// Called right after connection closed
func (s *Server) OnClientConnectionClosed(callback func(c *Client, err error)) {
	s.onClientConnectionClosed = callback
}

// Called when Client receives new message
func (s *Server) OnNewMessage(callback func(c *Client, message string)) {
	s.onNewMessage = callback
}

// Creates new Client instance and starts listening
func (s *Server) newClient(conn net.Conn) {
	client := &Client{
		conn:   conn,
		Server: s,
	}
	go client.listen()
	s.onNewClientCallback(client)
}

// Listens new connections channel and creating new client
func (s *Server) listenChannels() {
	for {
		select {
		case conn := <-s.joins:
			s.newClient(conn)
		}
	}
}

// Start network server
func (s *Server) Listen() {
	go s.listenChannels()

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal("Error starting TCP server.")
	}
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		s.joins <- conn
	}
}

func (s *Server) Addr() string {
	return s.address
}

// Creates new tcp server instance
func New(address string) *Server {
	log.Print("Creating server with address " + address)
	server := &Server{
		address: address,
		joins:   make(chan net.Conn),
	}

	server.OnNewClient(func(c *Client) {})
	server.OnNewMessage(func(c *Client, message string) {})
	server.OnClientConnectionClosed(func(c *Client, err error) {})

	return server
}
