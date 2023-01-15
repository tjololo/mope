package structs

import (
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func TestConfigReading(t *testing.T) {
	var config *Config
	b, err := os.ReadFile("testdata/config.yaml")
	if err != nil {
		t.Error(err)
	}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		t.Error(err)
	}
}
