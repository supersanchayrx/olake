package driver

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/datazip-inc/olake/constants"
	"github.com/datazip-inc/olake/utils"
)

// PrimaryConfig struct is used to define fields to connect to primary MSSQL database
type PrimaryConfig struct {
	Host             string            `json:"host"`
	Port             int               `json:"port"`
	Username         string            `json:"username"`
	Password         string            `json:"password"`
	JDBCURLParams    map[string]string `json:"jdbc_url_params"`
	SSLConfiguration *utils.SSLConfig  `json:"ssl"`
	SSHConfig        *utils.SSHConfig  `json:"ssh_config"`
}

// Config represents the configuration for connecting to a MSSQL database.
type Config struct {
	Host                   string            `json:"host"`
	Port                   int               `json:"port"`
	Database               string            `json:"database"`
	Username               string            `json:"username"`
	Password               string            `json:"password"`
	PrimaryConfig          *PrimaryConfig    `json:"primary_config"`
	MaxThreads             int               `json:"max_threads"`
	RetryCount             int               `json:"retry_count"`
	JDBCURLParams          map[string]string `json:"jdbc_url_params"`
	SSLConfiguration       *utils.SSLConfig  `json:"ssl"`
	ManageCaptureInstances bool              `json:"manage_capture_instances"`
	SSHConfig              *utils.SSHConfig  `json:"ssh_config"`
}

// Validate checks and normalises MSSQL configuration.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("empty host name")
	} else if strings.Contains(c.Host, "https") || strings.Contains(c.Host, "http") {
		return fmt.Errorf("host should not contain http or https")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port number: must be between 1 and 65535")
	}

	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.Password == "" {
		return fmt.Errorf("password is required")
	}
	if c.Database == "" {
		return fmt.Errorf("database is required")
	}

	if c.MaxThreads <= 0 {
		c.MaxThreads = constants.DefaultThreadCount
	}

	if c.RetryCount <= 0 {
		c.RetryCount = constants.DefaultRetryCount
	}

	if c.SSLConfiguration == nil {
		c.SSLConfiguration = &utils.SSLConfig{
			Mode: utils.SSLModeDisable,
		}
	}

	if c.PrimaryConfig != nil {
		if c.PrimaryConfig.Host == "" {
			return fmt.Errorf("empty Primary host name")
		} else if strings.Contains(c.PrimaryConfig.Host, "https") || strings.Contains(c.PrimaryConfig.Host, "http") {
			return fmt.Errorf("Primary host should not contain http or https")
		}

		if c.PrimaryConfig.Port <= 0 || c.PrimaryConfig.Port > 65535 {
			return fmt.Errorf("invalid Primary's port number: must be between 1 and 65535")
		}

		if c.PrimaryConfig.Username == "" {
			return fmt.Errorf("Primary's username is required")
		}
		if c.PrimaryConfig.Password == "" {
			return fmt.Errorf("Primary's password is required")
		}
	}

	err := c.SSLConfiguration.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate ssl config: %s", err)
	}

	return utils.Validate(c)
}

// a helper method that allows a centralized uri building for both replicas & primary db
func buildURI(host string, port int, username, password, database string, params map[string]string, ssl *utils.SSLConfig) string {
	if !strings.Contains(host, ":") {
		host = fmt.Sprintf("%s:%d", host, port)
	}

	query := url.Values{}

	for k, v := range params {
		query.Add(k, v)
	}

	query.Set("database", database)

	if ssl == nil {
		query.Set("encrypt", "disable")
	} else {
		switch string(ssl.Mode) {
		case utils.SSLModeDisable:
			query.Set("encrypt", "disable")
		case utils.SSLModeRequire:
			query.Set("encrypt", "true")
			query.Set("TrustServerCertificate", "true")
		default:
			query.Set("encrypt", "disable")
		}
	}

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(username, password),
		Host:     host,
		RawQuery: query.Encode(),
	}

	return u.String()
}

// URI returns the sqlserver:// connection string for go-mssqldb.
func (c *Config) URI() string {
	return buildURI(c.Host,
		c.Port,
		c.Username,
		c.Password,
		c.Database,
		c.JDBCURLParams,
		c.SSLConfiguration)
}

// Does the same job as above one but specifically for Primary connections
func (c *Config) PrimaryURI() string {
	return buildURI(c.PrimaryConfig.Host,
		c.PrimaryConfig.Port,
		c.PrimaryConfig.Username,
		c.PrimaryConfig.Password,
		c.Database,
		c.PrimaryConfig.JDBCURLParams,
		c.PrimaryConfig.SSLConfiguration)
}
