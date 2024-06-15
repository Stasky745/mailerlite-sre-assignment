package utils

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	emailv1 "github.com/Stasky745/mailerlite-sre-assignment/api/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetSecret(ctx context.Context, c client.Client, namespace, secretName string) (*corev1.Secret, error) {
	var secret corev1.Secret
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: secretName}, &secret); err != nil {
		return nil, err
	}
	return &secret, nil
}

func GetEmailConfig(ctx context.Context, c client.Client, namespace, emailSenderConfigName string) (*emailv1.EmailSenderConfig, error) {
	var emailSenderConfig emailv1.EmailSenderConfig
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: emailSenderConfigName}, &emailSenderConfig); err != nil {
		return nil, err
	}
	return &emailSenderConfig, nil
}
