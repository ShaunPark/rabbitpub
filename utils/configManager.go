package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ShaunPark/rabbitPub/types"

	klog "k8s.io/klog/v2"

	"gopkg.in/yaml.v2"
)

type ConfigManager struct {
	config       *types.Config
	lastReadTime time.Time
	configFile   string
}

func NewConfigManager(configFile string) *ConfigManager {
	config := readConfig(configFile)
	return &ConfigManager{config: config, lastReadTime: time.Now(), configFile: configFile}
}

func (c *ConfigManager) GetConfig() *types.Config {
	now := time.Now()
	if now.After(c.lastReadTime.Add(time.Minute * 1)) {
		file, err := os.Stat(c.configFile)
		if err != nil {
			klog.Error(err)
		}

		modifiedtime := file.ModTime()
		if modifiedtime.After(c.lastReadTime) {
			klog.V(3).Infof("Reload config file")
			c.config = readConfig(c.configFile)
		} else {
			klog.V(4).Infof("Skip reload config file. File hasn't been modified")
		}
		c.lastReadTime = now
	} else {
		klog.V(4).Infof("Skip reload config file")
	}

	return c.config
}

func readConfig(configFile string) *types.Config {
	fileName, _ := filepath.Abs(configFile)
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	var config types.Config

	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		panic(err)
	}
	return &config
}
