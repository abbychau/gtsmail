# Go SMTP Server

A bare-minimum TCP-level SMTP server implementation in Go for a server to send and receive emails.

## Features

- SMTP server listening on port 25
- Handles basic SMTP commands (HELO/EHLO, MAIL FROM, RCPT TO, DATA)
- Email validation for local domain (@abby.md)
- Support for concurrent connections using goroutines
- MX record lookup
- Custom CRLF scanning for proper SMTP protocol handling

## Installation

```sh
git clone <repository-url>
cd gtsmail
go build
```

## Usage

To start the SMTP server:

```sh
sudo go run main.go
```

Note: Running on port 25 requires root/administrator privileges, and ISP restrictions may apply.

## Email Forwarding

The program includes functionality to:
- Look up MX records for recipient domains
- Forward emails to appropriate mail servers
- Handle multiple MX records with priority ordering
- Support timeout and error handling

## Testing

Run the included tests with:

```sh
go test ./...
```

## File Structure

- [main.go](main.go) - SMTP server implementation
- [send.go](send.go) - Functions to send emails and handle MX records
- [send_test.go](send_test.go) - Test cases

## License

This project is available under the MIT License.