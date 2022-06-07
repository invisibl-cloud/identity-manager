package azurex

import (
	"errors"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// GetEnvironment returns azure.Environment based on Config's Environment
func (c *Config) GetEnvironment() (azure.Environment, error) {
	if c.Environment == "" {
		return azure.PublicCloud, nil
	}
	return azure.EnvironmentFromName(c.Environment)
}

// GetAuthorizer returns the autorest.Authorizer based on the available settings
func (c *Config) GetAuthorizer() (autorest.Authorizer, error) {

	//1. Client Credentials
	if c, e := c.GetClientCredentials(); e == nil {
		return c.Authorizer()
	}

	//2. Client Certificate
	if c, e := c.GetClientCertificate(); e == nil {
		return c.Authorizer()
	}

	//3. Username Password
	if c, e := c.GetUsernamePassword(); e == nil {
		return c.Authorizer()
	}

	// 4. MSI
	return c.GetMSI().Authorizer()
}

// GetClientCredentials creates a config object from the available client credentials.
// An error is returned if no client credentials are available.
func (c *Config) GetClientCredentials() (*auth.ClientCredentialsConfig, error) {
	secret := c.ClientSecret
	if secret == "" {
		return nil, errors.New("missing client secret")
	}
	config := auth.NewClientCredentialsConfig(c.ClientID, secret, c.TenantID)
	//config.AADEndpoint = c.Environment().ActiveDirectoryEndpoint
	//config.Resource = settings.Values[Resource]
	/*
		if auxTenants, ok := settings.Values[AuxiliaryTenantIDs]; ok {
			config.AuxTenants = strings.Split(auxTenants, ";")
			for i := range config.AuxTenants {
				config.AuxTenants[i] = strings.TrimSpace(config.AuxTenants[i])
			}
		}*/
	return &config, nil
}

// GetClientCertificate creates a config object from the available certificate credentials.
// An error is returned if no certificate credentials are available.
func (c *Config) GetClientCertificate() (*auth.ClientCertificateConfig, error) {
	certPath := c.CertificatePath
	if certPath == "" {
		return nil, errors.New("missing certificate path")
	}
	certPwd := c.CertificatePassword
	config := auth.NewClientCertificateConfig(certPath, certPwd, c.ClientID, c.TenantID)
	//config.AADEndpoint = settings.Environment.ActiveDirectoryEndpoint
	//config.Resource = settings.Values[Resource]
	return &config, nil
}

// GetUsernamePassword creates a config object from the available username/password credentials.
// An error is returned if no username/password credentials are available.
func (c *Config) GetUsernamePassword() (*auth.UsernamePasswordConfig, error) {
	username := c.Username
	password := c.Password
	if username == "" || password == "" {
		return nil, errors.New("missing username/password")
	}
	config := auth.NewUsernamePasswordConfig(username, password, c.ClientID, c.TenantID)
	//config.AADEndpoint = settings.Environment.ActiveDirectoryEndpoint
	//config.Resource = settings.Values[Resource]
	return &config, nil
}

// GetMSI creates a MSI config object from the available client ID.
func (c *Config) GetMSI() *auth.MSIConfig {
	config := auth.NewMSIConfig()
	config.ClientID = c.ClientID
	return &config
}

// GetDeviceFlow creates a device-flow config object from the available client and tenant IDs.
func (c *Config) GetDeviceFlow() *auth.DeviceFlowConfig {
	config := auth.NewDeviceFlowConfig(c.ClientID, c.TenantID)
	//config.AADEndpoint = settings.Environment.ActiveDirectoryEndpoint
	//config.Resource = settings.Values[Resource]
	return &config
}
