package cosmos

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/cosmos-db/mgmt/documentdb"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex"
)

type Client struct {
	Client *azurex.Client
}

func New(p *azurex.Client) *Client {
	c := &Client{Client: p}
	return c
}

func (x *Client) NewAccountsClient() documentdb.DatabaseAccountsClient {
	dc := documentdb.NewDatabaseAccountsClient(x.Client.GetConfig().SubscriptionID)
	dc.Authorizer = x.Client.GetAuthorizer()
	return dc
}

func (x *Client) GetKey(ctx context.Context, cosmosAccountName string) (string, error) {
	if cosmosAccountName == "" {
		return "", fmt.Errorf("cosmos account name should not be empty")
	}
	ac := x.NewAccountsClient()
	resp, err := ac.ListKeys(ctx, x.Client.GetConfig().ResourceGroup, cosmosAccountName)
	if err != nil {
		return "", err
	}
	if resp.PrimaryMasterKey != nil && resp.SecondaryMasterKey == nil {
		return "", fmt.Errorf("no keys found cosmos account: %s", cosmosAccountName)
	}
	if resp.PrimaryMasterKey != nil {
		return *resp.PrimaryMasterKey, nil
	}
	if resp.SecondaryMasterKey != nil {
		return *resp.SecondaryMasterKey, nil
	}
	return "", nil
}
