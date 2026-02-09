package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strings"
)

type AppConfig struct {
	ServerPort   string `mapstructure:"SERVER_PORT"`
	AppEnv       string `mapstructure:"APP_ENV"`
	KafkaBrokers string `mapstructure:"KAFKA_BROKERS"` // String olarak alıp split edeceğiz
	KafkaTopic   string `mapstructure:"KAFKA_TOPIC"`
}

func LoadConfig() (*AppConfig, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		zap.L().Warn("Environment file not found, using system environment variables", zap.Error(err))
	}

	var appConfig AppConfig
	if err := viper.Unmarshal(&appConfig); err != nil {
		zap.L().Fatal("Unable to decode config", zap.Error(err))
		return nil, err
	}

	return &appConfig, nil
}

// GetBrokers - Virgülle ayrılmış string'i slice'a çeviren yardımcı metod.
// Örn: "host1:9092,host2:9092" -> []string{"host1:9092", "host2:9092"}
func (c *AppConfig) GetBrokers() []string {
	return strings.Split(c.KafkaBrokers, ",")
}
