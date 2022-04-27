package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	v1alpha1 "github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/awsx"
	"github.com/invisibl-cloud/identity-manager/pkg/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

// RoleStatus returns iam role information
type RoleStatus struct {
	Name      string
	ARN       string
	NeedsSync bool
}

func New(config *awsx.Config, role *v1alpha1.WorkloadIdentity) *Client {
	return &Client{
		config: config,
		role:   role,
	}
}

// Client provides an interface with interacting with AWS
type Client struct {
	config    *awsx.Config
	iam       *iam.IAM
	sts       *sts.STS
	role      *v1alpha1.WorkloadIdentity
	accountID string
}

func (i *Client) internalRoleName() string {
	meta := i.role.ObjectMeta
	return meta.GetNamespace() + "-" + meta.GetName()
}

func (i *Client) roleName() string {
	roleName := i.role.Spec.Name
	if roleName == "" {
		roleName = i.internalRoleName()
	}
	if strings.HasSuffix(roleName, "-") { // prefix based
		roleName = roleName + i.internalRoleName()
	}
	//if strings.Contains(roleName, "{{") { // pattern based
	// TODO:
	//}
	// max 64 char
	if len(roleName) > 64 {
		return roleName[:64]
	}
	return roleName
}

func (i *Client) Prepare(ctx context.Context) error {
	sess, err := awsx.NewSession(*i.config)
	if err != nil {
		return err
	}

	i.sts = sts.New(sess)
	callerIdentity, err := i.sts.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}
	i.accountID = aws.StringValue(callerIdentity.Account)
	i.iam = iam.New(sess)
	return nil
}

// Reconcile IAM Role
func (i *Client) CreateOrUpdate(ctx context.Context) (*RoleStatus, error) {
	prevRoleName := i.role.Status.Name
	newRoleName := i.roleName()
	if prevRoleName != "" && prevRoleName != newRoleName {
		// i.log.Info("Role name changed!", "prevRole", prevRoleName, "newRole", newRoleName)
		if i.isExists(prevRoleName) {
			// i.log.Info("Found previous role! Deleting it!", "prevRole", prevRoleName)
			err := i.delete(prevRoleName)
			if err != nil {
				// i.log.Error(err, "Error deleting the role!", "prevRole", prevRoleName)
				return nil, err
			}
		} else {
			// i.log.Info("Previous role not found! Ignoring it!", "prevRole", prevRoleName)
		}
	}
	status, err := i.createOrSync(newRoleName)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// Finalize will delete iam role.
func (i *Client) Delete(ctx context.Context) error {
	roleName := i.role.Status.Name
	if roleName != "" && i.isExists(roleName) {
		// i.log.Info("Found iam role! Deleting it!", "role", roleName)
		err := i.delete(roleName)
		if err != nil {
			// i.log.Error(err, "Error deleting the role!", "prevRole", roleName)
			return err
		}
	}
	return nil
}

func (i *Client) createOrSync(roleName string) (*RoleStatus, error) {
	if i.isExists(roleName) {
		// i.log.Info("Found existing role! Syncing it!", "role", roleName)
		status, err := i.sync(roleName)
		if err != nil {
			// i.log.Error(err, "Error syncing the role!", "role", roleName)
			return nil, err
		}
		return status, nil
	}
	// i.log.Info("Creating the role!", "role", roleName)
	status, err := i.create(roleName)
	if err != nil {
		// i.log.Error(err, "Error creating the role!", "role", roleName)
		return nil, err
	}
	return status, nil
}

//Create creates an IAM role in AWS, based on a spec
func (i *Client) create(roleName string) (*RoleStatus, error) {
	input := &iam.CreateRoleInput{RoleName: &roleName}
	input.AssumeRolePolicyDocument = &i.role.Spec.AWS.AssumeRolePolicy // required
	if i.role.Spec.Description != "" {
		input.Description = &i.role.Spec.Description
	}
	if i.role.Spec.AWS.Path != "" {
		input.Path = &i.role.Spec.AWS.Path
	}
	if i.role.Spec.AWS.MaxSessionDuration > 0 {
		input.MaxSessionDuration = &i.role.Spec.AWS.MaxSessionDuration
	}
	// input.Tags
	createRoleOutput, err := i.iam.CreateRole(input)
	if err != nil {
		return nil, err
	}
	err = i.createInlinePolicies()
	if err != nil {
		return nil, err
	}
	err = i.attachPolicies()
	if err != nil {
		return nil, err
	}
	return &RoleStatus{Name: *createRoleOutput.Role.RoleName, ARN: *createRoleOutput.Role.Arn}, nil
}

//Delete deletes an IAM role
func (i *Client) delete(roleName string) error {
	currentPolicies, err := i.listInlinePolicies(roleName)
	if err != nil {
		return err
	}
	for _, policy := range currentPolicies {
		_, err = i.iam.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
			PolicyName: &policy,
			RoleName:   &roleName,
		})
		if err != nil {
			return err
		}
	}
	attachedPolicies, err := i.listAttachedPolicies(roleName)
	if err != nil {
		return err
	}
	for _, policy := range attachedPolicies {
		_, err = i.iam.DetachRolePolicy(&iam.DetachRolePolicyInput{
			PolicyArn: policy.PolicyArn,
			RoleName:  &roleName,
		})
		if err != nil {
			return err
		}
	}
	_, err = i.iam.DeleteRole(&iam.DeleteRoleInput{
		RoleName: &roleName,
	})

	return err
}

