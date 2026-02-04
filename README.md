package handlers

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"gitlab.com/your-username/your-repo-name/internal/store"
)

// UDPServer handles UDP requests for the key-value store.
type UDPServer struct {
	port string
	store *store.Store
	conn *net.UDPConn
}

// NewUDPServer creates a new UDP server instance.
func NewUDPServer(port string, s *store.Store) *UDPServer {
	return &UDPServer{
		port:  port,
		store: s,
	}
}

// Start initializes and starts the UDP server.
func (u *UDPServer) Start(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", ":"+u.port)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}
	
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to start UDP server: %w", err)
	}
	
	u.conn = conn
	log.Printf("Starting UDP server on :%s", u.port)
	
	go func() {
		<-ctx.Done()
		
		if err := u.conn.Close(); err != nil {
			log.Printf("Error closing UDP connection: %v", err)
		}
	}()
	
	buffer := make([]byte, 1024)
	
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("Error reading UDP packet: %v", err)
				continue
			}
		}
		
		command := string(buffer[:n])
		response := u.processCommand(command)
		
		if _, err := conn.WriteToUDP([]byte(response), clientAddr); err != nil {
			log.Printf("Error writing UDP response: %v", err)
		}
	}
}

func (u *UDPServer) processCommand(command string) string {
	command = strings.TrimSpace(command)
	parts := strings.Split(command, ":")
	
	if len(parts) < 2 {
		return "ERROR:Invalid command format"
	}
	
	operation := strings.ToUpper(parts[0])
	
	switch operation {
	case "GET":
		if len(parts) != 2 {
			return "ERROR:GET requires exactly one argument"
		}
		
		key := parts[1]
		value, exists := u.store.Get(key)
		if !exists {
			return "ERROR:Key not found"
		}
		
		return fmt.Sprintf("OK:%s", value)
		
	case "SET":
		if len(parts) != 3 {
			return "ERROR:SET requires exactly two arguments"
		}
		
		key := parts[1]
		value := parts[2]
		u.store.Set(key, value)
		
		return "OK:Value stored"
		
	case "DELETE":
		if len(parts) != 2 {
			return "ERROR:DELETE requires exactly one argument"
		}
		
		key := parts[1]
		deleted := u.store.Delete(key)
		if !deleted {
			return "ERROR:Key not found"
		}
		
		return "OK:Value deleted"
		
	case "KEYS":
		keys := u.store.Keys()
		if len(keys) == 0 {
			return "OK:"
		}
		
		return fmt.Sprintf("OK:%s", strings.Join(keys, ","))
		
	default:
		return "ERROR:Unknown operation"
	}
}
