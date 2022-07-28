package cosmos

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/cosmos-db/mgmt/documentdb"
	mocks "github.com/invisibl-cloud/identity-manager/pkg/mocks/azure"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex"
	"github.com/stretchr/testify/assert"
)

var resp = documentdb.DatabaseAccountListKeysResult{
	PrimaryMasterKey:   azurex.ToStringPtr("pk1"),
	SecondaryMasterKey: azurex.ToStringPtr("sk1"),
}

func TestCosmosClient_GetKey_Accuracy(t *testing.T) {
	ctx := context.Background()
	mc := &mocks.CosmosClient{}
	mc.AssertExpectations(t)
	mc.On("ListKeys", ctx, "", "test1").Return(resp, nil, nil)

	x, _ := azurex.New(azurex.WithEnv())
	c := &Client{
		Client: x,
		CosmosClientFactory: func() CosmosClient {
			return mc
		},
	}

	key, err := c.GetKey(ctx, "test1")

	assert.Nil(t, err)
	assert.Equal(t, key, "pk1")
}

func TestCosmosClient_ConnectionString_Accuracy(t *testing.T) {
	ctx := context.Background()
	mc := &mocks.CosmosClient{}
	mc.AssertExpectations(t)
	mc.On("ListKeys", ctx, "", "test1").Return(resp, nil, nil)

	x, _ := azurex.New(azurex.WithEnv())
	c := &Client{
		Client: x,
		CosmosClientFactory: func() CosmosClient {
			return mc
		},
	}

	key, err := c.GetConnectionString(ctx, "test1")

	assert.Nil(t, err)
	assert.Equal(t, key, fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=cosmos.azure.com", "test1", "pk1"))
}
