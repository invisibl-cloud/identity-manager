package gcpx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/invisibl-cloud/identity-manager/pkg/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Option for Client
type Option func(*Client) error

// WithConfigMapFunc is an option to configure Client Config
func WithConfigMapFunc(fn func() (map[string]any, error)) Option {
	return func(x *Client) error {
		data, err := fn()
		if err != nil {
			return err
		}
		return WithConfigMap(data)(x)
	}
}

// WithConfigData is an option to configure Client Config
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

// WithConfigMap is an option to configure Client Config
func WithConfigMap(m map[string]any) Option {
	return func(x *Client) error {
		// if region found, set it in location
		if region, ok := m["region"]; ok {
			m["location"] = region
		}
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

// WithConfig is an option to configure Client Config via Config struct.
func WithConfig(c *Config) Option {
	return func(x *Client) error {
		x.config = c
		return nil
	}
}

// WithEnv is an option to configure Client Config via env variables
func WithEnv() Option {
	return WithEnvPrefix("")
}

// WithEnvPrefix is an option to configure Client Config via env variables with prefix
func WithEnvPrefix(prefix string) Option {
	return func(x *Client) error {
		if x.config == nil {
			x.config = &Config{}
		}
		// base config
		x.config.Project = util.GetEnvString(x.config.Project, "GOOGLE_PROJECT", "GOOGLE_CLOUD_PROJECT", "GCLOUD_PROJECT", "CLOUDSDK_CORE_PROJECT")
		x.config.Location = util.GetEnvString(x.config.Location, "GOOGLE_REGION", "GCLOUD_REGION", "CLOUDSDK_COMPUTE_REGION")
		x.config.Zone = util.GetEnvString(x.config.Zone, "GOOGLE_ZONE", "GCLOUD_ZONE", "CLOUDSDK_COMPUTE_ZONE")
		x.config.CredentialsFile = util.GetEnvString(x.config.CredentialsFile, "GOOGLE_APPLICATION_CREDENTIALS")
		return nil
	}
}

const scopeCloudPlatform = "https://www.googleapis.com/auth/cloud-platform"

// New creates new Client
func New(opts ...Option) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	scopes := []string{}
	if c.config.Scopes != "" {
		scopes = strings.Split(c.config.Scopes, ",")
	}
	if len(scopes) == 0 {
		scopes = []string{
			scopeCloudPlatform,
			//"https://www.googleapis.com/auth/userinfo.email",
		}
	}

	credsJSON := c.config.Credentials
	if c.config.CredentialsFile != "" {
		dcreds, err := os.ReadFile(c.config.CredentialsFile)
		if err != nil {
			return nil, err
		}
		credsJSON = string(dcreds)
	}
	var creds *google.Credentials
	var err error
	if credsJSON != "" {
		creds, err = google.CredentialsFromJSON(context.Background(), []byte(credsJSON), scopes...)
	} else {
		creds, err = google.FindDefaultCredentials(context.Background(), scopes...)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting credentials - %w", err)
	}
	c.credentials = creds

	return c, nil
}

// Client holds gcp client
type Client struct {
	config      *Config
	credentials *google.Credentials
}

// GetCredentials returns google credentials
func (x *Client) GetCredentials() *google.Credentials {
	return x.credentials
}

// GetTokenOption returns creds as ClientOption
func (x *Client) GetTokenOption(ctx context.Context) option.ClientOption {
	return option.WithHTTPClient(oauth2.NewClient(ctx, x.credentials.TokenSource))
}

// GetConfig returns config
func (x *Client) GetConfig() *Config {
	return x.config
}

// Config struct
type Config struct {
	Project         string            `json:"project" yaml:"project"`
	Location        string            `json:"location" yaml:"location"`
	Zone            string            `json:"zone" yaml:"zone"`
	Scopes          string            `json:"scopes" yaml:"scopes"`
	RequestReason   string            `json:"request_reason" yaml:"request_reason"`
	Endpoints       map[string]string `json:"endpoints" yaml:"endpoints"`
	CredentialsFile string            `json:"credentials_file" yaml:"credentials_file"`
	Credentials     string            `json:"credentials" yaml:"credentials"`
}

// IsNotFound returns true if err is IsNotFound
func IsNotFound(err error) bool {
	if status.Code(err) == codes.NotFound {
		return true
	}
	if e, ok := err.(*googleapi.Error); ok {
		// e.Errors[0].Reason // notFound
		return e.Code == 404
	}
	return false
}
