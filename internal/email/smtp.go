package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type Message struct {
	From               string
	To                 []string
	Subject            string
	Text               string
	JSONAttachmentName string
	JSONAttachment     []byte
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	StartTLS bool
	Timeout  time.Duration
}

func Send(cfg SMTPConfig, msg Message) error {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Use smtp.Dial with proper timeout instead of manual dial
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Close()

	// Set deadline
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("helo: %w", err)
	}

	// StartTLS if requested and supported
	if cfg.StartTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName: cfg.Host,
				MinVersion: tls.VersionTLS12,
			}
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("starttls: %w", err)
			}
		}
	}

	// Auth if credentials provided
	if cfg.Username != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("auth failed: %w", err)
			}
		} else {
			return fmt.Errorf("auth required but not supported by server")
		}
	}

	// Send mail
	if err := client.Mail(msg.From); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	for _, to := range msg.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("RCPT TO <%s> failed: %w", to, err)
		}
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("sending message body: %w", err)
	}

	// Properly write the MIME message with correct \r\n
	rawMessage := buildMIME(msg)
	_, err = wc.Write(rawMessage)
	if err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return client.Quit()
}

// buildMIME builds a correct multipart/mixed message with proper CRLF
func buildMIME(msg Message) []byte {
	var buf bytes.Buffer

	boundary := "boundary_" + fmt.Sprintf("%09d", time.Now().UnixNano())

	// Helper to write header with proper \r\n
	header := func(key, value string) {
		// Fold long headers properly if needed (basic folding)
		if len(value) > 78 {
			fmt.Fprintf(&buf, "%s: %s\r\n", key, value[:78])
			for i := 78; i < len(value); i += 77 {
				end := i + 77
				if end > len(value) {
					end = len(value)
				}
				if i+77 < len(value) {
					fmt.Fprintf(&buf, " %s\r\n", value[i:end])
				} else {
					fmt.Fprintf(&buf, " %s\r\n", value[i:])
				}
			}
		} else {
			fmt.Fprintf(&buf, "%s: %s\r\n", key, value)
		}
	}

	header("From", msg.From)
	header("To", strings.Join(msg.To, ", "))
	header("Subject", msg.Subject)
	header("MIME-Version", "1.0")
	header("Content-Type", "multipart/mixed; boundary=\""+boundary+"\"")

	buf.WriteString("\r\n") // Empty line after headers

	// -- Text Part
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(msg.Text)
	if msg.Text != "" && !strings.HasSuffix(msg.Text, "\n") {
		buf.WriteString("\r\n")
	}
	buf.WriteString("\r\n") // end of part

	// -- JSON Attachment (if any)
	if len(msg.JSONAttachment) > 0 {
		filename := msg.JSONAttachmentName
		if filename == "" {
			filename = "data.json"
		}

		buf.WriteString("--" + boundary + "\r\n")
		buf.WriteString(fmt.Sprintf("Content-Type: application/json; name=\"%s\"\r\n", filename))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
		buf.WriteString("\r\n")

		encoder := base64.NewEncoder(base64.StdEncoding, &chunkWriter{buf: &buf, chunkSize: 76})
		_, _ = encoder.Write(msg.JSONAttachment)
		encoder.Close()
		buf.WriteString("\r\n") // end of part
	}

	// Final boundary
	buf.WriteString("--" + boundary + "--\r\n")

	return buf.Bytes()
}

// chunkWriter writes base64 data in 76-byte chunks followed by \r\n
type chunkWriter struct {
	buf       *bytes.Buffer
	chunkSize int
	lineLen   int
}

func (w *chunkWriter) Write(p []byte) (n int, err error) {
	for len(p) > 0 {
		remaining := w.chunkSize - w.lineLen
		if remaining == 0 {
			w.buf.WriteString("\r\n")
			w.lineLen = 0
			remaining = w.chunkSize
		}

		toWrite := len(p)
		if toWrite > remaining {
			toWrite = remaining
		}

		n += toWrite
		w.buf.Write(p[:toWrite])
		w.lineLen += toWrite
		p = p[toWrite:]
	}
	return n, nil
}