/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	emailv1 "github.com/Stasky745/mailerlite-sre-assignment/api/v1"
	"github.com/Stasky745/mailerlite-sre-assignment/pkg/mailer_send"
	"github.com/Stasky745/mailerlite-sre-assignment/pkg/mailgun"
	"github.com/Stasky745/mailerlite-sre-assignment/pkg/utils"
	"github.com/google/uuid"
)

// EmailReconciler reconciles a Email object
type EmailReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=email.mailerlite.io,resources=emails,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=email.mailerlite.io,resources=emails/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=email.mailerlite.io,resources=emails/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Email object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *EmailReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	email := &emailv1.Email{}
	if err := r.Get(ctx, req.NamespacedName, email); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Set MessageID if it's not
	if email.Status.MessageId == "" {
		email.Status.MessageId = uuid.NewString()
	}

	emailSenderConfig, emailSenderConfig_err := utils.GetEmailConfig(ctx, r.Client, req.Namespace, email.Spec.SenderConfigRef)
	if emailSenderConfig_err != nil {
		l.Error(emailSenderConfig_err, "Unable to fetch Secret", "ApiTokenSecretRef", emailSenderConfig.Spec.ApiTokenSecretRef)
		return ctrl.Result{}, emailSenderConfig_err
	}

	secret, secret_err := utils.GetSecret(ctx, r.Client, req.Namespace, emailSenderConfig.Spec.ApiTokenSecretRef)
	if secret_err != nil {
		l.Error(secret_err, "Unable to fetch Secret", "ApiTokenSecretRef", emailSenderConfig.Spec.ApiTokenSecretRef)
		return ctrl.Result{}, secret_err
	}

	apiKey := string(secret.Data["token"])

	var sendEmailFunc func(ctx context.Context, apiToken string, senderEmail string, recipientEmail string, subject string, body string) (string, error)

	switch strings.ToLower(emailSenderConfig.Spec.Provider) {
	case "mailersend":
		sendEmailFunc = mailer_send.SendEmail
	case "mailgun":
		sendEmailFunc = mailgun.SendEmail
	default:
		err := fmt.Errorf("unsupported provider: %s", emailSenderConfig.Spec.Provider)
		l.Error(err, "Failed to send email")
		email.Status = emailv1.EmailStatus{
			DeliveryStatus: "Failed",
			Error:          err.Error(),
		}
		if err_update := r.Status().Update(ctx, email); err != nil {
			return ctrl.Result{}, err_update
		}

		return ctrl.Result{}, nil
	}

	emailId, sendError := sendEmailFunc(ctx, apiKey, emailSenderConfig.Spec.SenderEmail, email.Spec.RecipientEmail, email.Spec.Subject, email.Spec.Body)
	if sendError != nil {
		l.Error(sendError, "Unable to send email", "email", email.Name)
		email.Status = emailv1.EmailStatus{
			DeliveryStatus: "Failed",
			Error:          sendError.Error(),
			MessageId:      emailId,
		}
	} else {
		l.Info("Email sent successfully", "email", email.Name)
		email.Status = emailv1.EmailStatus{
			DeliveryStatus: "Success",
			MessageId:      emailId,
		}
	}

	// Update the status of the custom resource
	if err := r.Status().Update(ctx, email); err != nil {
		l.Error(err, "Failed to update Email status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EmailReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&emailv1.Email{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return true
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		}).
		Complete(r)
}
