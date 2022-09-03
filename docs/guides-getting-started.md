# Getting started

This guide helps you to setup the cluster, install Identity Manager, deploy a workload identity and verify the working with a demo application. The following installation steps are applicable for AWS EKS cluster.

> Note: The minimum supported version of Kubernetes is `1.16.0`.

## Prerequisites

The following tools are needed to be installed in your system:

1. `Helm` - To install the Identity Manager
2. `kubectl` - To connect with kubernetes cluster

## Working with an existing cluster
For an existing cluster, a new namespace, a service account and an IAM role need to be created. A few policies that the Identity Manager needs for its working need to be attached to the newly created role.

1. Create the namespace and service account
```bash
kubectl create namespace identity-manager
kubectl create serviceaccount identity-manager -n identity-manager
```

2. Export the required environment values:
```bash
export ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)

export OIDC_PROVIDER=$(aws eks describe-cluster --name <cluster-name> --query "cluster.identity.oidc.issuer" --region <region> --output text | sed -e "s/^https:\/\///")

export NAMESPACE=identity-manager
export SERVICEACCNAME=identity-manager
export ROLENAME=identity-manager
```

3. Save the required IAM polices:
```bash
read -r -d '' IDENTITY_POLICIES <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "iam:AttachRolePolicy",
                "iam:CreateRole",
                "iam:DeleteRole",
                "iam:DeleteRolePolicy",
                "iam:DetachRolePolicy",
                "iam:GetRole",
                "iam:ListAttachedRolePolicies",
                "iam:ListRolePolicies",
                "iam:PutRolePolicy",
                "iam:UpdateAssumeRolePolicy",
                "iam:UpdateRole",
                "sts:GetCallerIdentity"
            ],
            "Resource": "*",
            "Effect": "Allow"
        }
    ]
}
EOF
echo "${IDENTITY_POLICIES}" > identity-policies.json
```

4. Save the trust relationship
```bash
read -r -d '' TRUST_RELATIONSHIP <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::${ACCOUNT_ID}:oidc-provider/${OIDC_PROVIDER}"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "${OIDC_PROVIDER}:aud": "sts.amazonaws.com",
          "${OIDC_PROVIDER}:sub": "system:serviceaccount:${NAMESPACE}:${SERVICEACCNAME}"
        }
      }
    }
  ]
}
EOF
echo "${TRUST_RELATIONSHIP}" > trust.json
```

5. Create the IAM role:
```bash
aws iam create-role --role-name ${ROLENAME} --assume-role-policy-document file://trust.json --description "identity-role"
```
Note down the IAM role ARN in the response.

6. Attach the IAM policies to the role:
```bash
aws iam put-role-policy --role-name ${ROLENAME} --policy-name identity-policy --policy-document file://identity-policies.json
```

7. Annotate the service account with the IAM role ARN that was noted in the step 5.
```bash
kubectl annotate serviceaccount -n identity-manager identity-manager \
eks.amazonaws.com/role-arn=<ROLE ARN>
```

At this point, the setup of the service account `identity-manager` is finished and is ready for the Identity Manager installation.

## Installing Identity Manager with Helm

1. Copy the newly created IAM role ARN and create an environment variable:
``` bash
export IAM_ROLE=<IAM role ARN>
```
2. Install Identity Manager
``` bash
helm repo add invisibl https://charts.invisibl.io

helm install my-identity-manager invisibl/identity-manager  --set provider.aws.enabled=true --set provider.aws.arn=$IAM_ROLE --set serviceAccount.create=false --set serviceAccount.name=identity-manager --namespace=identity-manager
```

The above command will install Identity Manager in `identity-manager` namespace and `identity-manager` service account. This service account is annotated to the IAM role that has the necessary IAM permissions for the Identity Manager. This service account is dedicated to Identity Manager and any workload identity should be deployed in other service accounts.

## Deploying demo application

1. Get your AWS account ID
```bash
export ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
```

2. Get your OIDC provider
```bash
export OIDC_PROVIDER=$(aws eks describe-cluster --name <cluster name> --query "cluster.identity.oidc.issuer" --region <region> --output text | sed -e "s/^https:\/\///")
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

## Troubleshooting

1. Identity Manager maintains the most recent log message in the `Status` field of the workload identity. In rare cases where the pods are failing to authenticate to the AWS services, the workload identity's `Status` fields can be queried to view the log which helps in debugging.
```bash
kubectl get workloadidentity my-identity-manager-demo -n demo -o yaml
```
2. If there are no error log message found in the `Status` field of the workload identity, it is worth checking if the namespace and the service account defined in the workload identity matches the trust policy in the `aws.assumeRolePolicy`. The following is an simple example template of the trust policy:
``` json
{
        "Version": "2012-10-17",
        "Statement": [
          {
            "Effect": "Allow",
            "Principal": {
              "Federated": "arn:aws:iam::454135189203:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/33FC4D52E1241A120425308D0853F923A"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
              "StringEquals": {
                "oidc.eks.us-east-1.amazonaws.com/id/33FC4D52E1241A120425308D0853F923A:sub": "system:serviceaccount:<namespace>:<service account>"
              }
            }
          }
        ]
      }
```