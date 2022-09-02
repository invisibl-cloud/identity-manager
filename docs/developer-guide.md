Developer guide
=================================================

## Getting Started

You must have a working [Go environment](https://go.dev/doc/install/) and then clone the repo:

```shell
git clone https://github.com/invisibl-cloud/identity-manager.git
cd identity-manager
```

If you want to run controller tests you also need to install kubebuilder's `envtest`.

The recommended way to do so is to install [setup-envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/tools/setup-envtest)

Here is an example on how to set it up:

```
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

# list available versions
setup-envtest list --os $(go env GOOS) --arch $(go env GOARCH)

# To use a specific version
setup-envtest use -p path 1.20.2

#To set environment variables
source <(setup-envtest use 1.20.2 -p env --os $(go env GOOS) --arch $(go env GOARCH))

```

for more information, please see [setup-envtest docs](https://github.com/kubernetes-sigs/controller-runtime/tree/master/tools/setup-envtest)

## Building & Testing

The project uses the `make` build system. It'll run code generators, tests and
static code analysis.

Building the operator binary and docker image:

```shell
make build
make docker-build IMG=identity-manager:latest
```

Run tests and lint the code:
```shell
make test
make lint
```

## Installing

To install the Identity Manager Operator into a Kubernetes Cluster run:

```shell
helm repo add invisibl https://charts.invisibl.io
helm repo update
helm install identity-manager invisibl/identity-manager
```

You can alternatively run the controller on your host system for development purposes:


```shell
make install
make run
```

## Deploy WorkloadIdentity

### Create a secret containing your AWS credentials

```shell
echo -n 'KEYID' > ./access-key
echo -n 'SECRETKEY' > ./secret-access-key
echo -n 'REGION' > ./region
kubectl create secret generic aws-credentials --from-file=./access-key  --from-file=./secret-access-key --from-file=./region
```

### Apply Workload Identity

Save the following workload identity yaml in `workload-identity.yaml`.

``` yaml
--8<-- "examples/basic-workload-identity.yaml"
```

Once the workload identity is applied, an IAM role will be created in AWS with the policies attached to the role.

To remove the CRDs run:

```shell
make uninstall
```

## Documentation

We use [mkdocs material](https://squidfunk.github.io/mkdocs-material/) to generate this
documentation. See `/docs` for the source code.

When writing documentation it is advised to run the mkdocs server with live reload:

```shell
mkdocs serve
```

Open `http://localhost:8000` in your browser.
