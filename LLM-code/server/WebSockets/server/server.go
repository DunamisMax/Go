package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// hub manages all active clients and broadcasts messages to them
type hub struct {
	mu      sync.Mutex
	clients map[*client]struct{}
	logger  *log.Logger
}

func newHub(logger *log.Logger) *hub {
	return &hub{
		clients: make(map[*client]struct{}),
		logger:  logger,
	}
}

func (h *hub) register(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
	h.logger.Printf("[INFO] User '%s' connected.", c.username)
}

func (h *hub) unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	h.logger.Printf("[INFO] User '%s' disconnected.", c.username)
}

func (h *hub) broadcast(sender *client, msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	message := fmt.Sprintf("%s: %s", sender.username, string(msg))
	h.logger.Printf("[MESSAGE] %s", message)
	for c := range h.clients {
		c.send([]byte(message))
	}
}

type client struct {
	conn     io.ReadWriteCloser
	username string
	mu       sync.Mutex
}

func (c *client) send(msg []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return writeWebSocketFrame(c.conn, 0x1, msg)
}

var globalHub *hub

func main() {
	// Set up logging to file and stdout
	logFile, err := os.OpenFile("activity.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open activity.log: %v", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(multiWriter, "", log.LstdFlags)

	globalHub = newHub(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		WebSocketHandler(w, r, logger)
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      LoggingMiddleware(mux, logger),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logger.Println("[INFO] Server starting on :8080")

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("ListenAndServe error: %v", err)
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request, logger *log.Logger) {
	if !isWebSocketUpgrade(r) {
		http.Error(w, "Not a WebSocket handshake", http.StatusBadRequest)
		return
	}

	rc := http.NewResponseController(w)
	conn, brw, err := rc.Hijack()
	if err != nil {
		logger.Printf("Hijack error: %v", err)
		return
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	acceptKey := computeAcceptKey(key)

	resp := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Accept: %s\r\n\r\n", acceptKey)
	if _, err := io.WriteString(conn, resp); err != nil {
		logger.Printf("Error writing handshake response: %v", err)
		conn.Close()
		return
	}

	// The first message from the client should be their username
	opcode, payload, err := readWebSocketFrame(brw)
	if err != nil {
		logger.Printf("Error reading username frame: %v", err)
		conn.Close()
		return
	}
	if opcode == 0x8 {
		// Client immediately closed?
		conn.Close()
		return
	}
	username := strings.TrimSpace(string(payload))
	if username == "" {
		username = "Anonymous"
	}

	c := &client{conn: conn, username: html.EscapeString(username)}
	globalHub.register(c)
	defer func() {
		globalHub.unregister(c)
		conn.Close()
	}()

	// Now read messages in a loop and broadcast them
	for {
		opcode, payload, err := readWebSocketFrame(brw)
		if err != nil {
			logger.Printf("Read frame error: %v", err)
			return
		}
		if opcode == 0x8 {
			// Close frame
			writeWebSocketFrame(conn, 0x8, []byte{})
			return
		}
		if opcode == 0x1 {
			// Text frame: broadcast
			globalHub.broadcast(c, payload)
		}
	}
}

func isWebSocketUpgrade(r *http.Request) bool {
	if !strings.EqualFold(r.Header.Get("Connection"), "Upgrade") {
		return false
	}
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return false
	}
	if r.Header.Get("Sec-WebSocket-Key") == "" {
		return false
	}
	if !strings.Contains(r.Header.Get("Sec-WebSocket-Version"), "13") {
		return false
	}
	return true
}

func computeAcceptKey(key string) string {
	const magicGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + magicGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func readWebSocketFrame(brw *bufio.ReadWriter) (byte, []byte, error) {
	if err := brw.Flush(); err != nil {
		return 0, nil, err
	}

	header := make([]byte, 2)
	if _, err := io.ReadFull(brw, header); err != nil {
		return 0, nil, err
	}
	fin := (header[0] & 0x80) != 0
	opcode := header[0] & 0x0f
	if !fin {
		return 0, nil, errors.New("fragmented frames not supported in this example")
	}

	mask := (header[1] & 0x80) != 0
	payloadLen := int64(header[1] & 0x7f)

	switch payloadLen {
	case 126:
		ext := make([]byte, 2)
		if _, err := io.ReadFull(brw, ext); err != nil {
			return 0, nil, err
		}
		payloadLen = int64(uint16(ext[0])<<8 | uint16(ext[1]))
	case 127:
		ext := make([]byte, 8)
		if _, err := io.ReadFull(brw, ext); err != nil {
			return 0, nil, err
		}
		payloadLen = int64((uint64(ext[0])<<56 | uint64(ext[1])<<48 |
			uint64(ext[2])<<40 | uint64(ext[3])<<32 |
			uint64(ext[4])<<24 | uint64(ext[5])<<16 |
			uint64(ext[6])<<8 | uint64(ext[7])))
	}

	var maskKey [4]byte
	if mask {
		if _, err := io.ReadFull(brw, maskKey[:]); err != nil {
			return 0, nil, err
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(brw, payload); err != nil {
		return 0, nil, err
	}

	if mask {
		for i := int64(0); i < payloadLen; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	return opcode, payload, nil
}

func writeWebSocketFrame(w io.Writer, opcode byte, payload []byte) error {
	var header []byte
	payloadLen := len(payload)

	switch {
	case payloadLen <= 125:
		header = []byte{0x80 | opcode, byte(payloadLen)}
	case payloadLen < 65536:
		header = []byte{0x80 | opcode, 126, byte(payloadLen >> 8), byte(payloadLen & 0xff)}
	default:
		header = []byte{0x80 | opcode, 127,
			byte(payloadLen >> 56), byte(payloadLen >> 48),
			byte(payloadLen >> 40), byte(payloadLen >> 32),
			byte(payloadLen >> 24), byte(payloadLen >> 16),
			byte(payloadLen >> 8), byte(payloadLen)}
	}

	if _, err := w.Write(header); err != nil {
		return err
	}
	if _, err := w.Write(payload); err != nil {
		return err
	}
	return nil
}

func LoggingMiddleware(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Printf("[HTTP] %s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
