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

// WorkloadIdentitySpec defines the desired state of WorkloadIdentity
type WorkloadIdentitySpec struct {
	// Name of the WorkloadIdentity
	// +optional
	Name string `json:"name,omitempty"`
	// Desc of the WorkloadIdentity
	// +optional
	Description string `json:"description,omitempty"`
	// Credentials to manage the WorkloadIdentity
	// +optional
	Credentials *Credentials `json:"credentials,omitempty"`
	// Provider of the WorkloadIdentity
	// +kubebuilder:validation:Enum=AWS;Azure
	// +required
	Provider Provider `json:"provider"`
	// AWS WorkloadIdentity
	// +optional
	AWS *WorkloadIdentityAWS `json:"aws,omitempty"`
	// Azure WorkloadIdentity
	// +optional
	Azure *WorkloadIdentityAzure `json:"azure,omitempty"`
	// WriteToSecretRef is a reference to a secret
	// +optional
	WriteToSecretRef *WriteToSecretRef `json:"writeToSecretRef,omitempty"`
}

type WriteToSecretRef struct {
	// Name of the secret
	// +required
	Name string `json:"name"`
	// Namespace of the secret
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// TemplateData is a template for the data to be written to the secret
	// +required
	TemplateData map[string]string `json:"templateData"`
}

// Provider of the WorkloadIdentity
type Provider string

const (
	// ProviderAWS is the AWS provider.
	ProviderAWS Provider = "AWS"
	// ProviderAzure is the Azure provider.
	ProviderAzure Provider = "Azure"
)

// A CredentialsSource is a source from which provider credentials may be
// acquired.
type CredentialsSource string

const (
	// CredentialsSourceSecret indicates that a provider should acquire
	// credentials from a secret.
	CredentialsSourceSecret CredentialsSource = "Secret"
)

type Credentials struct {
	// Source of the credentials
	// +kubebuilder:validation:Enum=Secret
	// +optional
	Source CredentialsSource `json:"source,omitempty"`
	// SecretRef to fetch the credentials
	// +optional
	SecretRef *SecretRef `json:"secretRef,omitempty"`
	// Properties indicates extra properties of credentials
	// +optional
	Properties map[string]string `json:"properties,omitempty"`
}

// SecretRef - secret reference
type SecretRef struct {
	// Namespace of the secret.
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Name of the secret.
	// +required
	Name string `json:"name"`
}

type WorkloadIdentityAzure struct {
	// RoleDefinitions is a list of role definitions
	// +optional
	RoleDefinitions []*RoleDefinition `json:"roleDefinitions,omitempty"`
	// RoleAssignments of the WorkloadIdentity
	// +optional
	RoleAssignments map[string]RoleAssignment `json:"roleAssignments,omitempty"`
	// Identity of the WorkloadIdentity
	// +optional
	Identity *AzureIdentity `json:"identity,omitempty"`
	// IdentityBinding of the WorkloadIdentity
	// +optional
	IdentityBinding *AzureIdentityBinding `json:"identityBinding,omitempty"`
}

type RoleDefinition struct {
	// ID of the role definition (this will be used to generate internal UUID for role)
	// +required
	ID string `json:"id"`
	// RoleName of the role definition
	// +required
	RoleName string `json:"roleName"`
	// RoleType of the role definition
	// +required
	RoleType string `json:"roleType,omitempty"`
	// Description of the role definition
	// +optional
	Description string `json:"description,omitempty"`
	// AssignableScopes is a list of assignable scopes
	// +optional
	AssignableScopes []string `json:"assignableScopes,omitempty"`
	// Permissions of the role definition
	// +required
	Permissions []RolePermission `json:"permissions"`
}

type RolePermission struct {
	// Actions is a list of actions
	// +optional
	Actions []string `json:"actions,omitempty"`
	// NotActions is a list of not actions
	// +optional
	NotActions []string `json:"notActions,omitempty"`
	// DataActions is a list of data actions
	// +optional
	DataActions []string `json:"dataActions,omitempty"`
	// NotDataActions is a list of not data actions
	// +optional
	NotDataActions []string `json:"notDataActions,omitempty"`
}

type AzureIdentity struct {
	// APIVersion of the identity
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`
	// Kind of the identity
	// +optional
	Kind string `json:"kind,omitempty"`
	// Metadata of the identity
	// +optional
	Metadata *Metadata `json:"metadata,omitempty"`
	// Spec of the identity
	// +optional
	Spec *AzureIdentitySpec `json:"spec,omitempty"`
}

type AzureIdentityBinding struct {
	// APIVersion of the IdentityBinding
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`
	// Kind of the IdentityBinding
	// +optional
	Kind string `json:"kind,omitempty"`
	// Metadata of the IdentityBinding
	// +optional
	Metadata *Metadata `json:"metadata,omitempty"`
	// Spec of the IdentityBinding
	// +optional
	Spec *AzureIdentityBindingSpec `json:"spec,omitempty"`
}

