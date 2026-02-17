package MailLog

import (
	"bytes"
	"flag"
	goMail "net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sa-kemper/peertubestats/assets/mail"
	"github.com/sa-kemper/peertubestats/internal/LogHelp"
)

var SmtpConf struct {
	Host        string `json:"smtp_host"`
	Port        int    `json:"smtp_port"`
	Username    string `json:"-"`
	Password    string `json:"-"`
	FromAddress string `json:"smtp_from_address"`
	ToAddress   string `json:"smtp_to_address"`
}

var LogBuffer *bytes.Buffer

func init() {
	flag.StringVar(&SmtpConf.Host, "smtpHost", "localhost", "SMTP server host")
	flag.IntVar(&SmtpConf.Port, "smtpPort", 25, "SMTP server port")
	flag.StringVar(&SmtpConf.Username, "smtpUsername", "", "SMTP username")
	flag.StringVar(&SmtpConf.Password, "smtpPassword", "", "SMTP password")
	flag.StringVar(&SmtpConf.FromAddress, "smtpFromAddress", "peertubestats@localhost", "SMTP from address, will use peertubestats @ smtpHost by default.")
	flag.StringVar(&SmtpConf.ToAddress, "smtpToAddress", "admin <root@localhost>", "the recipient list of the administrators EG. Alice <alice@example.com>, Bob <bob@example.com>, Eve <eve@example.com>")
	LogBuffer = new(bytes.Buffer)
}

// SendPanic is a function that is called when peertubestats is in distress, something extremely unexpected happen and the archival features of peertubestats cannot be guaranteed
// it takes a mail.PanicMail struct to deliver the mail, and may error, this should be handled using a file
// NOTE: maybe implement a warning in the frontend when this occurs and or fails
func SendPanic(panicMail mail.PanicMail) (err error) {
	println("[PANIC] panicMail being sent")
	// Send to
	var mailAddresses []*goMail.Address
	mailAddresses, err = goMail.ParseAddressList(SmtpConf.ToAddress)
	if err != nil {
		return
	}
	var mailAddressStrings []string

	for _, mailAddress := range mailAddresses {
		mailAddressStrings = append(mailAddressStrings, mailAddress.Address)
	}

	// Auth
	auth := smtp.PlainAuth("", SmtpConf.Username, SmtpConf.Password, SmtpConf.Host)

	// Message body
	var buf = make([]byte, 0)
	var message = bytes.NewBuffer(buf)

	message.WriteString("From: " + SmtpConf.FromAddress + "\r\n")
	message.WriteString("To: " + strings.Join(mailAddressStrings, ",") + "\r\n")
	err = mail.Templates.ExecuteTemplate(message, "fatal", panicMail)
	if err != nil {
		return
	}

	err = os.WriteFile("peertube-stats-log-from-"+time.Now().Format("2006-01-02-Time-15-04-05"+".txt"), LogBuffer.Bytes(), 0600)
	if err != nil {
		println(err.Error())
	}

	err = smtp.SendMail(SmtpConf.Host+":"+strconv.Itoa(SmtpConf.Port), auth, SmtpConf.FromAddress, mailAddressStrings, message.Bytes())
	return err
}

// SendMailOnFatalLog reads the log type of each log message and sends a panic mail when a fatal log message was recorded.
// this records and sends along the entire log of the application, this may cause higher memory use
func SendMailOnFatalLog() {
	var err error
	for LogBuffer == nil {
		time.Sleep(1 * time.Millisecond)
	}
	for {
		log := <-LogHelp.LogQueue
		if log != nil {
			LogBuffer.WriteString(log.String() + "\n")
		}
		if log == nil || log.LogLevelInt == LogHelp.Fatal {
			err = nil
			counter := 0
			err = SendPanic(mail.PanicMail{
				IncidentTimestamp: strconv.Itoa(int(time.Now().Unix())),
				ErrorMessage:      "A Fatal error has occurred",
				ErrorDetails:      LogBuffer.String(),
			})
			for err != nil && counter < 10 {
				// send until success
				err = SendPanic(mail.PanicMail{
					IncidentTimestamp: strconv.Itoa(int(time.Now().Unix())),
					ErrorMessage:      "A Fatal error has occurred",
					ErrorDetails:      LogBuffer.String(),
				})
				counter++
				time.Sleep(3 * time.Second)
				println("MAIL SEND ERROR")
				println(err.Error())
			}

			if counter >= 10 {
				_ = os.WriteFile("PANICMESSAGE.txt", LogBuffer.Bytes(), 0600)
				panic(err)
			}
			os.Exit(1)
		}
	}
}
