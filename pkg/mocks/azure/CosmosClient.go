// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	documentdb "github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2021-10-15/documentdb"

	mock "github.com/stretchr/testify/mock"
)

// CosmosClient is an autogenerated mock type for the CosmosClient type
type CosmosClient struct {
	mock.Mock
}

// ListKeys provides a mock function with given fields: ctx, resourceGroupName, accountName
func (_m *CosmosClient) ListKeys(ctx context.Context, resourceGroupName string, accountName string) (documentdb.DatabaseAccountListKeysResult, error) {
	ret := _m.Called(ctx, resourceGroupName, accountName)

	var r0 documentdb.DatabaseAccountListKeysResult
	if rf, ok := ret.Get(0).(func(context.Context, string, string) documentdb.DatabaseAccountListKeysResult); ok {
		r0 = rf(ctx, resourceGroupName, accountName)
	} else {
		r0 = ret.Get(0).(documentdb.DatabaseAccountListKeysResult)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, resourceGroupName, accountName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewCosmosClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewCosmosClient creates a new instance of CosmosClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewCosmosClient(t mockConstructorTestingTNewCosmosClient) *CosmosClient {
	mock := &CosmosClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
