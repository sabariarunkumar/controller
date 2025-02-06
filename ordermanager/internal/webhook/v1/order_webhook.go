/*
Copyright 2025.

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

package v1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	managev1 "order-manager/api/v1"
)

// nolint:unused
// log is for logging in this package.
var orderlog = logf.Log.WithName("order-resource")

// SetupOrderWebhookWithManager registers the webhook for Order in the manager.
func SetupOrderWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&managev1.Order{}).
		WithValidator(&OrderCustomValidator{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-manage-sap-v1-order,mutating=false,failurePolicy=fail,sideEffects=None,groups=manage.sap,resources=orders,verbs=create;update,versions=v1,name=vorder-v1.kb.io,admissionReviewVersions=v1

// OrderCustomValidator struct is responsible for validating the Order resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type OrderCustomValidator struct {
	Client client.Client
}

var _ webhook.CustomValidator = &OrderCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Order.
func (v *OrderCustomValidator) validateUniqueOrderID(ctx context.Context, resource *managev1.Order) error {
	var orderList managev1.OrderList
	if err := v.Client.List(ctx, &orderList); err != nil {
		return fmt.Errorf("failed to list Order resources: %v", err)
	}

	for _, existingResource := range orderList.Items {
		if existingResource.Spec.ID == resource.Spec.ID {
			return fmt.Errorf("order_id '%s' is already in use", resource.Spec.ID)
		}
	}
	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type StatusUpdate.
func (v *OrderCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {

	newResource, ok := obj.(*managev1.Order)
	if !ok {
		return nil, fmt.Errorf("expected a Order object but got %T", obj)
	}
	orderlog.Info("Validation for Order upon creation", "name", newResource.GetName())
	validationErr := v.validateUniqueOrderID(ctx, newResource)
	if validationErr != nil {
		return nil, validationErr
	}
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Order.
func (v *OrderCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldRes, ok := oldObj.(*managev1.Order)
	if !ok {
		return nil, fmt.Errorf("existing order object for the matching resource kind not found, got %T", oldRes)
	}
	newRes, ok := newObj.(*managev1.Order)
	if !ok {
		return nil, fmt.Errorf("expected a Order object for the newObj but got %T", newObj)
	}

	orderlog.Info("Validation for StatusUpdate upon update", "name", newRes.GetName())

	validationErr := v.validateUniqueOrderID(ctx, newRes)
	if validationErr != nil {
		return nil, validationErr
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Order.
func (v *OrderCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	order, ok := obj.(*managev1.Order)
	if !ok {
		return nil, fmt.Errorf("expected a Order object but got %T", obj)
	}
	orderlog.Info("Validation for Order upon deletion", "name", order.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrderCustomValidator) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&managev1.Order{}).
		WithValidator(r).
		Complete()
}
