package engine

import (
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/cyops-se/dd-inserter/db"
)

func InitMonitor() {
	InitSetting("monitor.smtp", "localhost", "SMTP server for outgoing alert e-mails")
	InitSetting("monitor.recipients", "", "Comma separated list of e-mail recipients")
	InitSetting("monitor.from", "no-reply@acme.com", "E-mail address representing the sender")
	InitSetting("monitor.subject", "Alert from dd-inserter", "Text representing the e-mail subject")
	InitSetting("monitor.usemx", "false", "If this is true, the mailer will find the recipient MX record instead of using the specified SMTP server")
}

func SendAlerts() {
	subject := getValue("monitor.subject")
	from := getValue("monitor.from")
	recipients := getValue("monitor.recipients")
	usemx := getValue("monitor.usemx")
	smtp := getValue("monitor.smtp")
	emails := strings.Split(recipients, ",")
	for _, to := range emails {
		if to != "" {
			sendAlert(from, to, subject, smtp, usemx)
		}
	}
}

func SendTestAlerts(recipients string) {
	subject := fmt.Sprintf("**TEST** %s **TEST**", getValue("monitor.subject"))
	from := getValue("monitor.from")
	usemx := getValue("monitor.usemx")
	smtp := getValue("monitor.smtp")
	emails := strings.Split(recipients, ",")
	for _, to := range emails {
		sendAlert(from, to, subject, smtp, usemx)
	}
}

func getValue(name string) string {
	value, _ := GetSetting(name)
	return value.Value
}

func sendAlert(from string, email string, subject string, host string, usemx string) (err error) {
	to := []string{email}

	if usemx == "true" {
		host, err = getMXRecord(email)
		if err != nil {
			db.Error("Error sending e-mail", "Failed to retrieve MX record for e-mail %s, error: %s", email, err.Error())
			return err
		}
	}

	target := fmt.Sprintf("%s:25", host)
	header := fmt.Sprintf("From: <%s>\r\nTo: <%s>\r\nSubject: %s", from, email, subject)
	body := "This is an automated alert from DD-INSERTER to inform you that data is not recevied as expected from the inside of the data diode."

	message := []byte(fmt.Sprintf("%s\r\n\r\n%s", header, body))

	// Send actual message
	err = smtp.SendMail(target, nil, from, to, message)
	if err != nil {
		db.Error("Error sending e-mail", "Sending mail to SMTP server: %s, to %s, from %s give error: %s", target, email, from, err.Error())
	}

	return err
}

func getMXRecord(to string) (mx string, err error) {
	var e *mail.Address
	e, err = mail.ParseAddress(to)
	if err != nil {
		return
	}

	domain := strings.Split(e.Address, "@")[1]

	var mxs []*net.MX
	mxs, err = net.LookupMX(domain)

	if err != nil {
		return
	}

	// mx = strings.TrimSuffix(mxs[len(mxs)-1].Host, ".")
	mx = strings.TrimSuffix(mxs[0].Host, ".")

	return
}