//Sync synchronizes an AWS IAM Role to a spec
func (i *Client) sync(roleName string) (*RoleStatus, error) {
	getRoleOutput, err := i.iam.GetRole(&iam.GetRoleInput{
		RoleName: &roleName,
	})
	if err != nil {
		return nil, err
	}
	awsRole := getRoleOutput.Role

	// sync assume role / trustrelationship
	existingAssumeRolePolicy, err := url.PathUnescape(aws.StringValue(awsRole.AssumeRolePolicyDocument))
	if err != nil {
		return nil, err
	}
	existingAssumeRolePolicy = toCompactJSON(existingAssumeRolePolicy)
	currentAssumeRolePolicy := toCompactJSON(i.role.Spec.AWS.AssumeRolePolicy)
	if existingAssumeRolePolicy != currentAssumeRolePolicy {
		_, err = i.iam.UpdateAssumeRolePolicy(&iam.UpdateAssumeRolePolicyInput{
			RoleName:       &roleName,
			PolicyDocument: &i.role.Spec.AWS.AssumeRolePolicy,
		})
		if err != nil {
			return nil, err
		}
	}

	// sync inline policy
	existingInlinePolicyNames, err := i.listInlinePolicies(roleName)
	if err != nil {
		return nil, err
	}
	inlinePolicyNames, inlinePolicyNameMapping := toInlinePolicyNames(i.role.Spec.AWS.InlinePolicies)
	syncSteps := util.FindSyncSteps(existingInlinePolicyNames, inlinePolicyNames)
	//// i.log.Info("Sync inline policies", "syncSteps", syncSteps)
	for _, policyName := range syncSteps.Add {
		err = i.createInlinePolicy(roleName, policyName, i.role.Spec.AWS.InlinePolicies[inlinePolicyNameMapping[policyName]])
		if err != nil {
			return nil, err
		}
	}
	for _, policyName := range syncSteps.Delete {
		err = i.deleteInlinePolicy(roleName, policyName)
		if err != nil {
			return nil, err
		}
	}

	// sync policy arns
	attachedPolicies, err := i.listAttachedPolicies(roleName)
	if err != nil {
		return nil, err
	}
	syncSteps = util.FindSyncSteps(toArns(attachedPolicies), i.toArns(i.role.Spec.AWS.Policies))
	for _, policyArn := range syncSteps.Add {
		err = i.attachPolicy(roleName, policyArn)
		if err != nil {
			return nil, err
		}
	}
	for _, policyArn := range syncSteps.Delete {
		err = i.detachPolicy(roleName, policyArn)
		if err != nil {
			return nil, err
		}
	}

	// sync max-session duration.
	if i.role.Spec.AWS.MaxSessionDuration > 0 && aws.Int64Value(awsRole.MaxSessionDuration) != i.role.Spec.AWS.MaxSessionDuration {
		_, err = i.iam.UpdateRole(&iam.UpdateRoleInput{
			RoleName:           &roleName,
			MaxSessionDuration: &i.role.Spec.AWS.MaxSessionDuration,
		})
		if err != nil {
			return nil, err
		}
	}

	// sync description
	if i.role.Spec.Description != "" && aws.StringValue(awsRole.Description) != i.role.Spec.Description {
		_, err = i.iam.UpdateRoleDescription(&iam.UpdateRoleDescriptionInput{
			Description: &i.role.Spec.Description,
			RoleName:    &roleName,
		})
		if err != nil {
			return nil, err
		}
	}
	return &RoleStatus{Name: roleName, ARN: aws.StringValue(awsRole.Arn)}, nil
}

func toCompactJSON(txt string) string {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, []byte(txt)); err != nil {
		return txt
	}
	return buffer.String()
}

func toInlinePolicyNames(pols map[string]string) ([]string, map[string]string) {
	m := map[string]string{}
	names := make([]string, len(pols))
	i := 0
	for pname, pval := range pols {
		name := pname
		// max length: 128
		if len(name) < 128 {
			name = name + "-" + util.MD5(pval)
		}
		if len(name) > 128 {
			name = name[:128]
		}
		names[i] = name
		m[name] = pname
		i++
	}
	return names, m
}

