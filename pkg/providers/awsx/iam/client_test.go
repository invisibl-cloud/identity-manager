package iam

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	v1alpha1 "github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	mocks "github.com/invisibl-cloud/identity-manager/pkg/mocks/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRoleName(t *testing.T) {
	testCases := []struct {
		desc             string
		workloadIdentity *v1alpha1.WorkloadIdentity
		expectedRoleName string
	}{
		{
			desc: "Happy path test case",
			workloadIdentity: &v1alpha1.WorkloadIdentity{
				Spec: v1alpha1.WorkloadIdentitySpec{
					Name: "S3InventoryRole",
				},
			},
			expectedRoleName: "S3InventoryRole",
		},
		{
			desc: "Spec.Name is empty",
			workloadIdentity: &v1alpha1.WorkloadIdentity{
				ObjectMeta: v1.ObjectMeta{
					Name:      "SomeInternalName",
					Namespace: "dev",
				},
			},
			expectedRoleName: "dev-SomeInternalName",
		},
		{
			desc: "Role name with '-' as suffix",
			workloadIdentity: &v1alpha1.WorkloadIdentity{
				Spec: v1alpha1.WorkloadIdentitySpec{
					Name: "S3InventoryRole-",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "SomeInternalName",
					Namespace: "dev",
				},
			},
			expectedRoleName: "S3InventoryRole-dev-SomeInternalName",
		},
	}

	for _, testCase := range testCases {
		client := &Client{
			role: testCase.workloadIdentity,
		}
		roleName := client.roleName()

		assert.Equal(t, testCase.expectedRoleName, roleName)
	}
}

func TestToInlinePolicyNames(t *testing.T) {
	inlinePolicies := map[string]string{
		"S3ReadPolicy": `{"policy":"arn:policy"}`,
		"S3ReadPolicy-eks-dev-wlk4-B1267BBA7830580562E3AD71DFC27CE7-2E3AD71DFC27CE7B1267BBA783058056-0562E3AD71DFC27CE7-2E3AD71DF-BA7830580562E3AD71DF": `{"policy":"arn:policy"}`,
	}
	expectedPolicyNames := []string{"S3ReadPolicy-3ceca7bf0f96b81a74f9af5ac4f60012",
		"S3ReadPolicy-eks-dev-wlk4-B1267BBA7830580562E3AD71DFC27CE7-2E3AD71DFC27CE7B1267BBA783058056-0562E3AD71DFC27CE7-2E3AD71DF-BA78305",
	}
	expectedPolicies := map[string]string{
		"S3ReadPolicy-3ceca7bf0f96b81a74f9af5ac4f60012": "S3ReadPolicy",
		"S3ReadPolicy-eks-dev-wlk4-B1267BBA7830580562E3AD71DFC27CE7-2E3AD71DFC27CE7B1267BBA783058056-0562E3AD71DFC27CE7-2E3AD71DF-BA78305": "S3ReadPolicy-eks-dev-wlk4-B1267BBA7830580562E3AD71DFC27CE7-2E3AD71DFC27CE7B1267BBA783058056-0562E3AD71DFC27CE7-2E3AD71DF-BA7830580562E3AD71DF",
	}

	pols, m := toInlinePolicyNames(inlinePolicies)
	assert.Equal(t, expectedPolicyNames, pols)
	assert.Equal(t, expectedPolicies, m)
}

func TestIsArn(t *testing.T) {
	testCases := []struct {
		policy         string
		expectedResult bool
	}{
		{
			policy:         "IAMReadOnlyAccess",
			expectedResult: false,
		},
		{
			policy:         "arn:aws:iam::aws:policy/AWSDirectConnectReadOnlyAccess",
			expectedResult: true,
		},
	}

	for _, testCase := range testCases {
		isArn := isArn(testCase.policy)
		assert.Equal(t, testCase.expectedResult, isArn)
	}
}

func TestToArns(t *testing.T) {
	testCases := []struct {
		policy      string
		expectedArn string
	}{
		{
			policy:      "IAMReadOnlyAccess",
			expectedArn: "arn:aws:iam::12345678:policy/IAMReadOnlyAccess",
		},
		{
			policy:      "arn:aws:iam::aws:policy/AWSDirectConnectReadOnlyAccess",
			expectedArn: "arn:aws:iam::aws:policy/AWSDirectConnectReadOnlyAccess",
		},
	}

	stsClient := &mocks.STS{}
	client := &Client{
		sts: stsClient,
	}
	stsClient.On("GetCallerIdentity", &sts.GetCallerIdentityInput{}).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("12345678"),
	}, nil)

	for _, testCase := range testCases {
		arn, err := client.getArn(testCase.policy)
		assert.Nil(t, err)
		assert.Equal(t, testCase.expectedArn, arn)
	}
}

