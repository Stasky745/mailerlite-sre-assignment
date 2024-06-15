package mailer_send

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mailersend/mailersend-go"
)

func VerifyEmailConfig(ctx context.Context, token string, email string) error {
	// Create an instance of mailersend client
	ms := mailersend.NewMailersend(token)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// I check if the domain is in the list of domains for the account
	// Wanted to use other methods of verification but weren't working

	options := &mailersend.ListDomainOptions{
		Page:  1,
		Limit: 25,
	}

	a, _, err := ms.Domain.List(ctx, options)

	if a == nil {
		err := fmt.Errorf("no domain list available")
		return err
	}

	domain := strings.Split(email, "@")[1]

	for i := 0; i < len(a.Data); i++ {
		if a.Data[i].Name == domain {
			return nil
		}
	}

	return err

}

func SendEmail(ctx context.Context, apiToken string, senderEmail string, recipientEmail string, subject string, body string) (string, error) {

	// Create an instance of the mailersend client
	ms := mailersend.NewMailersend(apiToken)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	from := mailersend.From{
		Email: senderEmail,
	}

	recipients := []mailersend.Recipient{
		{
			Email: recipientEmail,
		},
	}

	message := ms.Email.NewMessage()

	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(subject)
	message.SetText(body)

	res, err := ms.Email.Send(ctx, message)

	return res.Header.Get("X-Message-Id"), err
}
