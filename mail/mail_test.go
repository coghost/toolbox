package mail

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/coghost/xmail"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type MailSuite struct {
	suite.Suite
}

func TestMail(t *testing.T) {
	suite.Run(t, new(MailSuite))
}

func (s *MailSuite) SetupSuite() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func (s *MailSuite) TearDownSuite() {
}

func (s *MailSuite) TestNotify() {
	to := strings.Split(os.Getenv("EMAIL_TO"), ",")
	MAIL.SetupServer(xmail.QQExmailServer, to)
	err := MAIL.Notify(EmailDone, "test body")
	s.Nil(err)
}
