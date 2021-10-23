package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

//go:embed email-templates
var emailTemplateFS embed.FS

func (app *application) SendMail(from, to, subject, tmpl string, attachments map[string]string, data interface{}) error {

	templateToRender := fmt.Sprintf("email-templates/%s.html.tmpl", tmpl)

	// html template
	t, err := template.New("email-html").ParseFS(emailTemplateFS, templateToRender)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	// store template into buffer
	var tpl bytes.Buffer
	// "body" is defined in template as {{define "body"}}
	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println(err)
		return err
	}

	// html email version
	formattedMessage := tpl.String()

	// plain text template
	templateToRender = fmt.Sprintf("email-templates/%s.plain.tmpl", tmpl)
	t, err = template.New("email-plain").ParseFS(emailTemplateFS, templateToRender)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println(err)
		return err
	}

	// plain text email version
	plainMessage := tpl.String()

	// send mail
	server := mail.NewSMTPClient()
	server.Host = app.config.smtp.host
	server.Port = app.config.smtp.port
	server.Username = app.config.smtp.username
	server.Password = app.config.smtp.password
	server.Encryption = mail.EncryptionTLS
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// connect to server
	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	// send email
	email := mail.NewMSG()
	email.SetFrom(from).
		AddTo(to).
		SetSubject(subject)

	// set html body
	email.SetBody(mail.TextHTML, formattedMessage)

	// set plain text body
	email.AddAlternative(mail.TextPlain, plainMessage)

	// add attachments
	if len(attachments) > 0 {
		for k, v := range attachments {
			f := &mail.File{
				FilePath: v,
				Name:     k,
			}
			email.Attach(f)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	app.infoLog.Println("send mail")

	return nil
}
