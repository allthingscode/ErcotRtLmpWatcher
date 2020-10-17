package notify

// TODO:  Organize this with an event/subscriber model

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type configuration struct {
	Oauth2ConfigClientID     string
	Oauth2ConfigClientSecret string
	Oauth2TokenAccessToken   string
	Oauth2TokenRefreshToken  string
	To                       string
	SubjectPrefix            string
}

var config = configuration{}

// Configure initializes settings for this package
func Configure(
	Oauth2ConfigClientID string,
	Oauth2ConfigClientSecret string,
	Oauth2TokenAccessToken string,
	Oauth2TokenRefreshToken string,
	To string,
	SubjectPrefix string,
) {
	config.Oauth2ConfigClientID = Oauth2ConfigClientID
	config.Oauth2ConfigClientSecret = Oauth2ConfigClientSecret
	config.Oauth2TokenAccessToken = Oauth2TokenAccessToken
	config.Oauth2TokenRefreshToken = Oauth2TokenRefreshToken
	config.To = To
	config.SubjectPrefix = SubjectPrefix
}

// SendEmail will send an email to me
func SendEmail(subject string, body string) error {

	gmailService, err := OAuthGmailService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
	}

	gmailEmailTo := "To: " + config.To + "\r\n"
	gmailEmailSubject := "Subject: " + config.SubjectPrefix + "  " + subject + "\n"
	gmailEmailBody := body + "\r\n"
	gmailEmailMime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	gmailEmailMessage := []byte(gmailEmailTo + gmailEmailSubject + gmailEmailMime + "\n" + gmailEmailBody)

	var message gmail.Message
	message.Raw = base64.URLEncoding.EncodeToString(gmailEmailMessage)

	_, err = gmailService.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return err
	}
	return nil
}

// OAuthGmailService creates a gmail service client
// https://medium.com/wesionary-team/sending-emails-with-go-golang-using-smtp-gmail-and-oauth2-185ee12ab306
// TODO:  Put all the settings in a config file
func OAuthGmailService() (*gmail.Service, error) {

	oauth2Config := oauth2.Config{
		ClientID:     config.Oauth2ConfigClientID,
		ClientSecret: config.Oauth2ConfigClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost",
	}

	token := oauth2.Token{
		AccessToken:  config.Oauth2TokenAccessToken,
		RefreshToken: config.Oauth2TokenRefreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now(),
	}

	var tokenSource = oauth2Config.TokenSource(context.Background(), &token)

	srv, err := gmail.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, errors.New("gmail service not initialized")
	}

	return srv, nil
}
