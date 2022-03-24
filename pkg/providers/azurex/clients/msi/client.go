package msi

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/msi/mgmt/msi"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex"
)

type Client struct {
	*azurex.Client
	resourceGroup string
	location      string
}

func New(p *azurex.Client) *Client {
	return &Client{
		Client:        p,
		resourceGroup: p.GetConfig().ResourceGroup,
		location:      p.GetConfig().Location,
	}
}

func getUserAssignedIdentitiesClient(p *azurex.Client) msi.UserAssignedIdentitiesClient {
	c := msi.NewUserAssignedIdentitiesClient(p.GetConfig().SubscriptionID)
	c.Authorizer = p.GetAuthorizer()
	c.AddToUserAgent(azurex.UserAgent)
	return c
}

type Identity struct {
	ID             string
	Name           string
	Type           string
	Location       string
	SubscriptionID string
	ResourceGroup  string
	TenantID       string
	PrincipalID    string
	ClientID       string
}

func (c *Client) CreateOrUpdate(ctx context.Context, resourceName string, tags map[string]*string) (*Identity, error) {
	uai := getUserAssignedIdentitiesClient(c.Client)
	id, err := uai.CreateOrUpdate(ctx, c.resourceGroup, resourceName, msi.Identity{
		Location: &c.location,
		Tags:     tags,
	})
	if err != nil {
		return nil, err
	}
	return &Identity{
		Name:           azurex.ToString(id.Name),
		ID:             azurex.ToString(id.ID),
		Type:           azurex.ToString(id.Type),
		TenantID:       azurex.ToUUIDString(id.TenantID),
		PrincipalID:    azurex.ToUUIDString(id.PrincipalID),
		ClientID:       azurex.ToUUIDString(id.ClientID),
		Location:       c.location,
		ResourceGroup:  c.resourceGroup,
		SubscriptionID: c.Client.GetConfig().SubscriptionID,
	}, nil
}

func (c *Client) EnsureDelete(ctx context.Context, resourceName string) (bool, error) {
	uai := getUserAssignedIdentitiesClient(c.Client)
	_, err := uai.Get(ctx, c.resourceGroup, resourceName)
	if err != nil {
		if azurex.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}
	_, err = uai.Delete(ctx, c.resourceGroup, resourceName)
	if err != nil {
		return false, err
	}
	return false, nil
}
