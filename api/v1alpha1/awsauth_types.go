/*
Copyright 2022.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AWSAuthSpec defines the desired state of AWSAuth
type AWSAuthSpec struct {
	// MapRoles holds a list of MapRoleItem
	//+kubebuilder:validation:Optional
	MapRoles []MapRoleItem `json:"mapRoles,omitempty" yaml:"mapRoles,omitempty"`

	// MapUsers holds a list of MapUserItem
	//+kubebuilder:validation:Optional
	MapUsers []MapUserItem `json:"mapUsers,omitempty" yaml:"mapUsers,omitempty"`
}

// MapRoleItem defines the mapRole item of AWSAuth
type MapRoleItem struct {
	// The ARN of the IAM role to add
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=25
	RoleArn string `json:"rolearn" yaml:"rolearn"`

	// The user name within Kubernetes to map to the IAM role
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=1
	Username string `json:"username" yaml:"username"`

	// A list of groups within Kubernetes to which the role is mapped
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinItems=1
	Groups []string `json:"groups" yaml:"groups"`
}

// MapUserItem defines the mapUser item of AWSAuth
type MapUserItem struct {
	// The ARN of the IAM user to add
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=25
	UserArn string `json:"userarn" yaml:"userarn"`

	// The user name within Kubernetes to map to the IAM user
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=1
	Username string `json:"username" yaml:"username"`

	// A list of groups within Kubernetes to which the user is mapped to
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinItems=1
	Groups []string `json:"groups" yaml:"groups"`
}

// AWSAuthStatus defines the observed state of AWSAuth
type AWSAuthStatus struct {
	ConditionedStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSAuth is the Schema for the awsauths API
type AWSAuth struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSAuthSpec   `json:"spec,omitempty"`
	Status AWSAuthStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSAuthList contains a list of AWSAuth
type AWSAuthList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSAuth `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSAuth{}, &AWSAuthList{})
}

// GetSpec returns spec of AWSAuth
func (r *AWSAuth) GetSpec() any {
	return &r.Spec
}

// GetStatus returns status of AWSAuth
func (r *AWSAuth) GetStatus() any {
	return &r.Status
}

// GetSpecCopy returns spec's copy of AWSAuth
func (r *AWSAuth) GetSpecCopy() any {
	return r.Spec.DeepCopy()
}

// GetStatusCopy returns status's copy of AWSAuth
func (r *AWSAuth) GetStatusCopy() any {
	return r.Status.DeepCopy()
}

// GetConditionedStatus returns condition status of AWSAuth
func (r *AWSAuth) GetConditionedStatus() *ConditionedStatus {
	return &r.Status.ConditionedStatus
}
