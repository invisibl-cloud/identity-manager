# usage: ./generate.sh spec.md

gen-crd-api-reference-docs \
-template-dir . \
-config ./config.json \
-api-dir github.com/invisibl-cloud/identity-manager/api/v1alpha1 \
-out-file $1