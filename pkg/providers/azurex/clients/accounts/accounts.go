package accounts

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex"
)

type accountsClientFactory func() AccountsClient

// Client is the accounts client struct
type Client struct {
	Client *azurex.Client
	accountsClientFactory
}

// New creates a new accounts client
func New(p *azurex.Client) *Client {
	c := &Client{Client: p}
	c.accountsClientFactory = c.newAccountsClient
	return c
}

func (x *Client) newAccountsClient() AccountsClient {
	sc := storage.NewAccountsClient(x.Client.GetConfig().SubscriptionID)
	sc.Authorizer = x.Client.GetAuthorizer()
	return sc
}

// GetKey fetches access key for the storage account
func (x *Client) GetKey(ctx context.Context, storageAccountName string) (string, error) {
	if storageAccountName == "" {
		return "", fmt.Errorf("storage account name should not be empty")
	}
	ac := x.accountsClientFactory()
	resp, err := ac.ListKeys(ctx, x.Client.GetConfig().ResourceGroup, storageAccountName, storage.ListKeyExpandKerb)
	if err != nil {
		return "", err
	}
	if resp.Keys == nil {
		return "", fmt.Errorf("no keys found for account: %s", storageAccountName)
	}
	var key string
	for _, v := range *resp.Keys {
		if v.KeyName == nil || v.Value == nil {
			continue
		}
		key = *v.Value
	}
	return key, nil
}
