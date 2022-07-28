package accounts

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
)

type AccountsClient interface {
	ListKeys(ctx context.Context, resourceGroupName string, accountName string, expand storage.ListKeyExpand) (result storage.AccountListKeysResult, err error)
}
