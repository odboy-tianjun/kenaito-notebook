package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
)

type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type EmailService struct {
	cfg      EmailConfig
	configured bool
}

func NewEmailService(cfg EmailConfig) *EmailService {
	configured := cfg.Host != "" && cfg.Username != "" && cfg.Password != ""
	if !configured {
		log.Println("[Email] SMTP not configured, will print codes to console")
	}
	return &EmailService{cfg: cfg, configured: configured}
}

// SendVerificationCode 发送验证码邮件
func (s *EmailService) SendVerificationCode(to, code, purpose string) error {
	var subject, body string
	switch purpose {
	case "register":
		subject = "Kenaito Notebook - 注册验证码"
		body = fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; padding: 20px;">
  <h2 style="color: #333;">Kenaito Notebook</h2>
  <p>你正在注册账号，验证码如下：</p>
  <div style="background: #f5f7fa; padding: 16px; border-radius: 8px; text-align: center; margin: 20px 0;">
    <span style="font-size: 32px; font-weight: bold; color: #409eff; letter-spacing: 8px;">%s</span>
  </div>
  <p style="color: #909399; font-size: 14px;">验证码 5 分钟内有效，请勿泄露给他人。</p>
</div>`, code)
	case "reset":
		subject = "Kenaito Notebook - 重置密码验证码"
		body = fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; padding: 20px;">
  <h2 style="color: #333;">Kenaito Notebook</h2>
  <p>你正在重置密码，验证码如下：</p>
  <div style="background: #f5f7fa; padding: 16px; border-radius: 8px; text-align: center; margin: 20px 0;">
    <span style="font-size: 32px; font-weight: bold; color: #e6a23c; letter-spacing: 8px;">%s</span>
  </div>
  <p style="color: #909399; font-size: 14px;">验证码 5 分钟内有效，请勿泄露给他人。</p>
</div>`, code)
	default:
		return fmt.Errorf("unknown purpose: %s", purpose)
	}

	if !s.configured {
		log.Printf("[Email] SMTP not configured. Verification code for %s (%s): %s", to, purpose, code)
		return nil
	}

	return s.send(to, subject, body)
}

func (s *EmailService) send(to, subject, htmlBody string) error {
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)

	headers := make(map[string]string)
	headers["From"] = s.cfg.From
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	// SSL connection (port 465)
	if s.cfg.Port == "465" {
		return s.sendTLS(addr, auth, to, msg.String())
	}

	// STARTTLS connection (port 587)
	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(msg.String()))
}

func (s *EmailService) sendTLS(addr string, auth smtp.Auth, to, msg string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.cfg.Host,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	if err = client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}
	if _, err = w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("SMTP write failed: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("SMTP close data failed: %w", err)
	}

	return client.Quit()
}
