package msi

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/msi/mgmt/msi"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex"
)

// Client is the definition of the MSI client
type Client struct {
	*azurex.Client
	resourceGroup string
	location      string
}

// New expects *azurex.Client and returns *msi.Client
func New(p *azurex.Client) *Client {
	return &Client{
		Client:        p,
		resourceGroup: p.GetConfig().ResourceGroup,
		location:      p.GetConfig().Location,
	}
}

func getUserAssignedIdentitiesClient(p *azurex.Client) (msi.UserAssignedIdentitiesClient, error) {
	c := msi.NewUserAssignedIdentitiesClient(p.GetConfig().SubscriptionID)
	c.Authorizer = p.GetAuthorizer()
	err := c.AddToUserAgent(azurex.UserAgent)
	if err != nil {
		return msi.UserAssignedIdentitiesClient{}, err
	}
	return c, nil
}

// Identity is the definition of the Identity struct
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

// CreateOrUpdate performs creation of updation of identities
func (c *Client) CreateOrUpdate(ctx context.Context, resourceName string, tags map[string]*string) (*Identity, error) {
	uai, err := getUserAssignedIdentitiesClient(c.Client)
	if err != nil {
		return nil, err
	}

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

// EnsureDelete ensures deletion of the identity
func (c *Client) EnsureDelete(ctx context.Context, resourceName string) (bool, error) {
	uai, err := getUserAssignedIdentitiesClient(c.Client)
	if err != nil {
		return false, err
	}

	_, err = uai.Get(ctx, c.resourceGroup, resourceName)
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
