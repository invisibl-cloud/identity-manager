package awsx

import (
	"flag"

	"github.com/invisibl-cloud/identity-manager/pkg/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	// RegionName - aws region name
	RegionName = "region"
	// AccessKeyIDName - aws access key id name
	AccessKeyIDName = "aws_access_key_id"
	// SecretAccessKeyName - aws secret access key name
	// #nosec
	SecretAccessKeyName = "aws_secret_access_key"
	// RoleArnName - role arn name
	RoleArnName = "role_arn"
	// ExternalIDName - external id name
	ExternalIDName = "external_id"
)

// NewConfig expects the map of config data and returns
// the Config object
func NewConfig(m map[string][]byte) Config {
	cfg := Config{}
	if val, ok := m[RegionName]; ok {
		cfg.Region = string(val)
	}
	if val, ok := m[AccessKeyIDName]; ok {
		cfg.AccessKeyID = string(val)
	}
	if val, ok := m[SecretAccessKeyName]; ok {
		cfg.SecretAccessKey = string(val)
	}
	if val, ok := m[RoleArnName]; ok {
		cfg.RoleArn = string(val)
	}
	if val, ok := m[ExternalIDName]; ok {
		cfg.ExternalID = string(val)
	}
	return cfg
}

// NewSession expects Config and returns the *session.Session object
func NewSession(conf Config) (*session.Session, error) {
	// convert to aws config
	cfg := aws.NewConfig()
	cfgs := []*aws.Config{cfg}
	if conf.Region != "" {
		cfg.Region = aws.String(conf.Region)
	} else {
		cfg.Region = aws.String(util.GetEnvString(conf.Region, "AWS_REGION", "AWS_DEFAULT_REGION"))
	}
	// assume role.
	if conf.RoleArn != "" {
		cfgs1 := []*aws.Config{cfg}
		if conf.AccessKeyID != "" && conf.SecretAccessKey != "" {
			credsCfg := aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(conf.AccessKeyID, conf.SecretAccessKey, conf.SessionToken))
			cfgs1 = append(cfgs1, credsCfg)
		}
		sess1, err := session.NewSession(cfgs1...)
		if err != nil {
			return nil, err
		}
		creds := stscreds.NewCredentials(sess1, conf.RoleArn, func(arp *stscreds.AssumeRoleProvider) {
			arp.RoleSessionName = conf.Name
			//arp.Duration = 60 * time.Minute
			//arp.ExpiryWindow = 30 * time.Second
		})
		cfgs = append(cfgs, aws.NewConfig().WithCredentials(creds))
	} else {
		// static creds if any
		if conf.AccessKeyID != "" && conf.SecretAccessKey != "" {
			cfg.Credentials = credentials.NewStaticCredentials(conf.AccessKeyID, conf.SecretAccessKey, conf.SessionToken)
		}
	}
	sess, err := session.NewSession(cfgs...)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// Config - simple aws session config
type Config struct {
	Name            string `json:"-" ini:"-"`
	Region          string `json:"region" ini:"-"`
	AccessKeyID     string `json:"aws_access_key_id" ini:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key" ini:"aws_secret_access_key"`
	SessionToken    string `json:"aws_session_token" ini:"aws_session_token"`
	RoleArn         string `json:"role_arn" ini:"-"`
	ExternalID      string `json:"external_id" ini:"-"`
}

// CheckError - check aws error code.
func CheckError(err error, codes ...string) (bool, bool) {
	if aerr, ok := err.(awserr.Error); ok {
		for _, code := range codes {
			if aerr.Code() == code {
				return true, true
			}
		}
		return true, false
	}
	return false, false
}

// Options of AWS
type Options struct {
	PermissionsBoundaryARN string
}

// BindFlags will parse the given flagset for aws arg flags.
func (o *Options) BindFlags(fs *flag.FlagSet) {
	flag.StringVar(&o.PermissionsBoundaryARN, "aws-permissions-boundary-arn", "", "The permissions boundary arn.")
}
