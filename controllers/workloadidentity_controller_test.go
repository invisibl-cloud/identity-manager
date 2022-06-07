package controllers

import (
	"context"

	"github.com/invisibl-cloud/identity-manager/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("WorkloadIdentity Controller", func() {

	BeforeEach(func() {
		sa := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}
		err := k8sClient.Create(context.Background(), sa)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should handle the workload identity as expected", func() {
		spec := v1alpha1.WorkloadIdentitySpec{
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
						Name:      "test-sa",
						Namespace: "test",
						Action:    v1alpha1.ServiceAccountActionCreate,
					},
				},
			},
		}

		workloadIdentity := &v1alpha1.WorkloadIdentity{
			ObjectMeta: metav1.ObjectMeta{
				Name:      spec.Name,
				Namespace: spec.AWS.ServiceAccounts[0].Namespace,
			},
			Spec: spec,
		}

		By("Creating the workload identity successfully")
		Expect(k8sClient.Create(context.Background(), workloadIdentity)).Should(Succeed())
		wi := &v1alpha1.WorkloadIdentity{}
		err := k8sClient.Get(context.Background(), types.NamespacedName{Name: spec.Name, Namespace: "test"}, wi)
		Expect(err).ToNot(HaveOccurred())
		Expect(wi.Spec.Name).To(Equal(spec.Name))

		By("Updating the workload identity successfully")
		workloadIdentity.Spec.AWS.ServiceAccounts = []*v1alpha1.ServiceAccount{
			{
				Name:      "test-sa-update",
				Namespace: "test",
				Action:    v1alpha1.ServiceAccountActionUpdate,
			},
		}
		Expect(k8sClient.Update(context.Background(), workloadIdentity)).Should(Succeed())
		Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: spec.Name, Namespace: "test"}, wi)).Should(Succeed())
		Expect(wi.Spec.AWS.ServiceAccounts[0].Name).To(Equal(workloadIdentity.Spec.AWS.ServiceAccounts[0].Name))

		By("Deleting the scope")
		Expect(k8sClient.Delete(context.Background(), workloadIdentity)).Should(Succeed())
		Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: spec.Name, Namespace: "test"}, wi)).ShouldNot(Succeed())
	})

})
