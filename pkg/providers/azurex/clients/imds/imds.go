package imds

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	timeout = 80 * time.Second
)

// Check performs sanity check on Azure client
func Check(ctx context.Context, clientID string) {
	log := log.FromContext(ctx)
	resourceEndpoint := os.Getenv("AZURE_RESOURCE_ENDPOINT")
	if resourceEndpoint == "" {
		resourceEndpoint = "https://management.azure.com/"
	}

	str, err := CurlIMDSMetadataInstanceEndpoint()
	if err != nil {
		log.Info("CurlIMDSMetadataInstanceEndpoint", "error", err)
	} else {
		log.Info("CurlIMDSMetadataInstanceEndpoint", "str", str)
	}
	t1, err := GetTokenFromIMDS(resourceEndpoint)
	if err != nil {
		log.Info("GetTokenFromIMDS", "error", err)
	}
	t2, err := GetTokenFromIMDSWithUserAssignedID(resourceEndpoint, clientID)
	if err != nil {
		log.Info("GetTokenFromIMDSWithUserAssignedID", "error", err)
	}
	if t1 == nil || t2 == nil || !strings.EqualFold(t1.AccessToken, t2.AccessToken) {
		log.Info("Tokens acquired from IMDS with and without identity client ID do not match")
	}
	if t1 != nil {
		log.Info("Try decoding your token %s at https://jwt.io", t1.AccessToken)
	}
}

// CurlIMDSMetadataInstanceEndpoint performs get request to
// IMDS metadata instance endpoint and returns the result
func CurlIMDSMetadataInstanceEndpoint() (string, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/instance?api-version=2017-08-01", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
	//klog.Infof(`curl -H Metadata:true "http://169.254.169.254/metadata/instance?api-version=2017-08-01": %s`, body)
}

// GetTokenFromIMDS get service principal token for the resource
func GetTokenFromIMDS(resourceName string) (*adal.Token, error) {
	managedIdentityOpts := &adal.ManagedIdentityOptions{}
	spt, err := adal.NewServicePrincipalTokenFromManagedIdentity(resourceName, managedIdentityOpts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := spt.RefreshWithContext(ctx); err != nil {
		return nil, err
	}

	token := spt.Token()
	if token.IsZero() {
		return nil, fmt.Errorf("%+v is a zero token", token)
	}

	return &token, nil
}

// GetTokenFromIMDSWithUserAssignedID receives resource and identity client ID
// and returns the service principal token
func GetTokenFromIMDSWithUserAssignedID(resourceName string, identityClientID string) (*adal.Token, error) {
	managedIdentityOpts := &adal.ManagedIdentityOptions{ClientID: identityClientID}
	spt, err := adal.NewServicePrincipalTokenFromManagedIdentity(resourceName, managedIdentityOpts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := spt.RefreshWithContext(ctx); err != nil {
		return nil, err
	}

	token := spt.Token()
	if token.IsZero() {
		return nil, fmt.Errorf("%+v is a zero token", token)
	}

	return &token, nil
}
