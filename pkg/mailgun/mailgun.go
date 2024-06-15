package mailgun

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go"
)

func VerifyEmailConfig(ctx context.Context, token string, email string) error {
	// Create an instance of the Validator
	v := mailgun.NewEmailValidator(token)

	_, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	e, err := v.ValidateEmail(email, false)
	if err != nil {
		return err
	}

	if e.IsValid {
		return nil
	} else {
		err := fmt.Errorf("email is not valid")
		return err
	}

}

func SendEmail(ctx context.Context, apiToken string, senderEmail string, recipientEmail string, subject string, body string) (string, error) {
	domain := strings.Split(senderEmail, "@")[1]

	mg := mailgun.NewMailgun(domain, apiToken)

	message := mg.NewMessage(senderEmail, subject, body, recipientEmail)

	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, id, err := mg.Send(message)

	return id, err
}
