package accounts

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/Azure/go-autorest/autorest/to"
	mocks "github.com/invisibl-cloud/identity-manager/pkg/mocks/azure"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex"
	"github.com/stretchr/testify/assert"
)

var keysList = []storage.AccountKey{
	{
		KeyName: to.StringPtr("key1"),
		Value:   to.StringPtr("val1"),
	},
}

func TestAccountsClient_GetKey_Accuracy(t *testing.T) {
	ctx := context.Background()
	mc := &mocks.AccountsClient{}
	mc.AssertExpectations(t)
	mc.On("ListKeys", ctx, "", "test1", storage.ListKeyExpandKerb).Return(storage.AccountListKeysResult{Keys: &keysList}, nil, nil)

	x, _ := azurex.New(azurex.WithEnv())
	c := &Client{
		Client: x,
		accountsClientFactory: func() AccountsClient {
			return mc
		},
	}

	key, err := c.GetKey(ctx, "test1")

	assert.Nil(t, err)
	assert.Equal(t, key, "val1")
}
