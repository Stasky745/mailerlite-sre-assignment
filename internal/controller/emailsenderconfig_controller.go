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
)

// EmailSenderConfigReconciler reconciles a EmailSenderConfig object
type EmailSenderConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=email.mailerlite.io,resources=emailsenderconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=email.mailerlite.io,resources=emailsenderconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=email.mailerlite.io,resources=emailsenderconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EmailSenderConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *EmailSenderConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	emailSenderConfig := &emailv1.EmailSenderConfig{}
	if err := r.Get(ctx, req.NamespacedName, emailSenderConfig); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secret, secret_err := utils.GetSecret(ctx, r.Client, req.Namespace, emailSenderConfig.Spec.ApiTokenSecretRef)
	if secret_err != nil {
		l.Error(secret_err, "Unable to fetch Secret", "ApiTokenSecretRef", emailSenderConfig.Spec.ApiTokenSecretRef)
		return ctrl.Result{}, secret_err
	}

	apiKey := string(secret.Data["token"])

	var verifyEmailConfig func(ctx context.Context, token string, email string) error

	switch strings.ToLower(emailSenderConfig.Spec.Provider) {
	case "mailersend":
		verifyEmailConfig = mailer_send.VerifyEmailConfig
	case "mailgun":
		verifyEmailConfig = mailgun.VerifyEmailConfig
	default:
		err := fmt.Errorf("unsupported provider: %s", emailSenderConfig.Spec.Provider)
		l.Error(err, "Failed to set up emailSenderConfig")

		return ctrl.Result{}, nil
	}

	err := verifyEmailConfig(ctx, apiKey, emailSenderConfig.Spec.SenderEmail)
	if err != nil {
		l.Error(err, "Can't verify email", "email", emailSenderConfig.Spec.SenderEmail, "provider", emailSenderConfig.Spec.Provider)
		return ctrl.Result{}, err
	}

	l.Info("EmailSenderConfig", "Email", emailSenderConfig.Spec.SenderEmail)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EmailSenderConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&emailv1.EmailSenderConfig{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return true
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return true
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
