package configloader

import (
	"errors"
	"io"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Config holds the application configuration.
type Config struct {
	Controller      Controller `yaml:"controller"`
	LogDevMode      bool       `yaml:"logdevmode" envconfig:"NOPERATOR_DEV_MODE"`
	RequeueInterval int        `yaml:"requeueinterval" envconfig:"NOPERATOR_REQUEUE_INTERVAL"`
	CalicoASNRange  string     `yaml:"calicoasnrange" envconfig:"NOPERATOR_CALICO_ASN_RANGE"`
	L4lbTenant      string     `yaml:"l4lbtenant" envconfig:"NOPERATOR_L4LB_TENANT"`
	VPCID           int        `yaml:"vpcid" envconfig:"NOPERATOR_VPC_ID"`
}

// Controller holds the Netris controller connection configuration.
type Controller struct {
	Host     string `yaml:"host" envconfig:"CONTROLLER_HOST"`
	Login    string `yaml:"login" envconfig:"CONTROLLER_LOGIN"`
	Password string `yaml:"password" envconfig:"CONTROLLER_PASSWORD"`
	Insecure bool   `yaml:"insecure" envconfig:"CONTROLLER_INSECURE"`
}

// Load initializes the configuration from the given config file path and
// environment variables. Environment variables override file values.
// Returns the loaded configuration or an error.
func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	cfg, err := readConfig(f)
	if err != nil {
		return cfg, err
	}

	if len(cfg.Controller.Host) == 0 {
		return cfg, errors.New("netris controller credentials not set")
	}

	return cfg, nil
}

// readConfig reads the configuration from the given reader.
func readConfig(r io.Reader) (Config, error) {
	cfg := Config{}
	if err := yaml.NewDecoder(r).Decode(&cfg); err != nil {
		return cfg, err
	}
	return cfg, envconfig.Process("", &cfg)
}
