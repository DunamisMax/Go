package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
)

// This client will:
// 1. Prompt the user for an IP/port (or use default 127.0.0.1:8080).
// 2. Prompt for a username.
// 3. Connect to the server via WebSocket.
// 4. Send the username as the first message.
// 5. Listen for incoming messages in one goroutine.
// 6. Read user input in main goroutine and send to server.
// 7. Typing "quit" exits the client.

func main() {
	serverAddr := promptServerAddress()
	username := promptUsername()

	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/ws"}

	fmt.Printf("Connecting to %s...\n", u.String())

	// Connect to the server
	conn, err := dialWebSocket(u)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	// Send username as the first message
	if err := writeWebSocketFrame(conn, 0x1, []byte(username)); err != nil {
		fmt.Printf("Failed to send username: %v\n", err)
		return
	}

	// Start a goroutine to read messages from server
	go func() {
		for {
			opcode, payload, err := readWebSocketFrame(conn)
			if err != nil {
				fmt.Println("Connection closed by server.")
				os.Exit(0)
			}
			if opcode == 0x8 {
				fmt.Println("Server sent close frame. Exiting.")
				os.Exit(0)
			}
			fmt.Println(string(payload))
		}
	}()

	// Read user input and send to server
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if !scanner.Scan() {
			// EOF or error
			fmt.Println("Exiting...")
			return
		}
		msg := scanner.Text()
		if strings.ToLower(msg) == "quit" {
			writeWebSocketFrame(conn, 0x8, []byte{})
			return
		}
		if err := writeWebSocketFrame(conn, 0x1, []byte(msg)); err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
			return
		}
	}
}

func promptServerAddress() string {
	fmt.Print("Enter server IP and port (default 127.0.0.1:8080): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	addr := scanner.Text()
	if strings.TrimSpace(addr) == "" {
		addr = "127.0.0.1:8080"
	}
	return addr
}

func promptUsername() string {
	fmt.Print("Enter a username: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := scanner.Text()
	if strings.TrimSpace(username) == "" {
		username = "Anonymous"
	}
	return username
}

func dialWebSocket(u url.URL) (net.Conn, error) {
	// The handshake for WebSocket over standard library net/http requires us to do it manually.
	// We'll do a basic WebSocket handshake.
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return nil, err
	}

	// Perform WebSocket handshake:
	// Generate a Sec-WebSocket-Key and send required headers.
	key := generateWebSocketKey()
	req := fmt.Sprintf("GET %s HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Key: %s\r\n"+
		"Sec-WebSocket-Version: 13\r\n\r\n", u.RequestURI(), u.Host, key)

	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, err
	}

	// Read the response and validate
	resp := bufio.NewReader(conn)
	statusLine, err := resp.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, err
	}
	if !strings.Contains(statusLine, "101") {
		conn.Close()
		return nil, fmt.Errorf("server did not return 101 switching protocols")
	}

	// Read headers until blank line
	for {
		line, err := resp.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, err
		}
		if strings.TrimSpace(line) == "" {
			// Handshake complete
			break
		}
	}

	// Return the same conn, but wrapped in a *bufio.ReadWriter for frame functions
	// We'll just return the net.Conn and let read/writeWebSocketFrame handle it directly.
	return conn, nil
}

func generateWebSocketKey() string {
	// A valid key is a random 16-byte value base64 encoded. For simplicity:
	return "dGhlIHNhbXBsZSBub25jZQ==" // This is a static key used in RFC examples.
}

func readWebSocketFrame(conn net.Conn) (byte, []byte, error) {
	// We need a blocking read. We'll do a small header read first.
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
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
		if _, err := io.ReadFull(conn, ext); err != nil {
			return 0, nil, err
		}
		payloadLen = int64(uint16(ext[0])<<8 | uint16(ext[1]))
	case 127:
		ext := make([]byte, 8)
		if _, err := io.ReadFull(conn, ext); err != nil {
			return 0, nil, err
		}
		payloadLen = int64((uint64(ext[0])<<56 | uint64(ext[1])<<48 |
			uint64(ext[2])<<40 | uint64(ext[3])<<32 |
			uint64(ext[4])<<24 | uint64(ext[5])<<16 |
			uint64(ext[6])<<8 | uint64(ext[7])))
	}

	var maskKey [4]byte
	if mask {
		if _, err := io.ReadFull(conn, maskKey[:]); err != nil {
			return 0, nil, err
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return 0, nil, err
	}

	if mask {
		for i := int64(0); i < payloadLen; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	return opcode, payload, nil
}

func writeWebSocketFrame(conn net.Conn, opcode byte, payload []byte) error {
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

	if _, err := conn.Write(header); err != nil {
		return err
	}
	if _, err := conn.Write(payload); err != nil {
		return err
	}
	return nil
}
