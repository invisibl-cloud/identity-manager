package iam

import (
	"context"
	"fmt"

	"cloud.google.com/go/iam"
	iamadminv1 "cloud.google.com/go/iam/admin/apiv1"
	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/gcpx"
	"github.com/invisibl-cloud/identity-manager/pkg/util"
	"google.golang.org/api/option"
)

// Client for IAM
type Client struct {
	Client   *gcpx.Client
	location string
	project  string
}

// New creates new iam client
func New(p *gcpx.Client) *Client {
	return &Client{
		Client:   p,
		location: p.GetConfig().Location,
		project:  p.GetConfig().Project,
	}
}

// EnsureServiceAccountWithRoles makes sure SA is created or updated for desired state
func (x *Client) EnsureServiceAccountWithRoles(ctx context.Context, name string, ns string, sas []*v1alpha1.ServiceAccount, displayName string, desc string, roles []string, scope string) (string, error) {
	accountID := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", name, x.project)
	err := x.createOrUpdateServiceAccount(ctx, accountID, name, ns, sas, displayName, desc)
	if err != nil {
		return accountID, err
	}
	err = x.ensureServiceAccountRoles(ctx, accountID, roles, scope)
	if err != nil {
		return accountID, err
	}
	return accountID, err
}

func (x *Client) createOrUpdateServiceAccount(ctx context.Context, accountID string, name string, ns string, sas []*v1alpha1.ServiceAccount, displayName string, desc string) error {
	rname := fmt.Sprintf("projects/%s/serviceAccounts/%s", x.project, accountID)
	iamSvc, err := iamadminv1.NewIamClient(ctx, option.WithCredentials(x.Client.GetCredentials()))
	if err != nil {
		return fmt.Errorf("error creating iam admin client - %w", err)
	}
	isNotFound := false
	obj, err := iamSvc.GetServiceAccount(ctx, &adminpb.GetServiceAccountRequest{
		Name: rname,
	})
	if err != nil {
		isNotFound = gcpx.IsNotFound(err)
		if !isNotFound {
			return fmt.Errorf("error getting sa %s - %w", accountID, err)
		}
	}
	if isNotFound {
		// accountId: must be 6-30 characters long, and match the regular expression `[a-z]([-a-z0-9]*[a-z0-9]
		_, err = iamSvc.CreateServiceAccount(ctx, &adminpb.CreateServiceAccountRequest{
			Name:      "projects/" + x.project,
			AccountId: name,
			ServiceAccount: &adminpb.ServiceAccount{
				DisplayName: displayName,
				Description: desc,
			},
		})
		if err != nil {
			return fmt.Errorf("error creating sa %s - %w", accountID, err)
		}
		return nil
	}
	// update
	if obj.DisplayName != displayName || obj.Description != desc {
		_, err = iamSvc.UpdateServiceAccount(ctx, &adminpb.ServiceAccount{
			Name:        rname,
			DisplayName: displayName,
			Description: desc,
		})
		if err != nil {
			return fmt.Errorf("error updating sa %s - %w", accountID, err)
		}
	}
	// TODO: better.
	if len(sas) == 0 {
		return nil
	}
	policy, err := iamSvc.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: rname,
	})
	if err != nil {
		return fmt.Errorf("error getting sa iam policy %s - %w", accountID, err)
	}
	if x.ensurePolicy(ctx, policy.InternalProto,
		fmt.Sprintf("serviceAccount:%s.svc.id.goog[%s/%s]", x.project, util.DefaultString(sas[0].Namespace, ns), sas[0].Name),
		[]string{"roles/iam.workloadIdentityUser"}) {
		_, err = iamSvc.SetIamPolicy(ctx, &iamadminv1.SetIamPolicyRequest{
			Resource: rname,
			Policy:   policy,
		})
		if err != nil {
			return fmt.Errorf("error setting sa iam policy %s - %w", accountID, err)
		}
	}
	return nil
}

func (x *Client) ensureServiceAccountRoles(ctx context.Context, saName string, roles []string, scope string) error {
	projClient, err := resourcemanager.NewProjectsClient(ctx, option.WithCredentials(x.Client.GetCredentials()))
	if err != nil {
		return fmt.Errorf("error creating new projects client %s - %w", saName, err)
	}
	if scope == "" {
		scope = "projects/" + x.project
	}
	policy, err := projClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: scope,
	})
	if err != nil {
		return fmt.Errorf("error getting project iam policy %s - %w", saName, err)
	}
	if x.ensurePolicy(ctx, policy, fmt.Sprintf("serviceAccount:%s", saName), roles) {
		_, err = projClient.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
			Resource: scope,
			Policy:   policy,
		})
		if err != nil {
			return fmt.Errorf("error setting project iam policy %s - %w", saName, err)
		}
	}
	return nil
}

func (x *Client) ensurePolicy(ctx context.Context, policy *iampb.Policy, member string, roles []string) bool {
	px := &iam.Policy{InternalProto: policy}
	existingRoles := []string{}
	for _, role := range px.Roles() {
		if px.HasRole(member, role) {
			existingRoles = append(existingRoles, string(role))
		}
	}
	d := util.FindSyncSteps(existingRoles, roles)
	for _, role := range d.Add {
		px.Add(member, iam.RoleName(role))
	}
	for _, role := range d.Delete {
		px.Remove(member, iam.RoleName(role))
	}
	return len(d.Add) > 0 || len(d.Delete) > 0
}

// DeleteServiceAccount deletes SA
func (x *Client) DeleteServiceAccount(ctx context.Context, accountID string) error {
	if accountID == "" {
		return nil
	}
	rname := fmt.Sprintf("projects/%s/serviceAccounts/%s", x.project, accountID)
	iamSvc, err := iamadminv1.NewIamClient(ctx, option.WithCredentials(x.Client.GetCredentials()))
	if err != nil {
		return err
	}
	err = iamSvc.DeleteServiceAccount(ctx, &adminpb.DeleteServiceAccountRequest{
		Name: rname,
	})
	if err != nil {
		if gcpx.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting sa %s - %w", accountID, err)
	}
	return fmt.Errorf("deleting %s", accountID)
}
