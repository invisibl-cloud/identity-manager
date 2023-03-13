package aws

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/awsx"
	iamx "github.com/invisibl-cloud/identity-manager/pkg/providers/awsx/iam"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/awsx/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	identitymanageriov1alpha1 "github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
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

func TestAWSReconcile(t *testing.T) {
	testCases := []struct {
		desc                  string
		workloadIdentity      *v1alpha1.WorkloadIdentity
		setupMockExpectations func(*mocks.IAM, *mocks.STS)
	}{
		{
			desc: "Create new role and policies and update service account",
			workloadIdentity: &v1alpha1.WorkloadIdentity{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ccs-v1",
					UID:  "0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba",
				},
				Spec: v1alpha1.WorkloadIdentitySpec{
					Name:     "ccs-v1",
					Provider: v1alpha1.ProviderAWS,
					AWS: &v1alpha1.WorkloadIdentityAWS{
						AssumeRolePolicy: `{
							"Version": "2012-10-17",
							"Statement": [
							  {
								"Effect": "Allow",
								"Principal": {
								  "Federated": "arn:aws:iam::12345678:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/B1267BBA7830580562E3AD71DFC27CE7"
								},
								"Action": "sts:AssumeRoleWithWebIdentity",
								"Condition": {
								  "StringEquals": {
									"oidc.eks.us-east-1.amazonaws.com/id/B1267BBA7830580562E3AD71DFC27CE7:sub": "system:serviceaccount:dev:ccs-v1"
								  }
								}
							  }
							]
						  }`,
						InlinePolicies: map[string]string{
							"kms-read-0": `{
								"Version": "2012-10-17",
								"Statement": [
								  {
									"Effect": "Allow",
									"Action": [
									  "kms:DescribeKey",
									  "kms:GenerateDataKey",
									  "kms:Decrypt"
									],
									"Resource": [
									  "arn:aws:kms:us-east-1:12345678:key/27742849-43f5-c2fd-8930-0708b4ae6534"
									]
								  }
								]
							  }`,
						},
						ServiceAccounts: []*v1alpha1.ServiceAccount{
							{
								Name:      "dev-sa",
								Namespace: "dev",
								Action:    v1alpha1.ServiceAccountActionCreate,
							},
						},
					},
				},
			},
			setupMockExpectations: func(awsIAMClient *mocks.IAM, awsSTSClient *mocks.STS) {
				awsSTSClient.On("GetCallerIdentity", &sts.GetCallerIdentityInput{}).Return(
					&sts.GetCallerIdentityOutput{
						Account: aws.String("12345678"),
					}, nil,
				)

				awsIAMClient.On("GetRole", &iam.GetRoleInput{
					RoleName: aws.String("ccs-v1"),
				}).Return(nil, errors.New("role not found in AWS"))

				awsIAMClient.On("CreateRole", mock.MatchedBy(func(in *iam.CreateRoleInput) bool {
					return *in.RoleName == "ccs-v1"
				})).Return(
					&iam.CreateRoleOutput{
						Role: &iam.Role{
							Arn:      aws.String("arn:aws:iam::12345678:role/ccs-v1"),
							RoleName: aws.String("ccs-v1"),
						},
					}, nil,
				)

				awsIAMClient.On("PutRolePolicy", mock.MatchedBy(func(in *iam.PutRolePolicyInput) bool {
					return *in.RoleName == "ccs-v1"
				})).Return(nil, nil)
			},
		},
		{
			desc: "Update existing role and policies and update service account",
			workloadIdentity: &v1alpha1.WorkloadIdentity{
				Spec: v1alpha1.WorkloadIdentitySpec{
					Name:     "ccs-v1",
					Provider: v1alpha1.ProviderAWS,
					AWS: &v1alpha1.WorkloadIdentityAWS{
						AssumeRolePolicy: `{
							"Version": "2012-10-17",
							"Statement": [
							  {
								"Effect": "Allow",
								"Principal": {
								  "Federated": "arn:aws:iam::554248189203:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/B83A70562E3AD70581267BB1DFC27CE7"
								},
								"Action": "sts:AssumeRoleWithWebIdentity",
								"Condition": {
								  "StringEquals": {
									"oidc.eks.us-east-1.amazonaws.com/id/B83A70562E3AD70581267BB1DFC27CE7:sub": "system:serviceaccount:dev:ccs-v1"
								  }
								}
							  },
							  {
								"Effect": "Allow",
								"Principal": {
								  "AWS": "arn:aws:iam::554248189203:role/dev-eks-wl1-7szd5-keda-operator"
								},
								"Action": "sts:AssumeRole"
							  }
							]
						  }`,
						InlinePolicies: map[string]string{
							"kms-read-0": `{
								"Version": "2012-10-17",
								"Statement": [
								  {
									"Effect": "Allow",
									"Action": [
									  "kms:DescribeKey",
									  "kms:GenerateDataKey",
									  "kms:Decrypt"
									],
									"Resource": [
									  "arn:aws:kms:us-east-1:12345678:key/27742849-43f5-c2fd-8930-0708b4ae6534"
									]
								  }
								]
							  }`,
							"sqs-read-0": `{
								"Version": "2012-10-17",
								"Statement": [
								  {
									"Effect": "Allow",
									"Action": [
									  "sqs:ReceiveMessage"
									],
									"Resource": [
									  "arn:aws:sqs:us-east-1:554248189203:dev-sqs-demo-nmv5z-b660bd5f"
									]
								  }
								]
							  }`,
						},
						ServiceAccounts: []*v1alpha1.ServiceAccount{
							{
								Name:      "dev-sa",
								Namespace: "dev",
								Action:    v1alpha1.ServiceAccountActionUpdate,
							},
						},
					},
				},
				Status: v1alpha1.WorkloadIdentityStatus{
					Name: "ccs-v1",
					ID:   "arn:aws:iam::12345678:role/ccs-v1",
				},
			},
			setupMockExpectations: func(iamClient *mocks.IAM, stsClient *mocks.STS) {
				stsClient.On("GetCallerIdentity", &sts.GetCallerIdentityInput{}).Return(
					&sts.GetCallerIdentityOutput{
						Account: aws.String("12345678"),
					}, nil)

				iamClient.On("GetRole", &iam.GetRoleInput{
					RoleName: aws.String("ccs-v1"),
				}).Return(&iam.GetRoleOutput{
					Role: &iam.Role{
						AssumeRolePolicyDocument: aws.String(`{
								"Version": "2012-10-17",
								"Statement": [
								  {
									"Effect": "Allow",
									"Principal": {
									  "Federated": "arn:aws:iam::554248189203:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/B83A70562E3AD70581267BB1DFC27CE7"
									},
									"Action": "sts:AssumeRoleWithWebIdentity",
									"Condition": {
									  "StringEquals": {
										"oidc.eks.us-east-1.amazonaws.com/id/B83A70562E3AD70581267BB1DFC27CE7:sub": "system:serviceaccount:dev:ccs-v1"
									  }
									}
								  }`),
						Arn: aws.String("arn:aws:iam::12345678:role/ccs-v1"),
					},
				}, nil).Twice()

				iamClient.On("UpdateAssumeRolePolicy",
					mock.MatchedBy(func(in *iam.UpdateAssumeRolePolicyInput) bool {
						return *in.RoleName == "ccs-v1"
					})).Return(nil, nil)

				iamClient.On("ListRolePoliciesPages",
					mock.MatchedBy(func(in *iam.ListRolePoliciesInput) bool {
						return *in.RoleName == "ccs-v1"
					}),
					mock.AnythingOfType("func(*iam.ListRolePoliciesOutput, bool) bool")).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(func(*iam.ListRolePoliciesOutput, bool) bool)
					out := &iam.ListRolePoliciesOutput{
						PolicyNames: []*string{aws.String(`{
								"Version": "2012-10-17",
								"Statement": [
								  {
									"Effect": "Allow",
									"Action": [
									  "kms:DescribeKey",
									  "kms:GenerateDataKey",
									  "kms:Decrypt"
									],
									"Resource": [
									  "arn:aws:kms:us-east-1:12345678:key/27742849-43f5-c2fd-8930-0708b4ae6534"
									]
								  }
								]
							  }`)},
					}
					arg(out, true)
				})

				iamClient.On("PutRolePolicy",
					mock.MatchedBy(func(in *iam.PutRolePolicyInput) bool {
						return *in.RoleName == "ccs-v1"
					})).Return(nil, nil)

				iamClient.On("DeleteRolePolicy", mock.MatchedBy(func(in *iam.DeleteRolePolicyInput) bool {
					return *in.RoleName == "ccs-v1"
				})).Return(nil, nil)

				iamClient.On("ListAttachedRolePoliciesPages",
					mock.MatchedBy(func(in *iam.ListAttachedRolePoliciesInput) bool {
						return *in.RoleName == "ccs-v1"
					}),
					mock.AnythingOfType("func(*iam.ListAttachedRolePoliciesOutput, bool) bool")).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		k8sClient, err := getTestClient()
		require.Nil(t, err)

		awsIAMClient := &mocks.IAM{}
		awsSTSClient := &mocks.STS{}
		options := &awsx.Options{}
		testCase.setupMockExpectations(awsIAMClient, awsSTSClient)

		iamClient, err := iamx.New(awsIAMClient, awsSTSClient, testCase.workloadIdentity, options)
		assert.Nil(t, err)
		awsRec := &RoleReconciler{
			Client:    k8sClient,
			scheme:    k8sClient.Scheme(),
			res:       testCase.workloadIdentity,
			iamClient: iamClient,
		}

		err = newNamespace(k8sClient, testCase.workloadIdentity.Spec.AWS.ServiceAccounts[0].Namespace)
		assert.Nil(t, err)

		if testCase.workloadIdentity.Spec.AWS.ServiceAccounts[0].Action != v1alpha1.ServiceAccountActionCreate {
			err = newServiceAccount(k8sClient, testCase.workloadIdentity.Spec.AWS.ServiceAccounts[0].Name, testCase.workloadIdentity.Spec.AWS.ServiceAccounts[0].Namespace, testCase.workloadIdentity.Name)
			assert.Nil(t, err)
		}
		err = awsRec.Reconcile(context.Background())
		assert.Nil(t, err)

		sa, err := getServiceAccount(k8sClient, testCase.workloadIdentity.Spec.AWS.ServiceAccounts[0].Name, testCase.workloadIdentity.Spec.AWS.ServiceAccounts[0].Namespace)
		assert.Nil(t, err)

		assert.Equal(t, testCase.workloadIdentity.Status.ID, sa.Annotations[eksServiceAccountAnnotationKey])
		assert.Equal(t, testCase.workloadIdentity.Spec.Name, testCase.workloadIdentity.Status.Name)

		if testCase.workloadIdentity.Spec.AWS.ServiceAccounts[0].Action == v1alpha1.ServiceAccountActionCreate {
			assert.Equal(t, testCase.workloadIdentity.ObjectMeta.Name, sa.ObjectMeta.OwnerReferences[0].Name)
			assert.Equal(t, testCase.workloadIdentity.ObjectMeta.UID, sa.ObjectMeta.OwnerReferences[0].UID)
		}
	}
}

