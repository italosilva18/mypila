package services

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailService() *EmailService {
	return &EmailService{
		host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		port:     getEnv("SMTP_PORT", "587"),
		username: getEnv("SMTP_USERNAME", ""),
		password: getEnv("SMTP_PASSWORD", ""),
		from:     getEnv("SMTP_FROM", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (s *EmailService) SendPasswordResetEmail(toEmail, userName, resetToken, frontendURL string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)

	subject := "Recuperacao de Senha - MyPila"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #0d9488 0%%, #14b8a6 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; background: #0d9488; color: white; padding: 15px 30px; text-decoration: none; border-radius: 8px; font-weight: bold; margin: 20px 0; }
        .button:hover { background: #0f766e; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
        .warning { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>MyPila</h1>
            <p>Recuperacao de Senha</p>
        </div>
        <div class="content">
            <h2>Ola, %s!</h2>
            <p>Recebemos uma solicitacao para redefinir a senha da sua conta.</p>
            <p>Clique no botao abaixo para criar uma nova senha:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Redefinir Senha</a>
            </p>
            <div class="warning">
                <strong>Importante:</strong> Este link expira em 1 hora. Se voce nao solicitou a recuperacao de senha, ignore este email.
            </div>
            <p>Se o botao nao funcionar, copie e cole o link abaixo no seu navegador:</p>
            <p style="word-break: break-all; color: #0d9488;">%s</p>
        </div>
        <div class="footer">
            <p>Este email foi enviado automaticamente. Por favor, nao responda.</p>
            <p>&copy; 2024 MyPila - Gestao Financeira</p>
        </div>
    </div>
</body>
</html>`, userName, resetLink, resetLink)

	return s.sendEmail(toEmail, subject, body)
}

func (s *EmailService) sendEmail(to, subject, htmlBody string) error {
	if s.username == "" || s.password == "" {
		return fmt.Errorf("SMTP credentials not configured")
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	headers := fmt.Sprintf("From: MyPila <%s>\r\n", s.from)
	headers += fmt.Sprintf("Reply-To: %s\r\n", s.from)
	headers += fmt.Sprintf("To: %s\r\n", to)
	headers += fmt.Sprintf("Subject: %s\r\n", subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += "Content-Type: text/html; charset=UTF-8\r\n"
	headers += "X-Priority: 1\r\n"
	headers += "\r\n"

	message := []byte(headers + htmlBody)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	err := smtp.SendMail(addr, auth, s.from, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
