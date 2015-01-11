package loggers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	. "github.com/astaxie/beego/logs"
)

type MailcatcherWriter struct {
	Host               string   `json:"Host"`
	Subject            string   `json:"subject"`
	FromAddress        string   `json:"fromAddress"`
	RecipientAddresses []string `json:"sendTos"`
	Level              int      `json:"level"`
	Tls                bool     `json:"tls"`
}

func NewMailcatcherWriter() LoggerInterface {
	return &MailcatcherWriter{Host: "127.0.0.1:1025", Level: LevelTrace, Tls: false}
}

// config like:
//	{
//		"host":"127.0.0.1:1025",
//		"subject":"email title",
//		"fromAddress":"from@example.com",
//		"sendTos":["email1","email2"],
//		"level":LevelError,
//		"tls":false
//	}
func (s *MailcatcherWriter) Init(jsonconfig string) error {
	err := json.Unmarshal([]byte(jsonconfig), s)
	if err != nil {
		return err
	}
	return nil
}

func (s *MailcatcherWriter) sendMail(hostAddressWithPort string, fromAddress string, recipients []string, msgContent []byte) error {
	client, err := smtp.Dial(hostAddressWithPort)
	if err != nil {
		return err
	}

	if s.Tls {
		host, _, _ := net.SplitHostPort(hostAddressWithPort)
		tlsConn := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
		}
		if err = client.StartTLS(tlsConn); err != nil {
			return err
		}
	}

	if err = client.Mail(fromAddress); err != nil {
		return err
	}

	for _, rec := range recipients {
		if err = client.Rcpt(rec); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msgContent))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	err = client.Quit()
	if err != nil {
		return err
	}

	return nil
}

func (s *MailcatcherWriter) WriteMsg(msg string, level int) error {
	if level > s.Level {
		return nil
	}

	content_type := "Content-Type: text/plain" + "; charset=UTF-8"
	mailmsg := []byte("To: " + strings.Join(s.RecipientAddresses, ";") + "\r\nFrom: " + s.FromAddress + "<" + s.FromAddress +
		">\r\nSubject: " + s.Subject + "\r\n" + content_type + "\r\n\r\n" + fmt.Sprintf(".%s", time.Now().Format("2006-01-02 15:04:05")) + msg)

	err := s.sendMail(s.Host, s.FromAddress, s.RecipientAddresses, mailmsg)

	return err
}

func (s *MailcatcherWriter) Flush() {
	return
}

func (s *MailcatcherWriter) Destroy() {
	return
}

func init() {
	Register("mailcatcher", NewMailcatcherWriter)
}
