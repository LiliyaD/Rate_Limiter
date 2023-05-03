package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/LiliyaD/Rate_Limiter/pkg/constants"
	"github.com/spf13/viper"
)

var (
	configPath string
)

// func init() {
// 	flag.StringVar(&configPath, "config", "", "Search config path")
// 	flag.Parse()
// }

type Config struct {
	RateLimiterHost    string     `mapstructure:"rateLimiterHost"`
	SubnetPrefixLength int        `mapstructure:"subnetPrefixLength"`
	TimeCooldownSec    int        `mapstructure:"timeCooldownSec"`
	RateLimits         RateLimits `mapstructure:"rateLimits"`
}

type RateLimits struct {
	RequestsCount int `mapstructure:"requestsCount"`
	TimeSec       int `mapstructure:"timeSec"`
}

func InitConfig() (*Config, error) {
	flag.StringVar(&configPath, "config", "", "Search config path")
	flag.Parse()

	if configPath == "" {
		configPathFromEnv := os.Getenv(constants.ConfigPath)
		if configPathFromEnv != "" {
			configPath = configPathFromEnv
		}
	}

	cfg := &Config{}

	if configPath != "" {
		viper.SetConfigType(constants.Yaml)
		viper.SetConfigFile(configPath)

		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("viper.ReadInConfig %w", err)
		}

		if err := viper.Unmarshal(cfg); err != nil {
			return nil, fmt.Errorf("viper.Unnarshal %w", err)
		}
		return cfg, nil
	}

	rateLimiterHost := os.Getenv(constants.RateLimiterHost)
	subnetPrefixLength := os.Getenv(constants.SubnetPrefixLength)
	timeCooldownSec := os.Getenv(constants.TimeCooldownSec)
	requestsLimit := os.Getenv(constants.RequestsLimit)
	timeLimitSec := os.Getenv(constants.TimeLimitSec)

	if rateLimiterHost != "" && subnetPrefixLength != "" && timeCooldownSec != "" && requestsLimit != "" && timeLimitSec != "" {
		cfg.RateLimiterHost = rateLimiterHost

		var err error
		if cfg.SubnetPrefixLength, err = strconv.Atoi(subnetPrefixLength); err != nil {
			return nil, fmt.Errorf("convert string to int error %w", err)
		}

		if cfg.TimeCooldownSec, err = strconv.Atoi(timeCooldownSec); err != nil {
			return nil, fmt.Errorf("convert string to int error %w", err)
		}

		if cfg.RateLimits.RequestsCount, err = strconv.Atoi(requestsLimit); err != nil {
			return nil, fmt.Errorf("convert string to int error %w", err)
		}

		if cfg.RateLimits.TimeSec, err = strconv.Atoi(timeLimitSec); err != nil {
			return nil, fmt.Errorf("convert string to int error %w", err)
		}
	} else {
		return nil, fmt.Errorf("missing environment variable, check variables %s, %s, %s, %s, %s", constants.RateLimiterHost,
			constants.SubnetPrefixLength, constants.TimeCooldownSec, constants.RequestsLimit, constants.TimeLimitSec)
	}

	return cfg, nil
}
