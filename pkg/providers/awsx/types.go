//go:generate mockery --name IAM --name STS
package awsx

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

// IAM is the interface for the IAM API calls
type IAM interface {
	GetRole(*iam.GetRoleInput) (*iam.GetRoleOutput, error)
	CreateRole(*iam.CreateRoleInput) (*iam.CreateRoleOutput, error)
	DeleteRole(*iam.DeleteRoleInput) (*iam.DeleteRoleOutput, error)
	UpdateRole(*iam.UpdateRoleInput) (*iam.UpdateRoleOutput, error)
	UpdateRoleDescription(*iam.UpdateRoleDescriptionInput) (*iam.UpdateRoleDescriptionOutput, error)
	ListRolePoliciesPages(*iam.ListRolePoliciesInput, func(*iam.ListRolePoliciesOutput, bool) bool) error
	ListAttachedRolePoliciesPages(*iam.ListAttachedRolePoliciesInput, func(*iam.ListAttachedRolePoliciesOutput, bool) bool) error
	DeleteRolePolicy(*iam.DeleteRolePolicyInput) (*iam.DeleteRolePolicyOutput, error)
	DetachRolePolicy(*iam.DetachRolePolicyInput) (*iam.DetachRolePolicyOutput, error)
	UpdateAssumeRolePolicy(*iam.UpdateAssumeRolePolicyInput) (*iam.UpdateAssumeRolePolicyOutput, error)
	AttachRolePolicy(*iam.AttachRolePolicyInput) (*iam.AttachRolePolicyOutput, error)
	PutRolePolicy(*iam.PutRolePolicyInput) (*iam.PutRolePolicyOutput, error)
	PutRolePermissionsBoundary(input *iam.PutRolePermissionsBoundaryInput) (*iam.PutRolePermissionsBoundaryOutput, error)
	DeleteRolePermissionsBoundary(input *iam.DeleteRolePermissionsBoundaryInput) (*iam.DeleteRolePermissionsBoundaryOutput, error)
}

// STS is the interface for the STS API calls
type STS interface {
	GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)
}
