package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/awsx"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/awsx/iam"
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

var DefaultDuration = time.Duration(60) * time.Second

const eksServiceAccountAnnotationKey = "eks.amazonaws.com/role-arn"
const roleLabelKey = "identity-manager.io/name"
const managedByLabelKey = "identity-manager.io"
const managedByValueKey = "managed"

// Reconciler
type AWSRoleReconciler struct {
	client.Client
	scheme *runtime.Scheme
	res    *v1alpha1.WorkloadIdentity
	// internal
	iamClient *iam.Client
}

func NewAWSRoleReconciler(base *reconcilers.ReconcilerBase, res *v1alpha1.WorkloadIdentity) *AWSRoleReconciler {
	return &AWSRoleReconciler{
		Client: base.Client(),
		scheme: base.Scheme(),
		res:    res,
	}
}

// Initialize creates new RoleClient
func (r *AWSRoleReconciler) Prepare(ctx context.Context) error {
	conf, err := r.getConfig(ctx)
	if err != nil {
		//r.log.Error(err, "Failed to create AwsRole creds config")
		return err
	}
	r.iamClient = iam.New(&conf, r.res)
	return r.iamClient.Prepare(ctx)
}

func (r *AWSRoleReconciler) getConfig(ctx context.Context) (conf awsx.Config, err error) {
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

func (r *AWSRoleReconciler) Reconcile(ctx context.Context) error {
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

func (r *AWSRoleReconciler) doIAMRoleReconcile(ctx context.Context) error {
	status, err := r.iamClient.CreateOrUpdate(ctx)
	if err != nil {
		//r.log.Error(err, "Failed to reconcile AwsRole.")
		return err
	}
	if status != nil && status.Name != "" && status.ARN != "" &&
		(r.res.Status.Name != status.Name || r.res.Status.ID != status.ARN) {
		r.res.Status.ID = status.ARN
		r.res.Status.Name = status.Name
	}
	return nil
}

func (r *AWSRoleReconciler) Finalize(ctx context.Context) error {
	err := r.iamClient.Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *AWSRoleReconciler) doActions(ctx context.Context) error {
	// reconcile serviceaccount
	for _, sa := range r.res.Spec.AWS.ServiceAccounts {
		_, err := r.doServiceAccountReconcile(ctx, sa)
		if err != nil {
			return err
		}
	}

	// TODO: make sure all serviceaccounts has annotations?

	// do actions
	for _, p := range r.res.Spec.AWS.Pods {
		err := r.restartPods(ctx, p, r.res.Status.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *AWSRoleReconciler) doServiceAccountReconcile(ctx context.Context, saSpec *v1alpha1.ServiceAccount) (ctrl.Result, error) {
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
			sa := r.newServiceAccount(saName, saNamespace)
			//r.log.Info("Creating a new ServiceAccount", "Namespace", sa.Namespace, "Name", sa.Name)
			err = r.Create(ctx, sa)
			if err != nil {
				//r.log.Error(err, "Failed to create new ServiceAccount", "Namespace", sa.Namespace, "Name", sa.Name)
				return ctrl.Result{}, err
			}
			// ServiceAccount created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
		// TODO: If action:create but sa already found?
	case v1alpha1.ServiceAccountActionUpdate:
		if isNotFound {
			return ctrl.Result{}, fmt.Errorf("missing serviceaccount, cannot update")
		}
	}

	/*
		// delete sa only if its created & managed by this controller.
		if existingSA.Labels[managedByLabelKey] == managedByValueKey {
			// delete service account.
			err = r.Delete(r.ctx, existingSA)
			if err != nil {
				r.log.Error(err, "Failed to delete ServiceAccount", "Namespace", saNamespace, "Name", saName)
				return ctrl.Result{}, err
			}
			// ServiceAccount deleted successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	*/

	// wait till Role ARN is created.
	arn := r.res.Status.ID
	if arn == "" {
		return ctrl.Result{}, nil
	}

	// update service account
	if existingSA.Annotations[eksServiceAccountAnnotationKey] != arn {
		if len(existingSA.Annotations) == 0 {
			existingSA.Annotations = map[string]string{}
		}
		existingSA.Annotations[eksServiceAccountAnnotationKey] = arn
		err = r.Update(ctx, existingSA)
		if err != nil {
			//r.log.Error(err, "Failed to update ServiceAccount")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AWSRoleReconciler) restartPods(ctx context.Context, p *v1alpha1.AwsRoleSpecPod, arn string) error {
	pods := &corev1.PodList{}
	err := r.List(ctx, pods,
		client.InNamespace(util.DefaultString(p.Namespace, r.res.Namespace)),
		client.MatchingLabels(p.MatchLabels),
	)
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		count := 0
		for _, c := range pod.Spec.Containers {
			for _, env := range c.Env {
				if env.Value == arn && env.Name == "AWS_ROLE_ARN" {
					count++
				}
			}
		}
		found := count > 0 // TODO: count == len(pod.Spec.Containers) ??
		if !found {
			//r.log.Info("deleting pod", "pod", pod.Name, "namespace", pod.Namespace)
			err = r.Delete(ctx, &pod)
			if err != nil {
				//r.log.Info("error deleting pod", "pod", pod.Name, "error", err)
			} else {
				// requeue??
				// return nil
			}
		}
	}
	return nil
}

func (r *AWSRoleReconciler) newServiceAccount(saName string, saNamespace string) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: saNamespace,
			Labels:    map[string]string{managedByLabelKey: managedByValueKey, roleLabelKey: r.res.Name},
			//Annotations: map[string]string{eksServiceAccountAnnotationKey: m.Status.ARN},
		},
	}
	// Set AwsRole instance as the owner and controller (for gc)
	ctrl.SetControllerReference(r.res, sa, r.scheme)
	return sa
}
