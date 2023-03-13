package reconcilers

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/go-logr/logr"
	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"github.com/invisibl-cloud/identity-manager/pkg/consts"
	"github.com/invisibl-cloud/identity-manager/pkg/types"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Reconcile calls the Reconciler
func Reconcile(ctx context.Context, base *ReconcilerBase, req ctrl.Request, res types.Resource, rec types.Reconciler) (ctrl.Result, error) {
	log := base.Log(ctx)
	// Fetch the instance
	err := base.Client().Get(ctx, req.NamespacedName, res)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("resource not found. ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "failed to get resource")
		return ctrl.Result{}, err
	}

	// create new reconciler for the req
	reconciler := reconciler{
		base:    base,
		log:     log,
		req:     req,
		res:     res,
		rec:     rec,
		resCopy: res.DeepCopyObject().(client.Object),
	}
	// run reconcile
	return reconciler.Reconcile(ctx)
}

// Reconciler
type reconciler struct {
	base *ReconcilerBase
	log  logr.Logger
	req  ctrl.Request
	res  types.Resource
	rec  types.Reconciler
	// custom
	resCopy client.Object
	// internal
	specCopy   any
	statusCopy any
}

// Reconcile implements Reconciler
func (r *reconciler) Reconcile(ctx context.Context) (ctrl.Result, error) {
	r.specCopy = r.res.GetSpecCopy()
	r.statusCopy = r.res.GetStatusCopy()
	// Check if the instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	if r.res.GetDeletionTimestamp() != nil {
		return r.doDelete(ctx)
	}
	// Add finalizer to control object deletion
	switch r.res.GetObjectKind().GroupVersionKind().Kind {
	case "Dependency":
		controllerutil.RemoveFinalizer(r.res, consts.FinalizerKey)
	default:
		if !controllerutil.ContainsFinalizer(r.res, consts.FinalizerKey) {
			controllerutil.AddFinalizer(r.res, consts.FinalizerKey)
		}
	}
	return r.doReconcile(ctx)
}

func (r *reconciler) doDelete(ctx context.Context) (ctrl.Result, error) {
	// if no finalizer, return
	if !controllerutil.ContainsFinalizer(r.res, consts.FinalizerKey) {
		return ctrl.Result{}, nil
	}
	// if reconciler off, return
	//if CanIgnore(r.res) {
	//	return r.removeFinalizer(ctx)
	//}
	// if marked as orphan, return
	isOrphan := r.res.GetAnnotations()[consts.DeletePolicyKey] == consts.OrphanValue
	if isOrphan {
		return r.removeFinalizer(ctx)
	}
	// PreFinalize
	switch rec1 := r.rec.(type) {
	case types.FinalizeReconciler:
		err := rec1.PreFinalize(ctx)
		if err != nil {
			return r.doStatus(ctx, err)
		}
	}
	// Run finalization logic for FinalizerAnnotationKey. If the
	// finalization logic fails, don't remove the finalizer so
	// that we can retry during the next reconciliation.
	if err := r.rec.Finalize(ctx); err != nil {
		//return ctrl.Result{}, err
		return r.doStatus(ctx, err)
	}
	return r.removeFinalizer(ctx)
}

func (r *reconciler) removeFinalizer(ctx context.Context) (ctrl.Result, error) {
	// Remove FinalizerAnnotationKey. Once all finalizers have been
	// removed, the object will be deleted.
	controllerutil.RemoveFinalizer(r.res, consts.FinalizerKey)
	if true {
		err := r.base.Client().Update(ctx, r.res)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *reconciler) doReconcile(ctx context.Context) (ctrl.Result, error) {
	return r.doStatus(ctx, r.rec.Reconcile(ctx))
}

func (r *reconciler) getRequeueAfter(requeueAfter time.Duration) time.Duration {
	if requeueAfter > 0 {
		return requeueAfter
	}
	min := 30
	max := 120
	secs, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return time.Duration(min) * time.Second
	}
	return time.Duration(secs.Int64()) * time.Second
}

func (r *reconciler) isResourceUpdated(spec any) bool {
	return !equality.Semantic.DeepEqual(r.resCopy.GetAnnotations(), r.res.GetAnnotations()) ||
		!equality.Semantic.DeepEqual(r.resCopy.GetLabels(), r.res.GetLabels()) ||
		!equality.Semantic.DeepEqual(r.resCopy.GetFinalizers(), r.res.GetFinalizers()) ||
		!equality.Semantic.DeepEqual(r.specCopy, spec)
}

// requeueAfter = -1 => no requeue
// requeueAfter = 0 => requeue after random seconds
// requeueAfter > 0 => requeue after requeueAfter seconds
func (r *reconciler) doReturn(ctx context.Context, requeueAfter int64) (ctrl.Result, error) {
	canRequeue := requeueAfter >= 0

	// always check status first.
	if !equality.Semantic.DeepEqual(r.statusCopy, r.res.GetStatus()) {
		err := r.base.Client().Status().Update(ctx, r.res)
		if err != nil {
			r.log.Info("error updating resource status", "err", err)
		}
	}
	// then spec
	if r.isResourceUpdated(r.res.GetSpec()) {
		err := r.base.Client().Update(ctx, r.res)
		if err != nil {
			r.log.Info("error updating resource", "err", err)
		}
	}
	if canRequeue {
		requeueDuration := r.getRequeueAfter(time.Duration(requeueAfter) * time.Second)
		return ctrl.Result{RequeueAfter: requeueDuration}, nil
	}
	return ctrl.Result{}, nil
}

func (r *reconciler) doStatus(ctx context.Context, err error) (ctrl.Result, error) {
	var c v1alpha1.Condition
	if err != nil { // on error
		if err == types.ErrIgnore {
			return r.doReturn(ctx, -1) // TOOD: requeue false?
		}
		switch cr := err.(type) {
		case v1alpha1.Condition:
			c = cr
		case *v1alpha1.Condition:
			c = *cr
		default: // default error
			c = v1alpha1.ReconcileError(err)
		}
	} else { // onsuccess
		c = v1alpha1.ReconcileSuccess()
	}
	cs := r.res.GetConditionedStatus()
	if cs != nil {
		cs.SetConditions(c)
	}
	return r.doReturn(ctx, 0)
}
