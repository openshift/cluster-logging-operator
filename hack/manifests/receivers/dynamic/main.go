package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const ConfigPath = "/config/config.json"

type Config struct {
	Listen bool `json:"listen"`
}

type Server struct {
	listener      net.Listener
	stopChan      chan struct{}
	wg            sync.WaitGroup
	activeClients []net.Conn
	mu            sync.Mutex
}

func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		return err
	}

	s.stopChan = make(chan struct{})
	s.activeClients = make([]net.Conn, 0)

	s.wg.Add(1)
	go s.acceptLoop()

	fmt.Println("[SERVER] TCP Server listening on port 8080...")
	return nil
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		s.listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				return
			default:
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return
			}
		}

		fmt.Printf("[SERVER] Connection accepted from %s\n", conn.RemoteAddr())

		s.mu.Lock()
		s.activeClients = append(s.activeClients, conn)
		s.mu.Unlock()

		s.wg.Add(1)
		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	defer s.removeClient(conn)

	reader := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		req, err := http.ReadRequest(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("[SERVER] Failed to parse HTTP request: %v\n", err)
			}
			return
		}

		body, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			fmt.Printf("[SERVER] Failed to read request body: %v\n", err)
			s.sendResponseWithConnection(conn, http.StatusBadRequest, "Bad Request", false)
			return
		}

		if len(body) > 0 {
			maxLen := 100
			if len(body) < maxLen {
				maxLen = len(body)
			}
			fmt.Printf("[STDOUT] %s\n", string(body[:maxLen]))
		}

		keepAlive := shouldKeepAlive(req)
		s.sendResponseWithConnection(conn, http.StatusOK, "OK", keepAlive)

		if !keepAlive {
			return
		}

		conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	}
}

func (s *Server) sendResponse(conn net.Conn, statusCode int, statusText string) {
	s.sendResponseWithConnection(conn, statusCode, statusText, false)
}

func (s *Server) sendResponseWithConnection(conn net.Conn, statusCode int, statusText string, keepAlive bool) {
	connectionHeader := "close"
	if keepAlive {
		connectionHeader = "keep-alive"
	}
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: %s\r\n\r\n", statusCode, statusText, connectionHeader)
	conn.Write([]byte(response))
}

func shouldKeepAlive(req *http.Request) bool {
	if req.ProtoMajor < 1 || (req.ProtoMajor == 1 && req.ProtoMinor < 1) {
		return false
	}

	connectionHeader := req.Header.Get("Connection")
	if connectionHeader == "close" {
		return false
	}

	return true
}

func (s *Server) removeClient(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.activeClients {
		if c == conn {
			s.activeClients = append(s.activeClients[:i], s.activeClients[i+1:]...)
			break
		}
	}
}

func (s *Server) Stop() {
	fmt.Println("[SERVER] Initiating silent black-hole teardown...")

	close(s.stopChan)

	s.mu.Lock()
	for _, conn := range s.activeClients {
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetLinger(0)
		}
	}
	s.activeClients = nil
	s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
	}

	s.wg.Wait()

	fmt.Println("[SERVER] Server ghosted active clients and shut down listener.")
}

func readConfig() (bool, error) {
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		return false, nil
	}

	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return false, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return false, err
	}

	return config.Listen, nil
}

func main() {
	var server *Server
	var lastState *bool

	fmt.Println("[MAIN] Application loop online. Watching config changes...")

	for {
		currentListen, err := readConfig()
		if err != nil {
			fmt.Printf("[MAIN] Error checking config: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if lastState == nil || currentListen != *lastState {
			fmt.Printf("[MAIN] State change detected! listen = %t\n", currentListen)

			if currentListen {
				server = &Server{}
				if err := server.Start(); err != nil {
					fmt.Printf("[SERVER] Failed to start server: %v\n", err)
				}
			} else {
				if server != nil {
					server.Stop()
					server = nil
				}
			}

			lastState = &currentListen
		}

		time.Sleep(2 * time.Second)
	}
}
