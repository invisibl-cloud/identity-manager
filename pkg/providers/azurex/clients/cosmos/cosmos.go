package cosmos

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/cosmos-db/mgmt/documentdb"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex"
	"github.com/invisibl-cloud/identity-manager/pkg/util"
)

type cosmosClientFactory func() CosmosClient

// Client is the cosmos client struct
type Client struct {
	Client *azurex.Client
	cosmosClientFactory
}

// New creates a new cosmos client
func New(p *azurex.Client) *Client {
	c := &Client{Client: p}
	c.cosmosClientFactory = c.newAccountsClient
	return c
}

func (x *Client) newAccountsClient() CosmosClient {
	dc := documentdb.NewDatabaseAccountsClient(x.Client.GetConfig().SubscriptionID)
	dc.Authorizer = x.Client.GetAuthorizer()
	return dc
}

// GetKey fetches accesskey for the cosmos account
func (x *Client) GetKey(ctx context.Context, cosmosAccountName string) (string, error) {
	if cosmosAccountName == "" {
		return "", fmt.Errorf("cosmos account name should not be empty")
	}
	ac := x.cosmosClientFactory()
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

// GetConnectionString builds connection string for the cosmos account
func (x *Client) GetConnectionString(ctx context.Context, cosmosAccountName string) (string, error) {
	ac := x.cosmosClientFactory()
	resp, err := ac.ListKeys(ctx, x.Client.GetConfig().ResourceGroup, cosmosAccountName)
	if err != nil {
		return "", err
	}
	key := util.DefaultString(*resp.PrimaryMasterKey, *resp.SecondaryMasterKey)
	connString := fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=cosmos.azure.com", cosmosAccountName, key)
	return connString, nil
}
