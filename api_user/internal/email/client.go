package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/reversersed/go-web-services/tree/main/api_user/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
)

type smtpInfo struct {
	host     string
	port     int
	user     string
	password string
	active   bool
}

var smtpConfig *smtpInfo
var logger *logging.Logger

func init() {
	cfg := config.GetConfig()
	smtpConfig = &smtpInfo{
		host:     cfg.SmtpHost,
		password: cfg.SmtpPassword,
		port:     cfg.SmtpPort,
		user:     cfg.SmtpLogin,
		active:   true,
	}

	if len(smtpConfig.host) == 0 {
		smtpConfig.active = false
	}
	logger = logging.GetLogger()
}
func SendEmailConfirmationMessage(receiver string, userlogin string, code string) bool {
	if !smtpConfig.active {
		logger.Warnf("SMTP hosting is not provided. Confirmation email wasn't sent")
		return false
	}
	auth := smtp.PlainAuth("", smtpConfig.user, smtpConfig.password, smtpConfig.host)

	t, err := template.ParseFiles("templates/email.confirmation.html")
	if err != nil {
		logger.Errorf("can't find or parse html template: %s", err)
		return false
	}
	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	_, err = body.Write([]byte(fmt.Sprintf("From: %s \r\nTo: %s \r\nSubject: Подтверждение почты \n%s\n\n", smtpConfig.user, receiver, mimeHeaders)))
	if err != nil {
		logger.Errorf("can't create email header: %s", err)
		return false
	}

	err = t.Execute(&body, struct {
		ProjectName     string
		Login           string
		ConfirmationURL string
	}{
		ProjectName:     "Example",
		Login:           userlogin,
		ConfirmationURL: fmt.Sprintf("http://example.com/emailLogin?authcode=%s", code),
	})
	if err != nil {
		logger.Errorf("can't create email body: %s", err)
		return false
	}

	err = smtp.SendMail(fmt.Sprintf("%s:%d", smtpConfig.host, smtpConfig.port), auth, smtpConfig.user, []string{receiver}, body.Bytes())
	if err != nil {
		logger.Errorf("can't send email: %s", err)
		return false
	}
	logger.Infof("sent email confirmation message to %s from %s", receiver, smtpConfig.user)
	return true
}
