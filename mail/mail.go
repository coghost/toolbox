package mail

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/coghost/xmail"
	"github.com/matcornic/hermes/v2"
)

const (
	hostGmail  = "smtp.gmail.com"
	portGmail  = 587
	hostExmail = "smtp.exmail.qq.com"
	portExmail = 465
)

const (
	EmailDone    = "[✓] Done"
	EmailAlert   = "[✘] Alert"
	EmailCaptcha = "[✘] Captcha"
	Unknown      = "[✘] Unknown"
)

// Mailer represents an email client with configuration and mocking capabilities
type Mailer struct {
	mock bool

	serverCfg xmail.MailCfg
}

// MAIL is the exported mail client
var MAIL = &Mailer{}

// Mock sets the mocking state of the Mailer
func (m *Mailer) Mock(b bool) {
	m.mock = b
}

// SetupServer configures the mail server settings
// server: The type of mail server ("gmail" or "exmail")
// sendTo: A slice of recipient email addresses
func (m *Mailer) SetupServer(server string, sendTo []string) {
	username := os.Getenv("EMAIL_USERNAME")
	password := os.Getenv("EMAIL_PASSWORD")

	recipients := []string{}

	for _, v := range sendTo {
		v = strings.TrimSpace(v)
		if v != "" {
			recipients = append(recipients, v)
		}
	}

	m.serverCfg = xmail.MailCfg{
		From:     username,
		Password: password,
		To:       recipients,
	}

	switch server {
	case "gmail":
		m.serverCfg.Server = xmail.GmailServer
		m.serverCfg.Port = portGmail
		m.serverCfg.Host = hostGmail
	case "exmail":
		m.serverCfg.Server = xmail.QQExmailServer
		m.serverCfg.Port = portExmail
		m.serverCfg.Host = hostExmail
	default:
		log.Fatalf("unsupported server: %s", server)
	}
}

// Notify sends an email notification with the given subject and body
func (m *Mailer) Notify(subject, body string) error {
	if r := recover(); r != nil {
		log.Printf("cannot notify via email: %v", r)
	}

	if err := xmail.VerifyConfig(&m.serverCfg); err != nil {
		return err
	}

	subject = genSubject(subject)

	htmlBody, e := genHTMLBody(EmailDone, body)
	if e != nil {
		return e
	}

	if m.mock {
		log.Println(subject)
		log.Println(htmlBody)

		return nil
	}

	s := xmail.GenMailService(&m.serverCfg)

	return s.Notify(subject, htmlBody)
}

// genSubject generates the email subject with a given hint and the hostname
func genSubject(hint string) string {
	return fmt.Sprintf("%s: HOST %s", hint, hostname())
}

// genHTMLBody generates the HTML body for the email using the Hermes library
func genHTMLBody(title string, body string) (string, error) {
	her := hermes.Hermes{
		Product: hermes.Product{
			Name:        "XMail",
			Copyright:   fmt.Sprintf("Copyright © %d. All rights reserved.", time.Now().Year()),
			TroubleText: "If you have any questions please ask Hex for help.",
		},
	}

	raw := genBodyTable(title, body)

	return her.GenerateHTML(raw)
}

// genBodyTable generates the body table for the email using the Hermes library
func genBodyTable(title, intro string) hermes.Email {
	email := hermes.Email{
		Body: hermes.Body{
			Title: title,
			Intros: []string{
				intro,
			},
			Table: hermes.Table{
				Data: [][]hermes.Entry{
					{
						{Key: "Host", Value: hostname()},
						{Key: "Description", Value: ""},
					},
				},
				Columns: hermes.Columns{
					CustomWidth: map[string]string{
						"Host":        "25%",
						"Description": "75%",
					},
				},
			},
		},
	}

	return email
}

func hostname() string {
	name, _ := os.Hostname()
	return name
}