type AzureIdentitySpec struct {
	// Type of the identity
	// +optional
	Type int `json:"type,omitempty"`
}

type AzureIdentityBindingSpec struct {
	// Selector of the IdentityBinding
	// +optional
	Selector string `json:"selector,omitempty"`
}

type RoleAssignment struct {
	// Role of the role assignment
	// +required
	Role string `json:"role"`
	// Scope of the role assignment
	// +optional
	Scope string `json:"scope,omitempty"`
}

type WorkloadIdentityAWS struct {
	// Path of the Role
	// +optional
	// +kubebuilder:default=/
	Path string `json:"path,omitempty"`
	// MaxSessionDuration of the Role
	// +optional
	MaxSessionDuration int64 `json:"maxSessionDuration,omitempty"`
	// AssumeRolePolicy of the Role
	// +required
	AssumeRolePolicy string `json:"assumeRolePolicy"`
	// InlinePolicies of the Role
	// +optional
	InlinePolicies map[string]string `json:"inlinePolicies,omitempty"`
	// Policies of the Role
	// +optional
	Policies []string `json:"policies,omitempty"`
	// ServiceAccounts to be managed
	// +optional
	ServiceAccounts []*ServiceAccount `json:"serviceAccounts,omitempty"`
	// Pods to be managed
	// +optional
	Pods []*AwsRoleSpecPod `json:"pods,omitempty"`
}

type AwsRoleSpecPod struct {
	metav1.LabelSelector `json:",inline"`
	// Namespace of the Pod
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// A ServiceAccountAction indicates action to be perform on ServiceAccount
type ServiceAccountAction string

const (
	// ServiceAccountActionCreate indicates create service account
	ServiceAccountActionCreate ServiceAccountAction = "Create"
	// ServiceAccountActionUpdate indicates updating service account
	ServiceAccountActionUpdate ServiceAccountAction = "Update"
	// ServiceAccountActionDefault indicates no action
	ServiceAccountActionDefault ServiceAccountAction = ""
)

// ServiceAccount to create service account.
type ServiceAccount struct {
	// Action to be perform on ServiceAccount
	// +kubebuilder:validation:Enum=Update;Create
	Action ServiceAccountAction `json:"action,omitempty"`
	// Name of the ServiceAccount
	// +optional
	Name string `json:"name,omitempty"`
	// Namespace of the ServiceAccount
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Annotations to be added on ServiceAccount
	// +optional
	Annotations map[string]string `json:"Annotations,omitempty"`
}

type Resource struct {
	// APIVersion of the resource
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`
	// Kind of the resource
	// +optional
	Kind string `json:"kind,omitempty"`
	// Name of the resource
	// +optional
	Name string `json:"name,omitempty"`
	// Namespace of the resource
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

type ExternalResource struct {
	// ID of the external resource
	// +optional
	ID string `json:"id,omitempty"`
	// Type of the external resource
	// +optional
	Type string `json:"type,omitempty"`
}

// WorkloadIdentityStatus defines the observed state of WorkloadIdentity
type WorkloadIdentityStatus struct {
	ConditionedStatus `json:",inline"`
	// ID of the Identity
	// +optional
	ID string `json:"id,omitempty"`
	// Name of the Identity
	// +optional
	Name string `json:"name,omitempty"`
	// Resources managed by the Identity
	// +optional
	Resources []Resource `json:"resources,omitempty"`
	// External Resources managed bu the Identity
	// +optional
	ExternalResources []ExternalResource `json:"externalResources,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// WorkloadIdentity is the Schema for the workloadidentities API
type WorkloadIdentity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkloadIdentitySpec   `json:"spec,omitempty"`
	Status WorkloadIdentityStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WorkloadIdentityList contains a list of WorkloadIdentity
type WorkloadIdentityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkloadIdentity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkloadIdentity{}, &WorkloadIdentityList{})
}

type Metadata struct {
	// Name of the Resource
	// +optional
	Name string `json:"name,omitempty"`
	// Namespace of the Resource
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Labels of the Resource
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations of the Resource
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

func (r *WorkloadIdentity) GetSpec() interface{} {
	return &r.Spec
}

func (r *WorkloadIdentity) GetStatus() interface{} {
	return &r.Status
}

func (r *WorkloadIdentity) GetSpecCopy() interface{} {
	return r.Spec.DeepCopy()
}

func (r *WorkloadIdentity) GetStatusCopy() interface{} {
	return r.Status.DeepCopy()
}

func (r *WorkloadIdentity) GetConditionedStatus() *ConditionedStatus {
	return &r.Status.ConditionedStatus
}