func TestCreateOrUpdate(t *testing.T) {
	testCases := []struct {
		desc                  string
		workloadIdentity      *v1alpha1.WorkloadIdentity
		setupMockExpectations func(*mocks.IAM, *mocks.STS)
		expectedRoleStatus    *RoleStatus
	}{
		{
			desc: "Create new role and policies",
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
					},
				},
			},
			setupMockExpectations: func(iamClient *mocks.IAM, stsClient *mocks.STS) {
				stsClient.On("GetCallerIdentity", &sts.GetCallerIdentityInput{}).Return(
					&sts.GetCallerIdentityOutput{
						Account: aws.String("12345678"),
					}, nil)

				iamClient.On("GetRole", &iam.GetRoleInput{
					RoleName: aws.String("ccs-v1"),
				}).Return(nil, errors.New("role not found in AWS"))

				iamClient.On("CreateRole", mock.MatchedBy(func(in *iam.CreateRoleInput) bool {
					return *in.RoleName == "ccs-v1"
				})).Return(
					&iam.CreateRoleOutput{
						Role: &iam.Role{
							Arn:      aws.String("arn:aws:iam::12345678:role/ccs-v1"),
							RoleName: aws.String("ccs-v1"),
						},
					}, nil,
				)

				iamClient.On("PutRolePolicy", mock.MatchedBy(func(in *iam.PutRolePolicyInput) bool {
					return *in.RoleName == "ccs-v1"
				})).Return(nil, nil)

			},
			expectedRoleStatus: &RoleStatus{
				Name: "ccs-v1",
				ARN:  "arn:aws:iam::12345678:role/ccs-v1",
			},
		},
		{
			desc: "Update existing role and policies",
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
			expectedRoleStatus: &RoleStatus{
				Name: "ccs-v1",
				ARN:  "arn:aws:iam::12345678:role/ccs-v1",
			},
		},
	}

	for _, testCase := range testCases {
		iamClient := &mocks.IAM{}
		stsClient := &mocks.STS{}
		testCase.setupMockExpectations(iamClient, stsClient)

		client, err := New(iamClient, stsClient, testCase.workloadIdentity)
		assert.Nil(t, err)

		status, err := client.CreateOrUpdate(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, testCase.expectedRoleStatus.Name, status.Name)
		assert.Equal(t, testCase.expectedRoleStatus.ARN, status.ARN)
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		desc                  string
		workloadIdentity      *v1alpha1.WorkloadIdentity
		setupMockExpectations func(*mocks.IAM, *mocks.STS)
	}{
		{
			desc: "Delete the role and policis",
			workloadIdentity: &v1alpha1.WorkloadIdentity{
				Spec: v1alpha1.WorkloadIdentitySpec{
					Name:     "ccs-v2",
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
					},
				},
				Status: v1alpha1.WorkloadIdentityStatus{
					Name: "ccs-v2",
					ID:   "arn:aws:iam::12345678:role/ccs-v2",
				},
			},
			setupMockExpectations: func(iamClient *mocks.IAM, stsClient *mocks.STS) {
				stsClient.On("GetCallerIdentity", &sts.GetCallerIdentityInput{}).Return(&sts.GetCallerIdentityOutput{
					Account: aws.String("12345678"),
				}, nil)

				iamClient.On("GetRole", &iam.GetRoleInput{
					RoleName: aws.String("ccs-v2"),
				}).Return(&iam.GetRoleOutput{
					Role: &iam.Role{
						RoleName: aws.String("ccs-v1"),
						Arn:      aws.String("arn:aws:iam::12345678:role/ccs-v1"),
					},
				}, nil)

				iamClient.On("ListRolePoliciesPages",
					mock.MatchedBy(func(in *iam.ListRolePoliciesInput) bool {
						return *in.RoleName == "ccs-v2"
					}),
					mock.AnythingOfType("func(*iam.ListRolePoliciesOutput, bool) bool")).Return(nil)

				iamClient.On("ListAttachedRolePoliciesPages",
					mock.MatchedBy(func(in *iam.ListAttachedRolePoliciesInput) bool {
						return *in.RoleName == "ccs-v2"
					}),
					mock.AnythingOfType("func(*iam.ListAttachedRolePoliciesOutput, bool) bool")).Return(nil)

				iamClient.On("DeleteRole", &iam.DeleteRoleInput{
					RoleName: aws.String("ccs-v2"),
				}).Return(&iam.DeleteRoleOutput{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		iamClient := &mocks.IAM{}
		stsClient := &mocks.STS{}
		testCase.setupMockExpectations(iamClient, stsClient)

		client, err := New(iamClient, stsClient, testCase.workloadIdentity)
		assert.Nil(t, err)

		err = client.Delete(context.Background())
		assert.Nil(t, err)
	}
}

func TestListInlinePolicies(t *testing.T) {
	iamClient := &mocks.IAM{}
	stsClient := &mocks.STS{}

	stsClient.On("GetCallerIdentity", &sts.GetCallerIdentityInput{}).Return(
		&sts.GetCallerIdentityOutput{
			Account: aws.String("12345678"),
		}, nil)

	iamClient.On("ListRolePoliciesPages",
		mock.MatchedBy(func(in *iam.ListRolePoliciesInput) bool {
			return *in.RoleName == "ccs-v2"
		}),
		mock.AnythingOfType("func(*iam.ListRolePoliciesOutput, bool) bool")).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(func(*iam.ListRolePoliciesOutput, bool) bool)
		// arg["foo"] = "bar"
		out := &iam.ListRolePoliciesOutput{
			PolicyNames: []*string{aws.String("pol1"), aws.String("pol2")},
		}
		arg(out, true)
	})

	client, err := New(iamClient, stsClient, &v1alpha1.WorkloadIdentity{})
	assert.Nil(t, err)

	pols, err := client.listInlinePolicies("ccs-v2")
	assert.Nil(t, err)

	assert.Equal(t, []string{"pol1", "pol2"}, pols)
}
