package types

import (
	"context"
	"errors"

	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrIgnore = errors.New("IgnoreError")

type Reconciler interface {
	Reconcile(ctx context.Context) error
	Finalize(ctx context.Context) error
}

type FinalizeReconciler interface {
	PreFinalize(ctx context.Context) error
}

/*
type RefreshDisabled interface {
	RefreshDisabled() bool
}
*/

type ResourceBase interface {
	client.Object
	GetSpec() interface{}
	GetSpecCopy() interface{}
	GetStatus() interface{}
	GetStatusCopy() interface{}
}

type Resource interface {
	ResourceBase
	GetConditionedStatus() *v1alpha1.ConditionedStatus
}

// test whether the CRs implements Resource interface{}
var (
	_ = Resource(&v1alpha1.WorkloadIdentity{})
)
