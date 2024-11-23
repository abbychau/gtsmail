package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type EmailData struct {
	From string
	To   []string
	Data []string
}

func main() {
	listener, err := net.Listen("tcp", ":25")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("SMTP server listening on port 25")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	email := &EmailData{
		To:   make([]string, 0),
		Data: make([]string, 0),
	}

	// Send greeting
	respond(conn, "220 abby.md SMTP server ready")

	scanner := NewScanner(conn)
	state := "INIT"
	var dataMode bool

	for scanner.Scan() {
		line := scanner.Text()
		if dataMode {
			if line == "." {
				dataMode = false
				// Process the email
				processEmail(email)
				respond(conn, "250 Ok: message received")
				state = "INIT"
				continue
			}
			email.Data = append(email.Data, line)
			continue
		}

		cmd := strings.ToUpper(strings.Fields(line)[0])
		switch cmd {
		case "HELO", "EHLO":
			respond(conn, "250 Hello")
			state = "MAIL"
		case "MAIL":
			if state != "MAIL" {
				respond(conn, "503 Bad sequence of commands")
				continue
			}
			email.From = extractEmail(line)
			respond(conn, "250 Ok")
			state = "RCPT"
		case "RCPT":
			if state != "RCPT" && state != "DATA" {
				respond(conn, "503 Bad sequence of commands")
				continue
			}
			rcpt := extractEmail(line)
			if !strings.HasSuffix(rcpt, "@abby.md") {
				respond(conn, "550 No such user here")
				continue
			}
			email.To = append(email.To, rcpt)
			respond(conn, "250 Ok")
			state = "DATA"
		case "DATA":
			if state != "DATA" {
				respond(conn, "503 Bad sequence of commands")
				continue
			}
			respond(conn, "354 End data with <CR><LF>.<CR><LF>")
			dataMode = true
		case "QUIT":
			respond(conn, "221 Bye")
			return
		default:
			respond(conn, "500 Unknown command")
		}
	}
}

func respond(conn net.Conn, msg string) {
	conn.Write([]byte(msg + "\r\n"))
}

func extractEmail(line string) string {
	start := strings.Index(line, "<")
	end := strings.Index(line, ">")
	if start >= 0 && end > start {
		return line[start+1 : end]
	}
	return strings.Fields(line)[1]
}

func processEmail(email *EmailData) {
	// Here you can implement your email storage/processing logic
	log.Printf("Received email from: %s\n", email.From)
	log.Printf("To: %v\n", email.To)
	log.Printf("Content length: %d lines\n", len(email.Data))
}

// NewScanner creates a scanner that splits on CRLF
func NewScanner(conn net.Conn) *bufio.Scanner {
	scanner := bufio.NewScanner(conn)
	scanner.Split(scanCRLF)
	return scanner
}

func scanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := strings.Index(string(data), "\r\n"); i >= 0 {
		return i + 2, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
