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

	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"github.com/invisibl-cloud/identity-manager/pkg/consts"
	"github.com/invisibl-cloud/identity-manager/pkg/reconcilers"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// AWSAuthReconciler reconciles a AWSAuth object
type AWSAuthReconciler struct {
	Base *reconcilers.ReconcilerBase
}

//+kubebuilder:rbac:groups=identity-manager.io,resources=awsauths,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=identity-manager.io,resources=awsauths/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=identity-manager.io,resources=awsauths/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile AWSAuth
func (r *AWSAuthReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	res := &v1alpha1.AWSAuth{}
	rec := &aaReconciler{base: r.Base, res: res}
	return reconcilers.Reconcile(ctx, r.Base, req, res, rec)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSAuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AWSAuth{}).
		Complete(r)
}

// Reconciler
type aaReconciler struct {
	base *reconcilers.ReconcilerBase
	res  *v1alpha1.AWSAuth
}

// YAMLUnmarshal unmarshal yaml string
func YAMLUnmarshal(str string, out any) error {
	if str == "" {
		return nil
	}
	return yaml.Unmarshal([]byte(str), out)
}

// Reconcile reconciles the awsauth resource
func (r *aaReconciler) Reconcile(ctx context.Context) error {
	cm := &corev1.ConfigMap{}
	cm.Name = r.res.Name
	cm.Namespace = r.res.Namespace
	_, err := ctrl.CreateOrUpdate(ctx, r.base.Client(), cm, func() error {
		// load existing config
		existingMapRoles := []*MapRoleItem{}
		err := YAMLUnmarshal(cm.Data["mapRoles"], &existingMapRoles)
		if err != nil {
			return err
		}
		retainedMapRoles := []*MapRoleItem{}
		for _, mapRole := range existingMapRoles {
			// retain this
			if mapRole.Source == "" {
				retainedMapRoles = append(retainedMapRoles, mapRole)
			}
		}
		currentMapRoles, currentMapUsers, err := r.getCurrentItems(ctx)
		if err != nil {
			return err
		}
		// nothing to update. ignore and return immediately.
		if len(currentMapRoles)+len(currentMapUsers) == 0 {
			return nil
		}
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		data, err := yaml.Marshal(append(retainedMapRoles, currentMapRoles...))
		if err == nil {
			cm.Data["mapRoles"] = string(data)
		} else {
			r.base.Log(ctx).Info("ignoring.. error marshaling mapRoles", "err", err)
		}
		data, err = yaml.Marshal(currentMapUsers)
		if err == nil {
			cm.Data["mapUsers"] = string(data)
		} else {
			r.base.Log(ctx).Info("ignoring.. error marshaling mapUsers", "err", err)
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (r *aaReconciler) getCurrentItems(ctx context.Context) ([]*MapRoleItem, []*MapUserItem, error) {
	currentMapRoles := []*MapRoleItem{}
	currentMapUsers := []*MapUserItem{}
	// get list of aws-auth configs from configmaps
	l := &corev1.ConfigMapList{}
	err := r.base.Client().List(ctx, l, client.InNamespace(r.res.Namespace), client.MatchingLabels{
		consts.AWSAuthNameKey: r.res.Name,
	})
	if err != nil {
		return nil, nil, err
	}
	if len(l.Items) == 0 {
		return nil, nil, nil
	}
	for _, item := range l.Items {
		mapRolesItems := []*MapRoleItem{}
		mapUsersItems := []*MapUserItem{}
		err := YAMLUnmarshal(item.Data["mapRoles"], &mapRolesItems)
		if err != nil {
			r.base.Log(ctx).Info("ignoring.. error unmarshaling mapRoles in configmap", "cm.ns", item.GetNamespace(), "cm.name", item.GetName())
		}
		err = YAMLUnmarshal(item.Data["mapUsers"], &mapUsersItems)
		if err != nil {
			r.base.Log(ctx).Info("ignoring.. error unmarshaling mapUsers in configmap", "cm.ns", item.GetNamespace(), "cm.name", item.GetName())
		}
		for _, ritem := range mapRolesItems {
			if ritem.Username != "" && ritem.RoleArn != "" && len(ritem.Groups) > 0 {
				ritem.Source = item.Name
				currentMapRoles = append(currentMapRoles, ritem)
			}
		}
		for _, uitem := range mapUsersItems {
			if uitem.Username != "" && uitem.UserArn != "" && len(uitem.Groups) > 0 {
				uitem.Source = item.Name
				currentMapUsers = append(currentMapUsers, uitem)
			}
		}
	}
	return currentMapRoles, currentMapUsers, nil
}

// Finalize implements Finalizer interface
func (r *aaReconciler) Finalize(ctx context.Context) error {
	return nil
}

// MapRoleItem defines the mapRole item of AWSAuth
type MapRoleItem struct {
	Source   string   `json:"source,omitempty" yaml:"source,omitempty"`
	RoleArn  string   `json:"rolearn" yaml:"rolearn"`
	Username string   `json:"username" yaml:"username"`
	Groups   []string `json:"groups" yaml:"groups"`
}

// MapUserItem defines the mapUser item of AWSAuth
type MapUserItem struct {
	Source   string   `json:"source,omitempty" yaml:"source,omitempty"`
	UserArn  string   `json:"userarn" yaml:"userarn"`
	Username string   `json:"username" yaml:"username"`
	Groups   []string `json:"groups" yaml:"groups"`
}
