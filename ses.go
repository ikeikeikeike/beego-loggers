package loggers

import (
	"encoding/json"
	"strings"

	. "github.com/astaxie/beego/logs"
	"github.com/goamz/goamz/aws"
	"github.com/ikeikeikeike/gopkg/mailer"
)

type SesWriter struct {
	Name               string   `json:"name"`
	Endpoint           string   `json:"endpoint"`
	AccessKey          string   `json:"accesskey"`
	SecretKey          string   `json:"secretkey"`
	Subject            string   `json:"subject"`
	FromAddress        string   `json:"fromAddress"`
	RecipientAddresses []string `json:"sendTos"`
	Level              int      `json:"level"`
}

func NewSesWriter() LoggerInterface {
	return &SesWriter{Level: LevelTrace}
}

// config like:
//	{
//		"accesskey":"aws access key",
//		"secretkey":"aws secret key",
//		"name":"us-east-1",
//		"endpoint":"https://email.us-east-1.amazonaws.com",
//		"subject":"email title",
//		"fromAddress":"from@example.com",
//		"sendTos":["email1","email2"],
//		"level":LevelError
//	}
func (s *SesWriter) Init(jsonconfig string) error {
	err := json.Unmarshal([]byte(jsonconfig), s)
	if err != nil {
		return err
	}
	return nil
}

func (s *SesWriter) WriteMsg(msg string, level int) error {
	if level > s.Level {
		return nil
	}

	ses := mailer.NewSesMailer()
	ses.E.AddTos(s.RecipientAddresses)
	ses.E.SetSubject(s.Subject)
	ses.E.SetBodyHtml(msg)

	if s.FromAddress != "" {
		ses.E.SetSource(s.FromAddress)
	}
	if len(cl(s.AccessKey)) != 0 && len(cl(s.SecretKey)) != 0 {
		ses.SetAuth(aws.Auth{
			AccessKey: s.AccessKey,
			SecretKey: s.SecretKey,
		})
	}
	if len(cl(s.Name)) != 0 && len(cl(s.Endpoint)) != 0 {
		ses.SetRegion(aws.Region{
			Name:        s.Name,
			SESEndpoint: s.Endpoint,
		})
	}

	return ses.M.SendEmail(ses.E)
}

func (s *SesWriter) Flush() {}

func (s *SesWriter) Destroy() {}

func cl(s string) string {
	return strings.Trim(s, " ")
}

func init() {
	Register("ses", NewSesWriter)
}
