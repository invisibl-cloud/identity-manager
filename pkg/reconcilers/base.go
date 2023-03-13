package reconcilers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/invisibl-cloud/identity-manager/pkg/options"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ReconcilerBase defines the base reconciler struct
type ReconcilerBase struct {
	name       string
	scheme     *runtime.Scheme
	client     client.Client
	restConfig *rest.Config
	recorder   record.EventRecorder
	apireader  client.Reader
	options    *options.Options
}

// NewForManager expects name and manager.Manager and returns *ReconcilerBase
func NewForManager(name string, mgr manager.Manager, options *options.Options) *ReconcilerBase {
	return &ReconcilerBase{
		name:       name,
		scheme:     mgr.GetScheme(),
		client:     mgr.GetClient(),
		restConfig: mgr.GetConfig(),
		recorder:   mgr.GetEventRecorderFor(name),
		apireader:  mgr.GetAPIReader(),
		options:    options,
	}
}

// Name returns the Name of the ReconcilerBase
func (r *ReconcilerBase) Name() string {
	return r.name
}

// Log returns the Logger of the ReconcilerBase
func (r *ReconcilerBase) Log(ctx context.Context) logr.Logger {
	return log.FromContext(ctx)
}

// Scheme returns the Scheme of the ReconcilerBase
func (r *ReconcilerBase) Scheme() *runtime.Scheme {
	return r.scheme
}

// Client returns the Client of the ReconcilerBase
func (r *ReconcilerBase) Client() client.Client {
	return r.client
}

// Options returns Options from the ReconcilerBase
func (r *ReconcilerBase) Options() *options.Options {
	return r.options
}
