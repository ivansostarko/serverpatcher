package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Message struct {
	From    string
	To      []string
	Subject string
	Text    string

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
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	dialer := net.Dialer{Timeout: cfg.Timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return err
	}
	defer c.Quit()

	if cfg.StartTLS {
		tlsConfig := &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(tlsConfig); err != nil {
				return err
			}
		}
	}

	if cfg.Username != "" || cfg.Password != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err := c.Mail(msg.From); err != nil {
		return err
	}
	for _, rcpt := range msg.To {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	raw := buildMIME(msg)
	_, err = w.Write(raw)
	return err
}

func buildMIME(msg Message) []byte {
	var b bytes.Buffer

	boundary := "sp_boundary_" + fmt.Sprintf("%d", time.Now().UnixNano())

	writeHeader := func(k, v string) {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(v)
		b.WriteString("
")
	}

	writeHeader("From", msg.From)
	writeHeader("To", strings.Join(msg.To, ", "))
	writeHeader("Subject", msg.Subject)
	writeHeader("MIME-Version", "1.0")
	writeHeader("Content-Type", "multipart/mixed; boundary="+boundary)
	b.WriteString("
")

	// Text part
	b.WriteString("--" + boundary + "
")
	b.WriteString("Content-Type: text/plain; charset=utf-8
")
	b.WriteString("Content-Transfer-Encoding: 7bit

")
	b.WriteString(msg.Text)
	if !strings.HasSuffix(msg.Text, "
") {
		b.WriteString("
")
	}
	b.WriteString("
")

	// Attachment (optional)
	if len(msg.JSONAttachment) > 0 {
		name := msg.JSONAttachmentName
		if name == "" {
			name = "report.json"
		}
		enc := make([]byte, base64.StdEncoding.EncodedLen(len(msg.JSONAttachment)))
		base64.StdEncoding.Encode(enc, msg.JSONAttachment)

		b.WriteString("--" + boundary + "
")
		b.WriteString("Content-Type: application/json; name="" + name + ""
")
		b.WriteString("Content-Transfer-Encoding: base64
")
		b.WriteString("Content-Disposition: attachment; filename="" + name + ""

")

		for i := 0; i < len(enc); i += 76 {
			end := i + 76
			if end > len(enc) {
				end = len(enc)
			}
			b.Write(enc[i:end])
			b.WriteString("
")
		}
		b.WriteString("
")
	}

	b.WriteString("--" + boundary + "--
")
	return b.Bytes()
}
