# Quick start

This guide helps you to setup the cluster, install Identity Manager, deploy a workload identity and verify the working of the Identity Manager with a demo application. The following installation steps are applicable for AWS EKS cluster.

> Note: If you already have a cluster setup ready, you can refer to the [Getting started](guides-getting-started.md) guide.

## Prerequisites

The following tools are needed to be installed in your system:

1. `eksctl` - To create a cluster in AWS EKS
2. `Helm` - To install the Identity Manager
3. `kubectl` - To connect with kubernetes cluster

## Create an EKS cluster

The following instructions allows you to create a cluster in AWS EKS. They also takes care of creation of a namespace, service account and creation of a role with the IAM policies that are required by the Identity Manager and binds the role to the service account that is to be created.

1. The following eks config file helps you to create a cluster with the required IAM role and policies, an m5.large node in the `us-east-1` region. The mentioned policies are the minimum required permissions for Identity Manager. You can change the other config as per your preference. Save the following yaml as `eks-identity-manager.yaml`.
``` yaml
--8<-- "examples/eks-identity-manager.yaml"
```
2. Create the cluster using eksctl:  
`eksctl create cluster -f eks-identity-manager.yaml`

This will create an EKS cluster with the specified spec.

## Installing Identity Manager with Helm

1. Copy the cluster's IAM role: 
If the cluster is newly created using `eksctl`, copy the IAM role ARN of the cluster from CloudFormation template. The IAM role can be found in the annotation of the service account `identity-manager`.
``` bash
export IAM_ROLE=<IAM role>
```
2. Install Identity Manager
``` bash
helm repo add invisibl https://charts.invisibl.io

helm install my-identity-manager invisibl/identity-manager  --set provider.aws.enabled=true --set provider.aws.arn=$IAM_ROLE --set serviceAccount.create=false --set serviceAccount.name=identity-manager --namespace=identity-manager
```

The above command will install Identity Manager in `identity-manager` namespace and `identity-manager` service account. This service account is annotated to the IAM role that has the necessary IAM permissions for the Identity Manager. This service account is dedicated to Identity Manager and any workload identity should be deployed in another service accounts.


## Deploying demo application

1. Get your AWS account ID
```bash
export ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
```

2. Get your OIDC provider
```bash
export OIDC_PROVIDER=$(aws eks describe-cluster --name identity-manager-test --query "cluster.identity.oidc.issuer" --region us-east-1 --output text | sed -e "s/^https:\/\///")
```

3. Deploy demo application
``` bash
helm install my-identity-manager-demo invisibl/identity-manager-demo --set serviceAccount.name=sa-demo --namespace=demo --set  workloadIdentity.aws.accountId=${ACCOUNT_ID} --set workloadIdentity.aws.oidcProvider=${OIDC_PROVIDER} --create-namespace
```
The above command will deploy the workload identity and a demo application in the namespace `demo`. The identity manager will create an IAM role `my-identity-manager-demo` in AWS with the inline polices mentioned in the workload identity attached to the IAM role. The identity manager will also create a service account `sa-demo` and will annotate it with the newly created role to facilitate the role binding.

4. Verify the role binding for service account
``` bash
kubectl get serviceaccount sa-demo -n demo  -o=jsonpath='{.metadata.annotations}'
```
The response should look similar to the below one:
``` bash
{"eks.amazonaws.com/role-arn":"arn:aws:iam::<Account ID>:role/my-identity-manager-demo"}
```
5. Check the logs of the demo application pod and it should list your EC2 instances using the new role 
`my-identity-manager-demo`.
``` bash
time="2022-05-04T09:54:04Z" level=info msg="STS:"
time="2022-05-04T09:54:04Z" level=info msg="STS ARN: arn:aws:sts::<Account ID>:assumed-role/my-identity-manager-demo/48520678505362540620424"
time="2022-05-04T09:54:04Z" level=info msg="EC2:"
time="2022-05-04T09:54:04Z" level=info msg="Reservation ID: r-0532a81dd8ed78de1"
time="2022-05-04T09:54:04Z" level=info msg="Instance ID: i-078e85384f15b27b9"
time="2022-05-04T09:54:04Z" level=info msg="Reservation ID: r-0ec5be0a1e1017088"
time="2022-05-04T09:54:04Z" level=info msg="Instance ID: i-0efee718f18c10742"
time="2022-05-04T09:54:04Z" level=info msg="Reservation ID: r-0992e8b92ae857ddd"
time="2022-05-04T09:54:04Z" level=info msg="Instance ID: i-097b6eb735190898c"
time="2022-05-04T09:54:04Z" level=info msg="Reservation ID: r-0f7c4e3a8d62c0af7"
time="2022-05-04T09:54:04Z" level=info msg="Instance ID: i-09e2a542b827858de"
time="2022-05-04T09:54:04Z" level=info msg="Instance ID: i-0c77422d4e56c42c9"
time="2022-05-04T09:54:04Z" level=info msg="Reservation ID: r-0265da1370d12b44d"
time="2022-05-04T09:54:04Z" level=info msg="Instance ID: i-05d02af34a271e308"   
time="2022-05-04T09:54:04Z" level=info msg="Reservation ID: r-0efa1b0178917b544"
time="2022-05-04T09:54:04Z" level=info msg="Instance ID: i-04027edd12c82f6d6"
time="2022-05-04T09:54:04Z" level=info msg="Reservation ID: r-0b0d08c57fb60bb71"
time="2022-05-04T09:54:04Z" level=info msg="Instance ID: i-09d4114c468def93e"
time="2022-05-04T09:54:04Z" level=info msg="Reservation ID: r-09e8cc98f64abef83"
time="2022-05-04T09:54:04Z" level=info msg="Instance ID: i-0bc14ad13ef223d76"
time="2022-05-04T09:54:04Z" level=info
time="2022-05-04T09:54:04Z" level=info msg="Reservations count: 8"
time="2022-05-04T09:54:04Z" level=info msg="Instances count: 9"
```

## Uninstalling demo application with Helm

```bash
helm uninstall my-identity-manager-demo -n demo
```

## Uninstalling Identity Manager with Helm

```bash
helm uninstall my-identity-manager -n identity-manager
```
