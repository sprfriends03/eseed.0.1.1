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

// KYCSubmissionEmailData holds the data for KYC submission confirmation email template
type KYCSubmissionEmailData struct {
	Subject  string
	Username string
	Email    string
	AppName  string
}

// SendKYCSubmissionConfirmation sends a confirmation email when a member submits KYC for verification
func (s Mail) SendKYCSubmissionConfirmation(toEmail, username string) error {
	appName := env.EmailFromName

	data := &KYCSubmissionEmailData{
		Subject:  "KYC Submission Confirmation - " + appName,
		Username: username,
		Email:    toEmail,
		AppName:  appName,
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
				.status-box { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 5px; margin: 15px 0; }
				.info-box { background-color: #d1ecf1; border: 1px solid #bee5eb; padding: 15px; border-radius: 5px; margin: 15px 0; }
			</style>
		</head>
		<body>
			<div class="container">
				<h2>KYC Submission Received - {{.AppName}}</h2>
				<p>Hello {{.Username}},</p>
				<p>We have successfully received your Know Your Customer (KYC) documents for verification.</p>
				
				<div class="status-box">
					<strong>Status:</strong> Submitted for Review<br>
					<strong>Submitted On:</strong> {{ "now" | date "January 2, 2006 at 3:04 PM" }}
				</div>

				<div class="info-box">
					<h3>What happens next?</h3>
					<ul>
						<li>Our compliance team will review your submitted documents</li>
						<li>Verification typically takes 1-3 business days</li>
						<li>You will receive an email notification with the verification result</li>
						<li>Once approved, you'll have full access to all platform features</li>
					</ul>
				</div>

				<p><strong>Important:</strong> Please do not submit additional documents unless requested, as this may delay the verification process.</p>
				
				<p>If you have any questions about the verification process, please contact our support team.</p>
				
				<p>Thank you for your patience,<br>The {{.AppName}} Compliance Team</p>
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
		if uri.Scheme == "smtps" {
			mailPort = 465
		} else {
			mailPort = 587
		}
	}

	mailUser := uri.User.Username()
	mailPass, _ := uri.User.Password()
	fromAddress := env.EmailFromAddress
	if fromAddress == "" {
		fromAddress = mailUser
	}

	return s.Send(&Data{
		Host:      mailHost,
		Port:      mailPort,
		Username:  mailUser,
		Password:  mailPass,
		Subject:   data.Subject,
		Html:      htmlBody,
		Receivers: []string{toEmail},
	})
}

// KYCApprovalEmailData holds the data for KYC approval notification email template
type KYCApprovalEmailData struct {
	Subject  string
	Username string
	Email    string
	AppName  string
}

// SendKYCApprovalNotification sends a notification email when KYC is approved
func (s Mail) SendKYCApprovalNotification(toEmail, username string) error {
	appName := env.EmailFromName

	data := &KYCApprovalEmailData{
		Subject:  "KYC Verified - Welcome to " + appName,
		Username: username,
		Email:    toEmail,
		AppName:  appName,
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
				.success-box { background-color: #d4edda; border: 1px solid #c3e6cb; padding: 15px; border-radius: 5px; margin: 15px 0; }
				.button { background-color: #28a745; color: white; padding: 12px 25px; text-align: center; text-decoration: none; display: inline-block; border-radius: 5px; margin: 10px 0; }
				.features-box { background-color: #e2e3e5; border: 1px solid #d6d8db; padding: 15px; border-radius: 5px; margin: 15px 0; }
			</style>
		</head>
		<body>
			<div class="container">
				<h2>🎉 Congratulations! Your KYC is Verified</h2>
				<p>Hello {{.Username}},</p>
				
				<div class="success-box">
					<strong>✅ Verification Complete</strong><br>
					Your identity verification has been successfully completed and approved.
				</div>

				<p>You now have full access to all {{.AppName}} features and services.</p>

				<div class="features-box">
					<h3>What you can do now:</h3>
					<ul>
						<li>Access all premium features</li>
						<li>Complete transactions without limits</li>
						<li>Participate in exclusive member activities</li>
						<li>Enjoy enhanced security protections</li>
					</ul>
				</div>

				<p style="text-align: center;">
					<a href="#" class="button">Access Your Account</a>
				</p>

				<p>Thank you for completing the verification process. We're excited to have you as a verified member of our community!</p>
				
				<p>If you have any questions, please don't hesitate to contact our support team.</p>
				
				<p>Welcome aboard,<br>The {{.AppName}} Team</p>
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
		if uri.Scheme == "smtps" {
			mailPort = 465
		} else {
			mailPort = 587
		}
	}

	mailUser := uri.User.Username()
	mailPass, _ := uri.User.Password()
	fromAddress := env.EmailFromAddress
	if fromAddress == "" {
		fromAddress = mailUser
	}

	return s.Send(&Data{
		Host:      mailHost,
		Port:      mailPort,
		Username:  mailUser,
		Password:  mailPass,
		Subject:   data.Subject,
		Html:      htmlBody,
		Receivers: []string{toEmail},
	})
}

// KYCRejectionEmailData holds the data for KYC rejection notification email template
type KYCRejectionEmailData struct {
	Subject  string
	Username string
	Email    string
	Reason   string
	AppName  string
}

// SendKYCRejectionNotification sends a notification email when KYC is rejected
func (s Mail) SendKYCRejectionNotification(toEmail, username, reason string) error {
	appName := env.EmailFromName

	data := &KYCRejectionEmailData{
		Subject:  "KYC Verification Update - " + appName,
		Username: username,
		Email:    toEmail,
		Reason:   reason,
		AppName:  appName,
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
				.warning-box { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 5px; margin: 15px 0; }
				.reason-box { background-color: #f8d7da; border: 1px solid #f5c6cb; padding: 15px; border-radius: 5px; margin: 15px 0; }
				.action-box { background-color: #d1ecf1; border: 1px solid #bee5eb; padding: 15px; border-radius: 5px; margin: 15px 0; }
				.button { background-color: #007bff; color: white; padding: 12px 25px; text-align: center; text-decoration: none; display: inline-block; border-radius: 5px; margin: 10px 0; }
			</style>
		</head>
		<body>
			<div class="container">
				<h2>KYC Verification Update</h2>
				<p>Hello {{.Username}},</p>
				
				<div class="warning-box">
					<strong>⚠️ Additional Information Required</strong><br>
					Your recent KYC submission requires additional review before verification can be completed.
				</div>

				{{if .Reason}}
				<div class="reason-box">
					<h3>Review Details:</h3>
					<p>{{.Reason}}</p>
				</div>
				{{end}}

				<div class="action-box">
					<h3>Next Steps:</h3>
					<ul>
						<li>Please review the details above</li>
						<li>Prepare updated or additional documentation as needed</li>
						<li>Resubmit your KYC documents through your account dashboard</li>
						<li>Ensure all documents are clear, complete, and current</li>
					</ul>
				</div>

				<p style="text-align: center;">
					<a href="#" class="button">Update KYC Documents</a>
				</p>

				<p><strong>Tips for successful verification:</strong></p>
				<ul>
					<li>Ensure documents are high-quality images or PDFs</li>
					<li>All text should be clearly visible and legible</li>
					<li>Documents should not be expired</li>
					<li>Personal information should match across all documents</li>
				</ul>

				<p>If you have questions about this review or need assistance with your resubmission, please contact our support team.</p>
				
				<p>Thank you for your understanding,<br>The {{.AppName}} Compliance Team</p>
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
		if uri.Scheme == "smtps" {
			mailPort = 465
		} else {
			mailPort = 587
		}
	}

	mailUser := uri.User.Username()
	mailPass, _ := uri.User.Password()
	fromAddress := env.EmailFromAddress
	if fromAddress == "" {
		fromAddress = mailUser
	}

	return s.Send(&Data{
		Host:      mailHost,
		Port:      mailPort,
		Username:  mailUser,
		Password:  mailPass,
		Subject:   data.Subject,
		Html:      htmlBody,
		Receivers: []string{toEmail},
	})
}
