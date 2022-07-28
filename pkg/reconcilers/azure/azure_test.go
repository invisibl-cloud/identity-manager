package azure

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	identitymanageriov1alpha1 "github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getTestClient() (client.Client, error) {
	kubeBuilderAssetsPath, ok := os.LookupEnv("KUBEBUILDER_ASSETS")
	if !ok || kubeBuilderAssetsPath == "" {
		return nil, errors.New("KUBEBUILDER_ASSETS env variable is not set or empty")
	}

	var testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}
	cfg, err := testEnv.Start()
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()

	err = appsv1.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	err = corev1.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	err = identitymanageriov1alpha1.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return k8sClient, nil
}

func newNamespace(k8sClient client.Client, saNamespace string) error {
	sa := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: saNamespace,
		},
	}
	err := k8sClient.Create(context.Background(), sa)
	if err != nil {
		return err
	}

	return nil
}

func newServiceAccount(k8sClient client.Client, saName, saNamespace, resName string) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: saNamespace,
			// Labels:    map[string]string{managedByLabelKey: managedByValueKey, roleLabelKey: resName},
		},
	}
	err := k8sClient.Create(context.Background(), sa)
	if err != nil {
		return err
	}

	return nil
}

func getServiceAccount(k8sClient client.Client, saName, saNamespace string) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: saName, Namespace: saNamespace}, sa)
	if err != nil {
		return nil, err
	}

	return sa, nil
}
