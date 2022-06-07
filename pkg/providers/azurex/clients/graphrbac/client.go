package graphrbac

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	//"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"

	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex"
)

var uuidRE = regexp.MustCompile("^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$")

// Client is the RBAC client definition
type Client struct {
	*azurex.Client
	resourceGroup string
	location      string
}

// New returns the RBAC client
func New(p *azurex.Client) *Client {
	return &Client{
		Client:        p,
		resourceGroup: p.GetConfig().ResourceGroup,
		location:      p.GetConfig().Location,
	}
}

func getRoleDefinitionsClient(p *azurex.Client) (authorization.RoleDefinitionsClient, error) {
	c := authorization.NewRoleDefinitionsClient(p.GetConfig().SubscriptionID)
	c.Authorizer = p.GetAuthorizer()
	err := c.AddToUserAgent(azurex.UserAgent)
	if err != nil {
		return authorization.RoleDefinitionsClient{}, err
	}
	return c, nil
}

func getRoleAssignmentsClient(p *azurex.Client) (authorization.RoleAssignmentsClient, error) {
	c := authorization.NewRoleAssignmentsClient(p.GetConfig().SubscriptionID)
	c.Authorizer = p.GetAuthorizer()
	err := c.AddToUserAgent(azurex.UserAgent)
	if err != nil {
		return authorization.RoleAssignmentsClient{}, err
	}
	return c, nil
}

// CreateOrUpdateRoleDefinition creates or updates the role definition
func (c Client) CreateOrUpdateRoleDefinition(ctx context.Context, id string, scope string, prop authorization.RoleDefinitionProperties) error {
	rdc, err := getRoleDefinitionsClient(c.Client)
	if err != nil {
		return err
	}
	scope, err = c.ensureScope(scope)
	if err != nil {
		return err
	}
	_, err = rdc.CreateOrUpdate(ctx, scope, id, authorization.RoleDefinition{
		RoleDefinitionProperties: &prop,
	})
	return err
}

// DeleteRoleDefinition deletes role definition
func (c Client) DeleteRoleDefinition(ctx context.Context, scope, id string) error {
	rdc, err := getRoleDefinitionsClient(c.Client)
	if err != nil {
		return err
	}
	scope, err = c.ensureScope(scope)
	if err != nil {
		return err
	}
	_, err = rdc.Delete(ctx, scope, id)
	return err
}

// ListRoleAssignments gets all role assignments for the principal
func (c Client) ListRoleAssignments(ctx context.Context, principalID string) ([]*authorization.RoleAssignment, error) {
	rac, err := getRoleAssignmentsClient(c.Client)
	if err != nil {
		return nil, err
	}

	list := []*authorization.RoleAssignment{}
	filter := fmt.Sprintf("principalId eq '%s'", principalID)
	for l, err := rac.ListComplete(ctx, filter); l.NotDone(); err = l.NextWithContext(ctx) {
		if err != nil {
			return nil, err
		}
		v := l.Value()
		list = append(list, &v)
	}
	return list, nil
}

// GetRoleDefintionIDFromName gets the role definition ID from the role name
func (c Client) GetRoleDefintionIDFromName(ctx context.Context, name string, scope string) (string, error) {
	if strings.HasPrefix(name, "/") {
		return name, nil
	}
	if uuidRE.MatchString(name) {
		return fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s", c.GetConfig().SubscriptionID, name), nil
	}
	scope, err := c.ensureScope(scope)
	if err != nil {
		return "", err
	}
	rdc, err := getRoleDefinitionsClient(c.Client)
	if err != nil {
		return "", err
	}
	l, err := rdc.List(ctx, scope, fmt.Sprintf("roleName eq '%s'", name))
	if err != nil {
		return "", err
	}
	items := l.Values()
	switch len(items) {
	case 0:
		return "", fmt.Errorf("role %s not found", name)
	case 1:
		return to.String(items[0].ID), nil
	default:
		return "", fmt.Errorf("found multiple role definitions with name %q", name)
	}
}

// DeleteRoleAssignment deletes the role assingnment
func (c Client) DeleteRoleAssignment(ctx context.Context, id string) error {
	rac, err := getRoleAssignmentsClient(c.Client)
	if err != nil {
		return err
	}

	_, err = rac.DeleteByID(ctx, id)
	return err
}

// CreateRoleAssignment creates role assignment
func (c Client) CreateRoleAssignment(ctx context.Context, id string, principalID, roleDefinitionID, scope string) (string, error) {
	scope, err := c.ensureScope(scope)
	if err != nil {
		return "", err
	}
	rac, err := getRoleAssignmentsClient(c.Client)
	if err != nil {
		return "", err
	}

	p := authorization.RoleAssignmentCreateParameters{
		RoleAssignmentProperties: &authorization.RoleAssignmentProperties{
			RoleDefinitionID: azurex.ToStringPtr(roleDefinitionID),
			PrincipalID:      azurex.ToStringPtr(principalID),
		},
	}
	pr, err := rac.Create(ctx, scope, id, p)
	if err != nil {
		return "", err
	}
	return to.String(pr.ID), nil
}

func (c Client) ensureScope(scope string) (string, error) {
	if scope == "" {
		return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", c.Client.GetConfig().SubscriptionID, c.resourceGroup), nil
	}
	return c.ensureResourceID(scope)
}

func (c Client) ensureResourceID(id string) (string, error) {
	if strings.HasPrefix(id, "/") {
		return id, nil
	}
	sb := strings.Builder{}
	if !strings.HasPrefix(id, "/subscriptions/") {
		sb.WriteString("/subscriptions/")
		sb.WriteString(c.Client.GetConfig().SubscriptionID)
	}
	if !strings.Contains(id, "/resourceGroups/") {
		sb.WriteString("/resourceGroups/")
		sb.WriteString(c.resourceGroup)
	}
	sb.WriteString("/")
	if strings.Contains(id, "@") {
		parts := strings.Split(id, "@")
		switch parts[1] {
		case "dnszones":
			sb.WriteString("providers/Microsoft.Network/dnszones/")
		default:
			return "", fmt.Errorf("unsupported alias %s", parts[1])
		}
		sb.WriteString(parts[0])
	} else {
		sb.WriteString(id)
	}
	return sb.String(), nil
}
