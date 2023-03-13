package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"github.com/invisibl-cloud/identity-manager/pkg/options"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/awsx"
	iamc "github.com/invisibl-cloud/identity-manager/pkg/providers/awsx/iam"
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

// DefaultDuration is the duration used when no
// specific duration is needed
var DefaultDuration = time.Duration(60) * time.Second

const eksServiceAccountAnnotationKey = "eks.amazonaws.com/role-arn"
const roleLabelKey = "identity-manager.io/name"
const managedByLabelKey = "identity-manager.io"
const managedByValueKey = "managed"

// RoleReconciler is the struct that holds the
// required fields that are need to reconcile
type RoleReconciler struct {
	client.Client
	scheme  *runtime.Scheme
	options *options.Options
	res     *v1alpha1.WorkloadIdentity
	// internal
	iamClient *iamc.Client
}

// NewRoleReconciler expectes reconciler base and workload identity resource
// and returns a new RoleReconciler object
func NewRoleReconciler(base *reconcilers.ReconcilerBase, res *v1alpha1.WorkloadIdentity) *RoleReconciler {
	return &RoleReconciler{
		Client:  base.Client(),
		scheme:  base.Scheme(),
		res:     res,
		options: base.Options(),
	}
}

// Prepare initialize creates new IAMClient
func (r *RoleReconciler) Prepare(ctx context.Context) error {
	conf, err := r.getConfig(ctx)
	if err != nil {
		return err
	}
	sess, err := awsx.NewSession(conf)
	if err != nil {
		return err
	}

	iamC := iam.New(sess)
	stsC := sts.New(sess)

	iamClient, err := iamc.New(iamC, stsC, r.res, r.options)
	if err != nil {
		return err
	}
	r.iamClient = iamClient
	return nil
}

func (r *RoleReconciler) getConfig(ctx context.Context) (conf awsx.Config, err error) {
	creds := r.res.Spec.Credentials
	if creds == nil {
		return
	}
	switch creds.Source {
	case v1alpha1.CredentialsSourceSecret:
		if creds.SecretRef == nil {
			err = fmt.Errorf("missing secretRef for credentials")
			return
		}
		secret := &corev1.Secret{}
		secret.Name = util.DefaultString(creds.SecretRef.Name, r.res.Name)
		secret.Namespace = util.DefaultString(creds.SecretRef.Namespace, r.res.Namespace)
		err = r.Get(ctx, client.ObjectKeyFromObject(secret), secret)
		if err != nil {
			return
		}
		for k, v := range creds.Properties {
			secret.Data[k] = []byte(v)
		}
		conf = awsx.NewConfig(secret.Data)
		return
	}
	return
}

// Reconcile performs Reconcilation
func (r *RoleReconciler) Reconcile(ctx context.Context) error {
	// reconcile IAM Role
	err := r.doIAMRoleReconcile(ctx)
	if err != nil {
		return err
	}

	// wait till role is created.
	if r.res.Status.ID == "" {
		return fmt.Errorf("waiting for role to be created! ARN is empty")
	}

	return r.doActions(ctx)
}

func (r *RoleReconciler) doIAMRoleReconcile(ctx context.Context) error {
	status, err := r.iamClient.CreateOrUpdate(ctx)
	if err != nil {
		return err
	}
	if status != nil && status.Name != "" && status.ARN != "" &&
		(r.res.Status.Name != status.Name || r.res.Status.ID != status.ARN) {
		r.res.Status.ID = status.ARN
		r.res.Status.Name = status.Name
	}
	return nil
}

// Finalize is the implementation of Finalizer
func (r *RoleReconciler) Finalize(ctx context.Context) error {
	err := r.iamClient.Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoleReconciler) doActions(ctx context.Context) error {
	// reconcile serviceaccount
	for _, sa := range r.res.Spec.AWS.ServiceAccounts {
		_, err := r.doServiceAccountReconcile(ctx, sa)
		if err != nil {
			return err
		}
	}

	// do actions
	for _, p := range r.res.Spec.AWS.Pods {
		err := r.restartPods(ctx, p, r.res.Status.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RoleReconciler) doServiceAccountReconcile(ctx context.Context, saSpec *v1alpha1.ServiceAccount) (ctrl.Result, error) {
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

	if existingSA.Annotations[eksServiceAccountAnnotationKey] != arn {
		if len(existingSA.Annotations) == 0 {
			existingSA.Annotations = map[string]string{}
		}
		existingSA.Annotations[eksServiceAccountAnnotationKey] = arn
		err = r.Update(ctx, existingSA)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *RoleReconciler) restartPods(ctx context.Context, p *v1alpha1.AwsRoleSpecPod, arn string) error {
	pods := &corev1.PodList{}
	err := r.List(ctx, pods,
		client.InNamespace(util.DefaultString(p.Namespace, r.res.Namespace)),
		client.MatchingLabels(p.MatchLabels),
	)
	if err != nil {
		return err
	}
	for i, pod := range pods.Items {
		count := 0
		for _, c := range pod.Spec.Containers {
			for _, env := range c.Env {
				if env.Value == arn && env.Name == "AWS_ROLE_ARN" {
					count++
				}
			}
		}

		// if there are no containers with with aws role arn env, delete the pod
		// TODO: support rolling restart instead of delete
		if count == 0 {
			// ignore error
			_ = r.Delete(ctx, &pods.Items[i])
			// requeue for else?
		}
	}
	return nil
}

func (r *RoleReconciler) newServiceAccount(saName string, saNamespace string) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        saName,
			Namespace:   saNamespace,
			Labels:      map[string]string{managedByLabelKey: managedByValueKey, roleLabelKey: r.res.Name},
			Annotations: map[string]string{eksServiceAccountAnnotationKey: r.res.Status.ID},
		},
	}
	// Set AwsRole instance as the owner and controller (for gc)
	err := ctrl.SetControllerReference(r.res, sa, r.scheme)
	if err != nil {
		return nil, err
	}

	return sa, nil
}
