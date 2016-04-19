package config

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type Server struct {
	Address string `yaml:"address,omitempty"`
	Port    uint   `yaml:"port,omitempty"`
}

type Db struct {
	DbInfo     mgo.DialInfo `yaml:"dial_info,omitempty"`
	Collection string       `yaml:"collection,omitempty"`
}

type Config struct {
	Server         Server `yaml:"server,omitempty"`
	SearchUser     Db     `yaml:"search_user,omitempty"`
	SearchRepo     Db     `yaml:"search_repo,omitempty"`
	AnalysisConfig Db     `yaml:"analysis_config,omitempty"`
	MongoAuth      Db     `yaml:"mongo_auth,omitempty"`
}

// GetEndpointConnectionString builds and returns a string with the IP and port
// separated by a colon. Nothing special but anyway.
func (c Config) GetEndpointConnectionString() string {
	return fmt.Sprintf("%s:%d", c.Server.Address, c.Server.Port)
}

//validate the server configuration!
func validate_server(c *Server) error {
	if c.Address == "" {
		return fmt.Errorf("Server address must not be empty")
	}
	return nil
}

//validate the MongoDB configuration!
func validate_db(c *Db) error {
	if len(c.DbInfo.Addrs) == 0 {
		return fmt.Errorf("db.addrs must not be empty")
	}
	if c.DbInfo.Timeout == 0 {
		c.DbInfo.Timeout = 10 * time.Second
	}
	if c.DbInfo.Database == "" {
		return fmt.Errorf("db.database must not be empty")
	}
	if c.DbInfo.Username == "" {
		return fmt.Errorf("db.Username must not be empty")
	}
	if c.DbInfo.Password == "" {
		return fmt.Errorf("db.password_file must not be empty")
	}
	if c.Collection == "" {
		return fmt.Errorf("db.collection is required")
	}
	return nil
}

//validate the configuration
func validate(c *Config) error {
	if err := validate_server(&c.Server); err != nil {
		return err
	}
	if err := validate_db(&c.SearchUser); err != nil {
		return err
	}
	if err := validate_db(&c.SearchRepo); err != nil {
		return err
	}
	if err := validate_db(&c.AnalysisConfig); err != nil {
		return err
	}
	if err := validate_db(&c.MongoAuth); err != nil {
		return err
	}
	return nil
}

// LoadConfig parses all flags from the command line and returns
// an initialized Settings object and an error object if any. For instance if it
// cannot find the SSL certificate file or the SSL key file it will set the
// returned error appropriately.
func LoadConfig(path string) (*Config, error) {
	//fmt.Println("Starting Loading configuration!")
	c := &Config{}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config %s: %s", path, err)
	}
	if err = yaml.Unmarshal(contents, c); err != nil {
		return nil, fmt.Errorf("Failed to parse config: %s", err)
	}
	if err = validate(c); err != nil {
		return nil, fmt.Errorf("Invalid config: %s", err)
	}
	//fmt.Println("Loading configuration done!")
	return c, nil
}
