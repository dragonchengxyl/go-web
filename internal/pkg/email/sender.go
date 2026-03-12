package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/studio/platform/configs"
)

type Sender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSender(cfg configs.EmailConfig) *Sender {
	return &Sender{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
	}
}

func (s *Sender) Enabled() bool {
	return s != nil && s.host != "" && s.port > 0 && s.from != ""
}

func (s *Sender) SendPasswordReset(to, username, resetURL string) error {
	if !s.Enabled() {
		return fmt.Errorf("email sender is not configured")
	}

	subject := "Reset your Furry Community password"
	body := fmt.Sprintf(
		"Hello %s,\r\n\r\nWe received a request to reset your password.\r\n\r\nOpen the link below to choose a new password:\r\n%s\r\n\r\nIf you did not request this, you can ignore this email.\r\n",
		username,
		resetURL,
	)

	return s.send(to, subject, body)
}

func (s *Sender) SendEmailVerification(to, username, verifyURL string) error {
	if !s.Enabled() {
		return fmt.Errorf("email sender is not configured")
	}

	subject := "Verify your Furry Community email"
	body := fmt.Sprintf(
		"Hello %s,\r\n\r\nWelcome to Furry Community.\r\n\r\nOpen the link below to verify your email address:\r\n%s\r\n\r\nIf you did not create this account, you can ignore this email.\r\n",
		username,
		verifyURL,
	)

	return s.send(to, subject, body)
}

func (s *Sender) send(to, subject, body string) error {
	msg := []byte(strings.Join([]string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n"))

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	if s.port == 465 {
		return s.sendImplicitTLS(addr, to, msg)
	}
	return s.sendSMTP(addr, to, msg)
}

func (s *Sender) sendSMTP(addr, to string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial failed: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: s.host}
		if isLocalHost(s.host) {
			tlsConfig.InsecureSkipVerify = true
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp starttls failed: %w", err)
		}
	}

	if s.username != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT TO failed: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA failed: %w", err)
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return fmt.Errorf("smtp write failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp close failed: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit failed: %w", err)
	}
	return nil
}

func (s *Sender) sendImplicitTLS(addr, to string, msg []byte) error {
	tlsConfig := &tls.Config{ServerName: s.host}
	if isLocalHost(s.host) {
		tlsConfig.InsecureSkipVerify = true
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("smtp tls dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp new client failed: %w", err)
	}
	defer client.Close()

	if s.username != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT TO failed: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA failed: %w", err)
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return fmt.Errorf("smtp write failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp close failed: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit failed: %w", err)
	}
	return nil
}

func isLocalHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1"
}
