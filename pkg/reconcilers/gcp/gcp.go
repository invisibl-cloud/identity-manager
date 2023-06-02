package gcp

import (
	"context"
	"fmt"

	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/gcpx"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/gcpx/iam"
	"github.com/invisibl-cloud/identity-manager/pkg/reconcilers"
	"github.com/invisibl-cloud/identity-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const serviceAccountAnnotationKey = "iam.gke.io/gcp-service-account"
const roleLabelKey = "identity-manager.io/name"
const managedByLabelKey = "identity-manager.io"
const managedByValueKey = "managed"

// IdentityReconciler reconciles GCP Identity
type IdentityReconciler struct {
	client.Client
	scheme *runtime.Scheme
	res    *v1alpha1.WorkloadIdentity
	// internal
	debug bool
	gcpx  *gcpx.Client
	iamx  *iam.Client
}

// NewReconciler initializes IdentityReconciler
func NewReconciler(base *reconcilers.ReconcilerBase, res *v1alpha1.WorkloadIdentity) *IdentityReconciler {
	return &IdentityReconciler{
		Client: base.Client(),
		scheme: base.Scheme(),
		res:    res,
	}
}

// Prepare prepares for reconcilation
func (r *IdentityReconciler) Prepare(ctx context.Context) error {
	c, err := r.getGCP(ctx)
	if err != nil {
		return err
	}
	r.gcpx = c
	r.iamx = iam.New(r.gcpx)
	r.debug = r.res.Annotations["debug"] == "true"
	return nil
}

func (r *IdentityReconciler) getGCP(ctx context.Context) (*gcpx.Client, error) {
	creds := r.res.Spec.Credentials
	if creds == nil {
		return gcpx.New(gcpx.WithEnv())
	}
	configMap := map[string]any{}
	switch creds.Source {
	case v1alpha1.CredentialsSourceSecret:
		if creds.SecretRef == nil {
			return nil, fmt.Errorf("missing secretRef for credentials")
		}
		secret := &corev1.Secret{}
		secret.Name = util.DefaultString(creds.SecretRef.Name, r.res.Name)
		secret.Namespace = util.DefaultString(creds.SecretRef.Namespace, r.res.Namespace)
		err := r.Get(ctx, client.ObjectKeyFromObject(secret), secret)
		if err != nil {
			return nil, err
		}
		for k, v := range secret.Data {
			configMap[k] = string(v)
		}
	}
	for k, v := range creds.Properties {
		configMap[k] = v
	}
	return gcpx.New(gcpx.WithEnv(), gcpx.WithConfigMap(configMap))
}

// Reconcile reconciles the workload identity
func (r *IdentityReconciler) Reconcile(ctx context.Context) error {
	// reconcile
	err := r.doReconcile(ctx)
	if err != nil {
		return err
	}
	if r.res.Status.ID == "" {
		return fmt.Errorf("waiting for identity to be created")
	}
	return nil
}

func (r *IdentityReconciler) doReconcile(ctx context.Context) error {
	if r.res.Spec.GCP == nil {
		return fmt.Errorf("missing gcp spec")
	}
	name := util.DefaultString(r.res.Spec.Name, r.res.Name)
	id, err := r.iamx.EnsureServiceAccountWithRoles(ctx,
		name,
		r.res.Namespace,
		r.res.Spec.GCP.ServiceAccounts,
		r.res.Spec.DisplayName,
		r.res.Spec.Description,
		r.res.Spec.GCP.Roles,
		"",
	)
	if err != nil {
		return err
	}
	//log := log.FromContext(ctx)
	if id != "" {
		r.res.Status.ID = id
	}
	return r.doActions(ctx)
}

// Finalize implements Finalizer interface
func (r *IdentityReconciler) Finalize(ctx context.Context) error {
	if r.res.Spec.GCP == nil {
		return nil
	}
	err := r.iamx.DeleteServiceAccount(ctx, r.res.Status.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *IdentityReconciler) doActions(ctx context.Context) error {
	// reconcile serviceaccount
	for _, sa := range r.res.Spec.GCP.ServiceAccounts {
		_, err := r.doServiceAccountReconcile(ctx, sa)
		if err != nil {
			return err
		}
	}

	// do actions
	//for _, p := range r.res.Spec.GCP.Pods {
	//	err := r.restartPods(ctx, p, r.res.Status.ID)
	//	if err != nil {
	//		return err
	//	}
	//}
	return nil
}

func (r *IdentityReconciler) doServiceAccountReconcile(ctx context.Context, saSpec *v1alpha1.ServiceAccount) (ctrl.Result, error) {
	// do nothing if no action
	if saSpec.Action == v1alpha1.ServiceAccountActionDefault {
		return ctrl.Result{}, nil
	}

	// reconcile service account.
	existingSA := &corev1.ServiceAccount{}
	saName := util.DefaultString(saSpec.Name, r.res.Name)
	saNamespace := util.DefaultString(saSpec.Namespace, r.res.Namespace)
	err := r.Get(ctx, types.NamespacedName{Name: saName, Namespace: saNamespace}, existingSA)
	isNotFound := false
	if err != nil {
		isNotFound = errors.IsNotFound(err)
		if !isNotFound {
			return ctrl.Result{}, err
		}
	}

	switch saSpec.Action {
	case v1alpha1.ServiceAccountActionCreate:
		if isNotFound {
			// if need to create and its not found. create one.
			// Define a new ServiceAccount
			sa, err := r.newServiceAccount(saName, saNamespace)
			if err != nil {
				return ctrl.Result{}, err
			}
			err = r.Create(ctx, sa)
			if err != nil {
				return ctrl.Result{}, err
			}
			// ServiceAccount created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
		// else
		return ctrl.Result{}, fmt.Errorf("unable to create serviceaccount: serviceaccount already exists")

	case v1alpha1.ServiceAccountActionUpdate:
		if isNotFound {
			return ctrl.Result{}, fmt.Errorf("missing serviceaccount, cannot update")
		}
	}

	// wait till Role ARN is created.
	arn := r.res.Status.ID
	if arn == "" {
		return ctrl.Result{}, nil
	}

	// update service account
	if existingSA.Annotations == nil {
		existingSA.Annotations = make(map[string]string)
	}

	if existingSA.Annotations[serviceAccountAnnotationKey] != arn {
		if len(existingSA.Annotations) == 0 {
			existingSA.Annotations = map[string]string{}
		}
		existingSA.Annotations[serviceAccountAnnotationKey] = arn
		err = r.Update(ctx, existingSA)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *IdentityReconciler) newServiceAccount(saName string, saNamespace string) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        saName,
			Namespace:   saNamespace,
			Labels:      map[string]string{managedByLabelKey: managedByValueKey, roleLabelKey: r.res.Name},
			Annotations: map[string]string{serviceAccountAnnotationKey: r.res.Status.ID},
		},
	}
	// Set AwsRole instance as the owner and controller (for gc)
	err := ctrl.SetControllerReference(r.res, sa, r.scheme)
	if err != nil {
		return nil, err
	}

	return sa, nil
}
