package reconcilers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ReconcilerBase struct {
	name       string
	scheme     *runtime.Scheme
	client     client.Client
	restConfig *rest.Config
	recorder   record.EventRecorder
	apireader  client.Reader
}

func NewForManager(name string, mgr manager.Manager) *ReconcilerBase {
	return &ReconcilerBase{
		name:       name,
		scheme:     mgr.GetScheme(),
		client:     mgr.GetClient(),
		restConfig: mgr.GetConfig(),
		recorder:   mgr.GetEventRecorderFor(name),
		apireader:  mgr.GetAPIReader(),
	}
}

func (w *ReconcilerBase) Name() string {
	return w.name
}

func (w *ReconcilerBase) Log(ctx context.Context) logr.Logger {
	return log.FromContext(ctx)
}

func (w *ReconcilerBase) Scheme() *runtime.Scheme {
	return w.scheme
}

func (w *ReconcilerBase) Client() client.Client {
	return w.client
}

// GetClient returns the underlying client
func (r *ReconcilerBase) GetClient() client.Client {
	return r.client
}

// GetScheme returns the scheme
func (r *ReconcilerBase) GetScheme() *runtime.Scheme {
	return r.scheme
}
