package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	//"net/mail"
	"time"

	"github.com/dpramod/bookings/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

func listenForMail() {
	go func() {
		for {

			msg := <-app.MailChan

			sendMsg(msg)
		}
	}()
}

func sendMsg(m models.MailData) {
	//server := mail.NewSMTPClient()
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		//log.Println(err)
		errorLog.Println(err)
	}
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)

	if m.Template == " " {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		data, err := ioutil.ReadFile(fmt.Sprintf("./email-templates/%s", m.Template))

		if err != nil {
			app.ErrorLog.Println(err)
		}
		mailTemplate := string(data)
		msgToSend := strings.Replace(mailTemplate, "[%Body%] ", m.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	//email.SetBody(mail.TextHTML, m.Content)

	err = email.Send(client)

	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email sent !")
	}
}
