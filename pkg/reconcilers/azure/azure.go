package azure

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/uuid"
	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex/clients/graphrbac"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex/clients/imds"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/azurex/clients/msi"
	"github.com/invisibl-cloud/identity-manager/pkg/reconcilers"
	"github.com/invisibl-cloud/identity-manager/pkg/util"
	"github.com/valyala/fasttemplate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// IdentityReconciler is the definition of
// Azure Identity Reconciler struct
type IdentityReconciler struct {
	client.Client
	scheme *runtime.Scheme
	res    *v1alpha1.WorkloadIdentity
	// internal
	debug bool
	rbac  *graphrbac.Client
	msi   *msi.Client
}

// NewIdentityReconciler initializes IdentityReconciler
func NewIdentityReconciler(base *reconcilers.ReconcilerBase, res *v1alpha1.WorkloadIdentity) *IdentityReconciler {
	return &IdentityReconciler{
		Client: base.Client(),
		scheme: base.Scheme(),
		res:    res,
	}
}

// Prepare prepares the reconciler
func (r *IdentityReconciler) Prepare(ctx context.Context) error {
	c, err := r.getAzurex(ctx)
	if err != nil {
		return err
	}
	r.rbac = graphrbac.New(c)
	r.msi = msi.New(c)
	r.debug = r.res.Annotations["debug"] == "true"
	if _, ok := r.res.Annotations["check-imds"]; ok {
		imds.Check(ctx, c.GetConfig().ClientID)
	}
	return nil
}

