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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OrderSpec defines the desired state of Order.
type OrderSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ID...
	// +kubebuilder:validation:Required
	ID string `json:"order_id"`
	// Inventory...
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=10
	Inventory string `json:"inventory"`
	// Quantity defines the number of replicas (Default: 1, Min: 1, Max: 10)
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	Quantity int    `json:"quantity"`
	Seller   string `json:"seller,omitempty"`
}

// OrderStatus defines the observed state of Order.
type OrderStatus struct {
	// State
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Order is the Schema for the orders API.
type Order struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrderSpec   `json:"spec,omitempty"`
	Status OrderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OrderList contains a list of Order.
type OrderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Order `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Order{}, &OrderList{})
}
