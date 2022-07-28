package cosmos

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/cosmos-db/mgmt/documentdb"
)

type CosmosClient interface {
	ListKeys(ctx context.Context, resourceGroupName string, accountName string) (result documentdb.DatabaseAccountListKeysResult, err error)
}