func toArns(pols []*iam.AttachedPolicy) []string {
	arns := make([]string, len(pols))
	for ix, pol := range pols {
		arns[ix] = *pol.PolicyArn
	}
	return arns
}

func (i *Client) toArns(pols []string) []string {
	arns := make([]string, len(pols))
	for ix, pol := range pols {
		arn, _ := i.getArn(pol)
		arns[ix] = arn
	}
	return arns
}

// Exists Checks to see if a named IAM Role exists in AWS
func (i *Client) isExists(roleName string) bool {
	_, err := i.iam.GetRole(&iam.GetRoleInput{
		RoleName: &roleName,
	})
	return err == nil
}

// Attaches policies found in the spec to a named IAM role
func (i *Client) attachPolicies() error {
	if i.role.Spec.AWS.Policies == nil || len(i.role.Spec.AWS.Policies) == 0 {
		return nil
	}
	roleName := i.roleName()
	for _, policy := range i.role.Spec.AWS.Policies {
		policyArn, err := i.getArn(policy)
		if err != nil {
			return err
		}
		err = i.attachPolicy(roleName, policyArn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Client) attachPolicy(roleName string, policyArn string) error {
	_, err := i.iam.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: &policyArn,
		RoleName:  &roleName,
	})
	if err != nil {
		return err
	}
	return nil
}

func (i *Client) detachPolicy(roleName string, policyArn string) error {
	_, err := i.iam.DetachRolePolicy(&iam.DetachRolePolicyInput{
		PolicyArn: &policyArn,
		RoleName:  &roleName,
	})
	if err != nil {
		return err
	}
	return nil
}

// Creates inline polices defined in a spec and attaches it to a role
func (i *Client) createInlinePolicies() error {
	if i.role.Spec.AWS.InlinePolicies == nil || len(i.role.Spec.AWS.InlinePolicies) == 0 {
		return nil
	}
	roleName := i.roleName()
	for policyName, policy := range i.role.Spec.AWS.InlinePolicies {
		err := i.createInlinePolicy(roleName, policyName, policy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Client) createInlinePolicy(roleName, policyName, policy string) error {
	policy, err := i.resolvePolicy(policy)
	if err != nil {
		return err
	}
	_, err = i.iam.PutRolePolicy(&iam.PutRolePolicyInput{
		RoleName:       &roleName,
		PolicyName:     &policyName,
		PolicyDocument: &policy,
	})
	if err != nil {
		return err
	}
	return nil
}

func (i *Client) deleteInlinePolicy(roleName, policyName string) error {
	_, err := i.iam.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
		PolicyName: &policyName,
		RoleName:   &roleName,
	})
	if err != nil {
		// check if already deleted
		_, ok := awsx.CheckError(err, iam.ErrCodeNoSuchEntityException)
		if ok {
			return nil
		}
		return err
	}
	return nil
}

func (i *Client) resolvePolicy(policy string) (string, error) {
	//if strings.Contains(policy, "://") {
	// TODO: !!
	//}
	return policy, nil
}

// Returns the ARN of a policy; allows for simply naming policies
func (i *Client) getArn(policyName string) (string, error) {
	if isArn(policyName) {
		return policyName, nil
	}
	if i.accountID == "" {
		callerIdentity, err := i.sts.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			return "", err
		}
		i.accountID = aws.StringValue(callerIdentity.Account)
	}
	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", i.accountID, policyName), nil
}

// Returns if a policy string is an ARN
func isArn(policy string) bool {
	return strings.Contains(policy, "arn:aws:iam")
}

// Paginate over inline policies
func (i *Client) listInlinePolicies(roleName string) ([]string, error) {
	var policyNamesPointers []*string
	err := i.iam.ListRolePoliciesPages(&iam.ListRolePoliciesInput{RoleName: &roleName}, func(currentPolicies *iam.ListRolePoliciesOutput, lastPage bool) bool {
		policyNamesPointers = append(policyNamesPointers, currentPolicies.PolicyNames...)
		return true
	})
	if err != nil {
		return nil, err
	}
	return aws.StringValueSlice(policyNamesPointers), nil
}

// Paginate over attached policies
func (i *Client) listAttachedPolicies(roleName string) ([]*iam.AttachedPolicy, error) {
	var policyPointers []*iam.AttachedPolicy
	err := i.iam.ListAttachedRolePoliciesPages(&iam.ListAttachedRolePoliciesInput{RoleName: &roleName}, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
		policyPointers = append(policyPointers, page.AttachedPolicies...)
		return true
	})
	if err != nil {
		return nil, err
	}
	return policyPointers, nil
}
