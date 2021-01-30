package config

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type grpcServerConfig struct {
	Host     string
	Port     int64
	CertPath string
}

type grpcClientConfig struct {
	Host string
	Port int64
}

// Config struct
type Config struct {
	Server grpcServerConfig
	Client grpcClientConfig
}

// Init function load config.yml file and initializes global configuration variables
func Init(path string) Config {
	logrus.Println("[*] Initializing config...")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if path != "" {
		viper.AddConfigPath(path)
	}

	cwd, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}

	viper.AddConfigPath(cwd)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Fatal("[-] Error: No config file found")
		} else {
			logrus.Fatal("[-] Error: Config file malformed")
		}
	} else {
		logrus.Printf("[+] Using config file: %s\n", viper.ConfigFileUsed())
	}

	//grpcServer
	setConfigItem("server.host", "localhost")
	setConfigItem("server.port", 9999)
	setConfigItem("server.certPath", cwd+"/server/cert")

	//grpcClient
	setConfigItem("server.host", "localhost")
	setConfigItem("server.port", 9999)

	return toStruct()

}

func setConfigItem(key string, value interface{}) string {
	viper.SetDefault(key, value)
	return key
}

func toStruct() Config {
	c := Config{
		Server: grpcServerConfig{
			Host:     viper.GetString("server.host"),
			Port:     viper.GetInt64("server.port"),
			CertPath: viper.GetString("server.certPath"),
		},
		Client: grpcClientConfig{
			Host: viper.GetString("server.host"),
			Port: viper.GetInt64("server.port"),
		},
	}

	return c
}
