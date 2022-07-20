package accounts

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex"
)

type Client struct {
	Client *azurex.Client
}

func New(p *azurex.Client) *Client {
	c := &Client{Client: p}
	return c
}

func (x *Client) NewAccountsClient() storage.AccountsClient {
	sc := storage.NewAccountsClient(x.Client.GetConfig().SubscriptionID)
	sc.Authorizer = x.Client.GetAuthorizer()
	return sc
}

func (x *Client) GetKey(ctx context.Context, storageAccountName string) (string, error) {
	if storageAccountName == "" {
		return "", fmt.Errorf("storage account name should not be empty")
	}
	ac := storage.NewAccountsClient(x.Client.GetConfig().SubscriptionID)
	resp, err := ac.ListKeys(ctx, x.Client.GetConfig().ResourceGroup, storageAccountName, storage.ListKeyExpandKerb)
	if err != nil {
		return "", err
	}
	if resp.Keys == nil {
		return "", fmt.Errorf("no keys found for account: %s", storageAccountName)
	}
	for _, v := range *resp.Keys {
		if v.KeyName == nil || v.Value == nil {
			continue
		}
		return *v.Value, nil
	}
	return "", nil
}
