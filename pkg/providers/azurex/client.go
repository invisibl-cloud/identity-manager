package azurex

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/gofrs/uuid"
	"github.com/invisibl-cloud/identity-manager/pkg/util"
)

const (
	// UserAgent is the user agent addition that identifies the Azure client
	UserAgent = "identity-manager-azure-client"
)

// Option is the function syntax definition that is
// used as a generic return type for config options
type Option func(*Client) error

// WithConfigMapFunc executes fn and returns Option with
// fn's returned map value
func WithConfigMapFunc(fn func() (map[string]any, error)) Option {
	return func(x *Client) error {
		data, err := fn()
		if err != nil {
			return err
		}
		return WithConfigMap(data)(x)
	}
}

// WithConfigData expects data and returns
// Option with a populated map
func WithConfigData(data []byte) Option {
	return func(x *Client) error {
		m := map[string]any{}
		err := json.Unmarshal(data, &m)
		if err != nil {
			return err
		}
		return WithConfigMap(m)(x)
	}
}

// WithEnv populates env data to Client
// and returns Option
func WithEnv() Option {
	return func(x *Client) error {
		if x.config == nil {
			x.config = &Config{}
		}
		// base config
		x.config.SubscriptionID = util.DefaultString(os.Getenv("AZURE_SUBSCRIPTION_ID"), x.config.SubscriptionID)
		x.config.TenantID = util.DefaultString(os.Getenv("AZURE_TENANT_ID"), x.config.TenantID)
		x.config.ClientID = util.DefaultString(os.Getenv("AZURE_CLIENT_ID"), x.config.ClientID)
		x.config.ClientSecret = util.DefaultString(os.Getenv("AZURE_CLIENT_SECRET"), x.config.ClientSecret)
		x.config.Environment = util.DefaultString(os.Getenv("AZURE_ENVIRONMENT"), x.config.Environment)
		// custom config
		x.config.Location = util.DefaultString(os.Getenv("AZURE_LOCATION"), x.config.Location)
		x.config.ResourceGroup = util.DefaultString(os.Getenv("AZURE_RESOURCE_GROUP"), x.config.ResourceGroup)
		return nil
	}
}

// WithConfigMap expects map[string]interface
// and returns Option
func WithConfigMap(m map[string]any) Option {
	return func(x *Client) error {
		// if region found, set it in location
		if region, ok := m["region"]; ok {
			m["location"] = region
		}
		if x.config == nil {
			x.config = &Config{}
		}
		data, err := json.Marshal(m)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, x.config)
	}
}

// WithConfig expects Config and
// returns Option
func WithConfig(c *Config) Option {
	return func(x *Client) error {
		x.config = c
		return nil
	}
}

// New initializes Client with multiple Option
func New(opts ...Option) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	authorizer, err := c.config.GetAuthorizer()
	if err != nil {
		return nil, err
	}
	c.authorizer = authorizer
	return c, nil
}

// Client is the definition of the Client object
type Client struct {
	config     *Config
	authorizer autorest.Authorizer
}

// GetAuthorizer returns client's authorizer
func (x *Client) GetAuthorizer() autorest.Authorizer {
	return x.authorizer
}

// GetConfig returns client's config
func (x *Client) GetConfig() *Config {
	return x.config
}

// Config - simple aws session config
type Config struct {
	// AzureEnv
	Environment string `json:"environment" yaml:"environment"`
	// IDs
	TenantID       string `json:"tenantId" yaml:"tenantId"`
	SubscriptionID string `json:"subscriptionId" yaml:"subscriptionId"`
	// Auth1
	ClientID     string `json:"clientId" yaml:"clientId"`
	ClientSecret string `json:"clientSecret" yaml:"clientSecret"`
	// Auth2
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	// Auth3
	Certificate         string `json:"certificate" yaml:"certificate"`
	CertificatePath     string `json:"certificatePath" yaml:"certificatePath"`
	CertificatePassword string `json:"certificatePassword" yaml:"certificatePassword"`
	// Defaults
	ResourceGroup string `json:"resourceGroup" yaml:"resourceGroup"`
	Location      string `json:"location" yaml:"location"`
	// ???
	ADResource string `json:"adResource" yaml:"adResource"`
	// ???
	UseManagedIdentityExtension bool   `json:"useManagedIdentityExtension" yaml:"useManagedIdentityExtension"`
	UserAssignedIdentityID      string `json:"userAssignedIdentityID" yaml:"userAssignedIdentityID"`
	UseDeviceFlow               bool   `json:"useDeviceFlow" yaml:"useDeviceFlow"`
	// URLs
	ActiveDirectoryEndpointURL     string `json:"activeDirectoryEndpointUrl" yaml:"activeDirectoryEndpointUrl"`
	ResourceManagerEndpointURL     string `json:"resourceManagerEndpointUrl" yaml:"resourceManagerEndpointUrl"`
	ActiveDirectoryGraphResourceID string `json:"activeDirectoryGraphResourceId" yaml:"activeDirectoryGraphResourceId"`
	SQLManagementEndpointURL       string `json:"sqlManagementEndpointUrl,omitempty" yaml:"sqlManagementEndpointUrl,omitempty"`
	GalleryEndpointURL             string `json:"galleryEndpointUrl,omitempty" yaml:"galleryEndpointUrl,omitempty"`
	ManagementEndpointURL          string `json:"managementEndpointUrl,omitempty" yaml:"managementEndpointUrl,omitempty"`
	//authorizationServerURL string
	//keepResources          bool
	//groupName              string // deprecated, use baseGroupName instead
	//baseGroupName          string
	//userAgent              string
	//cloudName              string = "AzurePublicCloud"
}

// DefaultConfigMap is the config map with
// default settings
var DefaultConfigMap = map[string]string{
	"environment":                    azure.PublicCloud.Name,
	"activeDirectoryEndpointUrl":     azure.PublicCloud.ActiveDirectoryEndpoint,
	"resourceManagerEndpointUrl":     azure.PublicCloud.ResourceManagerEndpoint,
	"activeDirectoryGraphResourceId": azure.PublicCloud.ActiveDirectoryEndpoint,
	"galleryEndpointUrl":             azure.PublicCloud.GalleryEndpoint,
	"managementEndpointUrl":          azure.PublicCloud.ServiceManagementEndpoint,
}

// ToUUIDString expects uuid and returns string
func ToUUIDString(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}

// ToString returns underlying string of s
// It returns "" if s is nil
func ToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ToStringPtr returns pointer to s
func ToStringPtr(s string) *string {
	return &s
}

// IsNotFound returns a value indicating whether the given error represents that the resource was not found.
func IsNotFound(err error) bool {
	detailedError, ok := err.(autorest.DetailedError)
	if !ok {
		return false
	}

	statusCode, ok := detailedError.StatusCode.(int)
	if !ok {
		return false
	}

	return statusCode == http.StatusNotFound
}
