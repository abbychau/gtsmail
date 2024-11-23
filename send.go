package main

import (
	"bufio"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"
)

// Added structure for MX records
type mxRecord struct {
	host string
	pref uint16
}

// Function to get MX records for a domain
func getMXRecords(domain string) ([]mxRecord, error) {
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return nil, err
	}

	records := make([]mxRecord, len(mxRecords))
	for i, mx := range mxRecords {
		records[i] = mxRecord{
			host: strings.TrimSuffix(mx.Host, "."),
			pref: mx.Pref,
		}
	}

	// Sort by preference (lower is higher priority)
	sort.Slice(records, func(i, j int) bool {
		return records[i].pref < records[j].pref
	})

	return records, nil
}

// Function to send email to a target domain
func sendEmail(from, to string, data []string) error {
	// Extract domain from recipient email
	parts := strings.Split(to, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email address: %s", to)
	}
	domain := parts[1]

	// Get MX records
	mxRecords, err := getMXRecords(domain)
	if err != nil {
		return fmt.Errorf("failed to lookup MX records: %v", err)
	}

	if len(mxRecords) == 0 {
		return fmt.Errorf("no MX records found for domain: %s", domain)
	}

	// Try each MX record in order of preference
	var lastErr error
	for _, mx := range mxRecords {
		err = sendToMX(mx.host, from, to, data)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("failed to send email via any MX server: %v", lastErr)
}

func sendToMX(mxHost, from, to string, data []string) error {
	// Connect to MX server
	conn, err := net.DialTimeout("tcp", mxHost+":25", 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to MX server: %v", err)
	}
	defer conn.Close()

	// Set connection deadlines
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Read greeting
	br := bufio.NewReader(conn)
	_, err = br.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read greeting: %v", err)
	}

	// SMTP conversation
	commands := []struct {
		cmd      string
		expected string
	}{
		{fmt.Sprintf("HELO abby.md\r\n"), "250"},
		{fmt.Sprintf("MAIL FROM:<%s>\r\n", from), "250"},
		{fmt.Sprintf("RCPT TO:<%s>\r\n", to), "250"},
		{"DATA\r\n", "354"},
	}

	for _, cmd := range commands {
		_, err = conn.Write([]byte(cmd.cmd))
		if err != nil {
			return fmt.Errorf("failed to send command %s: %v", cmd.cmd, err)
		}

		response, err := br.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response for %s: %v", cmd.cmd, err)
		}

		if !strings.HasPrefix(response, cmd.expected) {
			return fmt.Errorf("unexpected response for %s: %s", cmd.cmd, response)
		}
	}

	// Send email data
	for _, line := range data {
		_, err = conn.Write([]byte(line + "\r\n"))
		if err != nil {
			return fmt.Errorf("failed to send email data: %v", err)
		}
	}

	// End data
	_, err = conn.Write([]byte(".\r\n"))
	if err != nil {
		return fmt.Errorf("failed to send end of data: %v", err)
	}

	response, err := br.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read data response: %v", err)
	}

	if !strings.HasPrefix(response, "250") {
		return fmt.Errorf("failed to send email: %s", response)
	}

	// Quit
	_, err = conn.Write([]byte("QUIT\r\n"))
	if err != nil {
		return fmt.Errorf("failed to send quit command: %v", err)
	}

	return nil
}

// Example of how to send a new email
func sendNewEmail(from, to string, subject string, body string) error {
	// Construct email headers and body
	data := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}

	return sendEmail(from, to, data)
}
