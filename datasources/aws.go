package datasources

import (
	"log"
	"start/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

/*
Send email
*/
func SendEmail(recipient, subject, htmlBody, fromEmail string) error {
	config := config.GetConfig()
	charSet := "UTF-8"
	sender := fromEmail
	sesConfig := &aws.Config{
		Region: aws.String(config.GetString("ses.region")),
		Credentials: credentials.NewStaticCredentials(
			config.GetString("ses.id_key"),
			config.GetString("ses.private_key"),
			"",
		)}
	sess, errNewSession := session.NewSession(sesConfig)
	if errNewSession != nil {
		return errNewSession
	}
	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String("textBody"),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, errSendEmail := svc.SendEmail(input)
	log.Println(result)
	return errSendEmail
}
