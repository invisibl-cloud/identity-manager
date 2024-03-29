// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	context "context"

	storage "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"
	mock "github.com/stretchr/testify/mock"
)

// AccountsClient is an autogenerated mock type for the AccountsClient type
type AccountsClient struct {
	mock.Mock
}

// ListKeys provides a mock function with given fields: ctx, resourceGroupName, accountName, expand
func (_m *AccountsClient) ListKeys(ctx context.Context, resourceGroupName string, accountName string, expand storage.ListKeyExpand) (storage.AccountListKeysResult, error) {
	ret := _m.Called(ctx, resourceGroupName, accountName, expand)

	var r0 storage.AccountListKeysResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, storage.ListKeyExpand) (storage.AccountListKeysResult, error)); ok {
		return rf(ctx, resourceGroupName, accountName, expand)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, storage.ListKeyExpand) storage.AccountListKeysResult); ok {
		r0 = rf(ctx, resourceGroupName, accountName, expand)
	} else {
		r0 = ret.Get(0).(storage.AccountListKeysResult)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, storage.ListKeyExpand) error); ok {
		r1 = rf(ctx, resourceGroupName, accountName, expand)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewAccountsClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewAccountsClient creates a new instance of AccountsClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAccountsClient(t mockConstructorTestingTNewAccountsClient) *AccountsClient {
	mock := &AccountsClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
