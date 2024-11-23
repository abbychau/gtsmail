package main

import (
	"testing"
)

func TestSendNewEmail(t *testing.T) {
	from := "sender@example.com"
	to := "recipient@example.com"
	subject := "Test Subject"
	body := "This is a test email body."

	err := sendNewEmail(from, to, subject, body)
	if err != nil {
		t.Errorf("sendNewEmail() error = %v", err)
	}
}
