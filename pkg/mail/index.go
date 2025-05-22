package mail

import (
	"app/env"
	"app/store"
	"bytes"
	"net/url"
	"strconv"
	"text/template"

	"gopkg.in/gomail.v2"
)

type Data struct {
	Host      string
	Port      int
	Username  string
	Password  string
	Subject   string
	Html      string
	Renames   []string
	Filenames []string
	Receivers []string
}

type Mail struct {
	store *store.Store
}

func New(store *store.Store) *Mail {
	return &Mail{store}
}

func (s Mail) Send(data *Data) error {
	m := gomail.NewMessage()
	m.SetHeader("From", data.Username)
	m.SetHeader("To", data.Receivers...)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/html", data.Html)

	if len(data.Renames) != len(data.Filenames) {
		data.Renames = data.Filenames
	}
	for i := range data.Filenames {
		m.Attach(data.Filenames[i], gomail.Rename(data.Renames[i]))
	}

	return gomail.NewDialer(data.Host, data.Port, data.Username, data.Password).DialAndSend(m)
}

func (s Mail) Parse(html string, data any) (string, error) {
	tmpl, err := template.New("").Parse(html)
	if err != nil {
		return "", err
	}
	buffer := new(bytes.Buffer)
	if err = tmpl.Execute(buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

type Password struct {
	Subject  string
	Domain   string
	Username string
	Email    string
	Password string
	Keycode  string
}

func (s Mail) SendPassword(data *Password) error {
	html, err := s.Parse(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<style>
				table, th, td {
					padding: 5px;
					border-collapse: collapse;
					border: 1px solid black;
				}
			</style>
		</head>
		<body>
			<table style="width:300px">
				<tr>
					<td>Domain:</td>
					<td><a href="{{.Domain}}" target="_blank">{{.Domain}}</a></td>
				</tr>
				<tr>
					<td>Username:</td>
					<td><b>{{.Username}}</b></td>
				</tr>
				<tr>
					<td>Password:</td>
					<td><b>{{.Password}}</b></td>
				</tr>
				<tr>
					<td>Keycode:</td>
					<td><b>{{.Keycode}}</b></td>
				</tr>
			</table>
		</body>
		</html>`,
		data,
	)
	if err != nil {
		return err
	}

	uri, _ := url.Parse(env.MailUri)
	mailHost := uri.Hostname()
	mailPort, _ := strconv.Atoi(uri.Port())
	mailUser := uri.User.Username()
	mailPass, _ := uri.User.Password()

	return s.Send(&Data{Host: mailHost, Port: mailPort, Username: mailUser, Password: mailPass, Subject: data.Subject, Html: html, Receivers: []string{data.Email}})
}