func (r *IdentityReconciler) getAzurex(ctx context.Context) (*azurex.Client, error) {
	creds := r.res.Spec.Credentials
	if creds == nil {
		return azurex.New(azurex.WithEnv())
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
	return azurex.New(azurex.WithEnv(), azurex.WithConfigMap(configMap))
}

// Reconcile reconciles the workload identity
func (r *IdentityReconciler) Reconcile(ctx context.Context) error {
	// reconcile
	id, err := r.doReconcile(ctx)
	if err != nil {
		return err
	}

	if id != nil {
		r.res.Status.ID = id.ID
		r.res.Status.Name = id.Name
	}

	if r.res.Status.ID == "" {
		return fmt.Errorf("waiting for identity to be created")
	}

	aid, err := r.doAzureIdentity(ctx, id)
	if err != nil {
		return err
	}

	err = r.doAzureIdentityBinding(ctx, aid)
	if err != nil {
		return err
	}

	if r.res.Spec.WriteToSecretRef != nil {
		err = r.doSecret(ctx, id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *IdentityReconciler) doSecret(ctx context.Context, id *msi.Identity) error {
	tmplData := map[string]any{
		"identity.id":          id.ID,
		"identity.resourceID":  id.ID,
		"identity.clientID":    id.ClientID,
		"identity.principalID": id.PrincipalID,
		"identity.name":        id.Name,
		"identity.tenantID":    id.TenantID,
		"tenantId":             id.TenantID,
		"subscriptionId":       id.SubscriptionID,
		"resourceGroup":        id.ResourceGroup,
		"location":             id.Location,
	}
	s := &corev1.Secret{}
	s.Name = util.DefaultString(r.res.Spec.WriteToSecretRef.Name, r.res.Name)
	s.Namespace = util.DefaultString(r.res.Spec.WriteToSecretRef.Namespace, r.res.Namespace)
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, s, func() error {
		err := controllerutil.SetControllerReference(r.res, s, r.scheme)
		if err != nil {
			return err
		}
		if len(s.Data) == 0 {
			s.Data = map[string][]byte{}
		}
		for k, v := range r.res.Spec.WriteToSecretRef.TemplateData {
			s.Data[k] = []byte(fasttemplate.ExecuteString(v, "<(", ")", tmplData))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error while creating or updating the object: %w", err)
	}

	return nil
}

func (r *IdentityReconciler) doAzureIdentity(ctx context.Context, id *msi.Identity) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("aadpodidentity.k8s.io/v1")
	u.SetKind("AzureIdentity")
	u.SetName(util.DefaultString(r.res.Spec.Name, r.res.Name))
	u.SetNamespace(r.res.Namespace)
	spec := r.res.Spec.Azure.Identity
	itype := 0
	if spec != nil {
		if spec.APIVersion != "" {
			u.SetAPIVersion(spec.APIVersion)
		}
		if spec.Kind != "" {
			u.SetKind(spec.Kind)
		}
		if spec.Metadata != nil {
			if spec.Metadata.Name != "" {
				u.SetName(spec.Metadata.Name)
			}
			if spec.Metadata.Namespace != "" {
				u.SetNamespace(spec.Metadata.Namespace)
			}
		}
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, u, func() error {
		err := controllerutil.SetControllerReference(r.res, u, r.scheme)
		if err != nil {
			return err
		}
		if spec != nil {
			if spec.Metadata != nil {
				labels := u.GetLabels()
				if len(labels) == 0 {
					labels := map[string]string{}
					u.SetLabels(labels)
				}
				for k, v := range spec.Metadata.Labels {
					labels[k] = v
				}
				annotations := u.GetLabels()
				if len(annotations) == 0 {
					annotations := map[string]string{}
					u.SetLabels(annotations)
				}
				if len(spec.Metadata.Annotations) == 0 { // default behaviour
					annotations["aadpodidentity.k8s.io/Behavior"] = "namespaced"
				}
				for k, v := range spec.Metadata.Annotations {
					annotations[k] = v
				}
			}
			if spec.Spec != nil && spec.Spec.Type != 0 {
				itype = spec.Spec.Type
			}
		}
		o := u.UnstructuredContent()
		o["spec"] = map[string]any{
			"type":       itype,
			"resourceID": r.res.Status.ID,
			"clientID":   id.ClientID,
		}
		return nil
	})
	if err != nil {
		return u, fmt.Errorf("error while creating or updating the object: %w", err)
	}
	return u, nil
}

func (r *IdentityReconciler) doAzureIdentityBinding(ctx context.Context, aid *unstructured.Unstructured) error {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("aadpodidentity.k8s.io/v1")
	u.SetKind("AzureIdentityBinding")
	u.SetName(aid.GetName())
	u.SetNamespace(aid.GetNamespace())
	spec := r.res.Spec.Azure.IdentityBinding
	if spec != nil {
		if spec.APIVersion != "" {
			u.SetAPIVersion(spec.APIVersion)
		}
		if spec.Kind != "" {
			u.SetKind(spec.Kind)
		}
		if spec.Metadata != nil {
			if spec.Metadata.Name != "" {
				u.SetName(spec.Metadata.Name)
			}
			if spec.Metadata.Namespace != "" {
				u.SetNamespace(spec.Metadata.Namespace)
			}
		}
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, u, func() error {
		err := controllerutil.SetControllerReference(r.res, u, r.scheme)
		if err != nil {
			return err
		}
		selector := aid.GetName()
		if spec != nil {
			if spec.Metadata != nil {
				labels := u.GetLabels()
				if len(labels) == 0 {
					labels := map[string]string{}
					u.SetLabels(labels)
				}
				for k, v := range spec.Metadata.Labels {
					labels[k] = v
				}
				annotations := u.GetLabels()
				if len(annotations) == 0 {
					annotations := map[string]string{}
					u.SetLabels(annotations)
				}
				for k, v := range spec.Metadata.Annotations {
					annotations[k] = v
				}
			}
			if spec.Spec != nil && spec.Spec.Selector != "" {
				selector = spec.Spec.Selector
			}
		}
		o := u.UnstructuredContent()
		o["spec"] = map[string]any{
			"azureIdentity": aid.GetName(),
			"selector":      selector,
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error while creating or updating the object: %w", err)
	}

	return nil
}

func (r *IdentityReconciler) doReconcile(ctx context.Context) (*msi.Identity, error) {
	log := log.FromContext(ctx)
	// TODO: what if the name gets changed?
	id, err := r.msi.CreateOrUpdate(ctx, util.DefaultString(r.res.Spec.Name, r.res.Name), map[string]*string{
		"managed-by": to.StringPtr("identity-manager.io"),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateOrUpdate: %w", err)
	}

	// Sync Custom Roles
	for _, v := range r.res.Spec.Azure.RoleDefinitions {
		hashData := strings.Join([]string{string(r.res.UID), r.res.Namespace, r.res.Name, v.ID, v.RoleName}, "/")
		id := uuid.NewMD5(uuid.Nil, []byte(hashData)).String()
		p := []authorization.Permission{}
		for _, permission := range v.Permissions {
			p = append(p, authorization.Permission{
				Actions:        to.StringSlicePtr(permission.Actions),
				NotActions:     to.StringSlicePtr(permission.NotActions),
				DataActions:    to.StringSlicePtr(permission.DataActions),
				NotDataActions: to.StringSlicePtr(permission.NotDataActions),
			})
		}
		err := r.rbac.CreateOrUpdateRoleDefinition(ctx, id, "", authorization.RoleDefinitionProperties{
			RoleName:    to.StringPtr(v.RoleName),
			RoleType:    to.StringPtr(v.RoleType),
			Description: to.StringPtr(v.Description),
			Permissions: &p,
		})
		if err != nil {
			return nil, fmt.Errorf("CreateOrUpdateRoleDefinition: %w", err)
		}
	}

	// Sync Role Assignments
	existing, err := r.rbac.ListRoleAssignments(ctx, id.PrincipalID)
	if err != nil {
		return nil, fmt.Errorf("ListRoleAssignments: %w", err)
	}
	existingIds := make([]string, len(existing))
	for i, ra := range existing {
		existingIds[i] = to.String(ra.Name)
	}

	newIds := []string{}
	idMap := map[string]string{}
	for k, a := range r.res.Spec.Azure.RoleAssignments {
		// if this algo gets changed, all role assignments will be recreated.
		// unless we store the role assignment ID in the resource.
		hashData := strings.Join([]string{id.PrincipalID, r.res.Namespace, r.res.Name, k, a.Role, a.Scope}, "/")
		raID := uuid.NewMD5(uuid.Nil, []byte(hashData)).String()
		newIds = append(newIds, raID)
		idMap[raID] = k
	}
	syncSteps := util.FindSyncSteps(existingIds, newIds)
	if r.debug {
		log.Info("debug", "existingIds", existingIds, "newIds", newIds, "syncSteps", syncSteps)
	}
	for _, raid := range syncSteps.Add {
		err = r.attachRoleDefinition(ctx, raid, id.PrincipalID, r.res.Spec.Azure.RoleAssignments[idMap[raid]])
		if err != nil {
			return nil, err
		}
	}
	for _, raid := range syncSteps.Delete {
		err = r.detachRoleDefinition(ctx, raid)
		if err != nil {
			return nil, err
		}
	}
	// TODO: update not needed?
	return id, nil
}

func (r *IdentityReconciler) detachRoleDefinition(ctx context.Context, id string) error {
	return r.rbac.DeleteRoleAssignment(ctx, id)
}

func (r *IdentityReconciler) attachRoleDefinition(ctx context.Context, id string, principalID string, a v1alpha1.RoleAssignment) error {
	roleDefinitionID, err := r.rbac.GetRoleDefintionIDFromName(ctx, a.Role, "")
	if err != nil {
		return fmt.Errorf("GetRoleDefintionIDFromName: %w", err)
	}
	fid, err := r.rbac.CreateRoleAssignment(ctx, id, principalID, roleDefinitionID, a.Scope)
	if err != nil {
		// TODO: check for RoleAssignmentExists
		return fmt.Errorf("CreateRoleAssignment: %s, %w", a.Role, err)
	}
	_ = fid
	return nil
}

// Finalize implements Finalizer interface
func (r *IdentityReconciler) Finalize(ctx context.Context) error {
	ok, err := r.msi.EnsureDelete(ctx, util.DefaultString(r.res.Spec.Name, r.res.Name))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("deleting identity %s", r.res.Spec.Name)
	}
	return nil
}
