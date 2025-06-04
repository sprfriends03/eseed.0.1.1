package mail

import (
	"app/env"
	"app/store"
	"bytes"
	"fmt"
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

// VerificationEmailData holds the data for the verification email template.
type VerificationEmailData struct {
	Subject          string
	Username         string
	Email            string
	VerificationLink string
	AppName          string // e.g., Seed eG
}

// SendMemberVerificationEmail sends an email to the user with a link to verify their email address.
func (s Mail) SendMemberVerificationEmail(toEmail, username, verificationToken, verificationLinkBase string) error {
	appName := env.EmailFromName // Or a more general app name from env
	verificationLink := verificationLinkBase + verificationToken

	data := &VerificationEmailData{
		Subject:          "Verify Your Email for " + appName,
		Username:         username,
		Email:            toEmail,
		VerificationLink: verificationLink,
		AppName:          appName,
	}

	htmlBody, err := s.Parse(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>{{.Subject}}</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 0; padding: 20px; color: #333; }
				.container { background-color: #f9f9f9; padding: 20px; border-radius: 5px; }
				.button { background-color: #4CAF50; color: white; padding: 10px 20px; text-align: center; text-decoration: none; display: inline-block; border-radius: 5px; }
			</style>
		</head>
		<body>
			<div class="container">
				<h2>Welcome to {{.AppName}}, {{.Username}}!</h2>
				<p>Please verify your email address to complete your registration and activate your account.</p>
				<p>Click the button below to verify your email:</p>
				<p style="text-align: center;">
					<a href="{{.VerificationLink}}" class="button">Verify Email</a>
				</p>
				<p>If you cannot click the button, please copy and paste the following link into your browser:</p>
				<p><a href="{{.VerificationLink}}">{{.VerificationLink}}</a></p>
				<p>If you did not request this, please ignore this email.</p>
				<p>Thanks,<br>The {{.AppName}} Team</p>
			</div>
		</body>
		</html>`, data)
	if err != nil {
		return err
	}

	uri, err := url.Parse(env.MailUri)
	if err != nil {
		return fmt.Errorf("failed to parse MailUri: %w", err)
	}

	mailHost := uri.Hostname()
	mailPortStr := uri.Port()
	mailPort, err := strconv.Atoi(mailPortStr)
	if err != nil {
		// Attempt to determine default port if not specified or invalid
		if uri.Scheme == "smtps" { // typically 465, but gomail handles STARTTLS on 587 too
			mailPort = 465 // Or 587 if gomail handles it implicitly
		} else { // default smtp
			mailPort = 587 // Or 25
		}
		// logrus.Warnf("Could not parse mail port '%s', using default %d. Error: %v", mailPortStr, mailPort, err)
	}

	mailUser := uri.User.Username()
	mailPass, _ := uri.User.Password()
	fromAddress := env.EmailFromAddress
	if fromAddress == "" {
		fromAddress = mailUser // Fallback if EmailFromAddress is not set
	}

	return s.Send(&Data{
		Host:      mailHost,
		Port:      mailPort,
		Username:  mailUser,
		Password:  mailPass,
		Subject:   data.Subject,
		Html:      htmlBody,
		Receivers: []string{toEmail},
		// Assuming Filenames and Renames are not needed for verification email
	})
}
