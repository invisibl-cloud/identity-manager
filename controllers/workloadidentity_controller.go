/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"github.com/invisibl-cloud/identity-manager/pkg/reconcilers"
	"github.com/invisibl-cloud/identity-manager/pkg/reconcilers/aws"
	"github.com/invisibl-cloud/identity-manager/pkg/reconcilers/azure"
	"github.com/invisibl-cloud/identity-manager/pkg/types"
)

// WorkloadIdentityReconciler reconciles a WorkloadIdentity object
type WorkloadIdentityReconciler struct {
	Base *reconcilers.ReconcilerBase
}

//+kubebuilder:rbac:groups=identity-manager.io,resources=workloadidentities,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=identity-manager.io,resources=workloadidentities/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=identity-manager.io,resources=workloadidentities/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch;delete
//+kubebuilder:rbac:groups=aadpodidentity.k8s.io,resources=azureidentities,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aadpodidentity.k8s.io,resources=azureidentitybindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WorkloadIdentity object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *WorkloadIdentityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	res := &v1alpha1.WorkloadIdentity{}
	rec := &wiReconciler{base: r.Base, res: res}
	return reconcilers.Reconcile(ctx, r.Base, req, res, rec)
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkloadIdentityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.WorkloadIdentity{}).
		// Not needed, crashes in EKS env.
		//Owns(util.UnstructuredObject("aadpodidentity.k8s.io/v1", "AzureIdentity")).
		//Owns(util.UnstructuredObject("aadpodidentity.k8s.io/v1", "AzureIdentityBinding")).
		Complete(r)
}

// Reconciler
type wiReconciler struct {
	base *reconcilers.ReconcilerBase
	res  *v1alpha1.WorkloadIdentity
}

type wiReconcilerInterface interface {
	types.Reconciler
	Prepare(context.Context) error
}

// Reconcile reconciles the workload identity
func (r *wiReconciler) Reconcile(ctx context.Context) error {
	var rec wiReconcilerInterface
	switch r.res.Spec.Provider {
	case v1alpha1.ProviderAWS:
		rec = aws.NewRoleReconciler(r.base, r.res)
	case v1alpha1.ProviderAzure:
		rec = azure.NewIdentityReconciler(r.base, r.res)
	default:
		return fmt.Errorf("unknown provider %s", r.res.Spec.Provider)
	}
	err := rec.Prepare(ctx)
	if err != nil {
		return err
	}
	return rec.Reconcile(ctx)
}

// Finalize implements Finalizer interface
func (r *wiReconciler) Finalize(ctx context.Context) error {
	var rec wiReconcilerInterface
	switch r.res.Spec.Provider {
	case v1alpha1.ProviderAWS:
		rec = aws.NewRoleReconciler(r.base, r.res)
	case v1alpha1.ProviderAzure:
		rec = azure.NewIdentityReconciler(r.base, r.res)
	default:
		return nil
	}
	err := rec.Prepare(ctx)
	if err != nil {
		return err
	}
	return rec.Finalize(ctx)
}