func TestGetConfig(t *testing.T) {
	wi := &v1alpha1.WorkloadIdentity{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ccs-v1",
			UID:  "0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba",
		},
		Spec: v1alpha1.WorkloadIdentitySpec{
			Name:     "ccs-v1",
			Provider: v1alpha1.ProviderAWS,
			Credentials: &identitymanageriov1alpha1.Credentials{
				Source: "Secret",
				SecretRef: &identitymanageriov1alpha1.SecretRef{
					Name:      "test-secret",
					Namespace: "dev",
				},
			},
		},
	}

	k8sClient, err := getTestClient()
	require.Nil(t, err)
	region := "dXMtZWFzdC0xCg=="
	awsAccessKeyID := "U0wxbk0yeHhqYW1RSzZCSUU3N3Q4YzRCTStDUlFibAo="
	// #nosec
	awsSecretAccessKey := "REhhdTg4L1RNSUpzRGhYOWR5cnhBbzFwOEkyaDgzb1JhMndSbXUK"

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "dev",
		},
		Data: map[string][]byte{
			"region":                []byte(region),
			"aws_access_key_id":     []byte(awsAccessKeyID),
			"aws_secret_access_key": []byte(awsSecretAccessKey),
		},
	}

	err = newNamespace(k8sClient, "dev")
	assert.Nil(t, err)

	err = k8sClient.Create(context.Background(), secret)
	require.Nil(t, err)

	awsRec := &RoleReconciler{
		Client: k8sClient,
		scheme: k8sClient.Scheme(),
		res:    wi,
	}

	conf, err := awsRec.getConfig(context.Background())
	require.Nil(t, err)

	assert.Equal(t, region, conf.Region)
	assert.Equal(t, awsAccessKeyID, conf.AccessKeyID)
	assert.Equal(t, awsSecretAccessKey, conf.SecretAccessKey)
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
			Labels:    map[string]string{managedByLabelKey: managedByValueKey, roleLabelKey: resName},
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
