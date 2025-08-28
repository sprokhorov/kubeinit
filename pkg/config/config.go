package config

import (
	"errors"
	"os"
	"regexp"
)

type Config struct {
	HelmfileFile  string
	CloudProvider CloudProvider
	ClusterName   string
}

type CloudProvider string

const (
	AWS   CloudProvider = "aws"
	Azure CloudProvider = "azure"
	GCP   CloudProvider = "gcp"
)

func New() (*Config, error) {
	cfg := &Config{}
	if err := cfg.lookup(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) lookup() error {
	if err := c.fetchHelmfile(); err != nil {
		return err
	}
	if err := c.fetchCloudProvider(); err != nil {
		return err
	}
	if err := c.fetchClusterName(); err != nil {
		return err
	}
	return nil
}

func (c *Config) fetchHelmfile() error {
	helmfileFile, exists := os.LookupEnv("HELMFILE_FILE")
	if !exists {
		return errors.New("HELMFILE_FILE environment variable is not set")
	}
	matched, err := regexp.MatchString(`^(file|https?|s3|git)://`, helmfileFile)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("HELMFILE_FILE must start with file://, http://, https://, s3://, or git://")
	}
	c.HelmfileFile = helmfileFile
	return nil
}

func (c *Config) fetchCloudProvider() error {
	cloudProvider, exists := os.LookupEnv("CLOUD_PROVIDER")
	if !exists {
		return errors.New("CLOUD_PROVIDER environment variable is not set")
	}
	switch CloudProvider(cloudProvider) {
	case AWS, Azure, GCP:
		c.CloudProvider = CloudProvider(cloudProvider)
	default:
		return errors.New("CLOUD_PROVIDER must be one of: aws, azure, gcp")
	}
	return nil
}

func (c *Config) fetchClusterName() error {
	clusterName, exists := os.LookupEnv("CLUSTER_NAME")
	if !exists {
		return errors.New("CLUSTER_NAME environment variable is not set")
	}
	c.ClusterName = clusterName
	return nil
}
