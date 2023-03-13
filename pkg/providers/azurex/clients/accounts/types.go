//go:generate mockery --name AccountsClient
package accounts

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
)

// AccountsClient is the mock interface for accounts client
type AccountsClient interface {
	ListKeys(ctx context.Context, resourceGroupName string, accountName string, expand storage.ListKeyExpand) (result storage.AccountListKeysResult, err error)
}
