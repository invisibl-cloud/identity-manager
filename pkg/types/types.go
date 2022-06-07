package types

import (
	"context"
	"errors"

	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ErrIgnore defines the ignorable error
var ErrIgnore = errors.New("IgnoreError")

// Reconciler is the interface that facilitates Reconcile and Finalize
type Reconciler interface {
	Reconcile(ctx context.Context) error
	Finalize(ctx context.Context) error
}

// FinalizeReconciler is the interface that facilitates PreFinalize
type FinalizeReconciler interface {
	PreFinalize(ctx context.Context) error
}

// ResourceBase is the interface that facilitates getters
type ResourceBase interface {
	client.Object
	GetSpec() interface{}
	GetSpecCopy() interface{}
	GetStatus() interface{}
	GetStatusCopy() interface{}
}

// Resource is the interface that facilitates getters
type Resource interface {
	ResourceBase
	GetConditionedStatus() *v1alpha1.ConditionedStatus
}

// test whether the CRs implements Resource interface{}
var (
	_ = Resource(&v1alpha1.WorkloadIdentity{})
)
