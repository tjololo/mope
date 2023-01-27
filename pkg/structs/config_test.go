package structs

import (
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
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
	if config.Project.ID == nil {
		t.Error("expected project.id to be not null")
	}
	if config.Project.IDs != nil {
		t.Error("expected project.ids to be null")
	}
	expected := [1]int{1}
	actual := config.Project.GetIDs()
	if reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected list of id returned from GetIDs. Expected %v got %v", actual, expected)
	}
}

func TestConfigReadingIDList(t *testing.T) {
	var config *Config
	b, err := os.ReadFile("testdata/config-ids.yaml")
	if err != nil {
		t.Error(err)
	}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		t.Error(err)
	}
	if config.Project.ID != nil {
		t.Error("expected project.id to be null")
	}
	if config.Project.IDs == nil {
		t.Error("expected project.ids to be not null")
	}
	expected := [2]int{1, 2}
	if reflect.DeepEqual(config.Project.GetIDs(), expected) {
		t.Errorf("unexpected list of id returned from GetIDs. Expected %v got %v", config.Project.GetIDs(), expected)
	}
}

func TestConfigReadingBothIDs(t *testing.T) {
	var config *Config
	b, err := os.ReadFile("testdata/config-both-id.yaml")
	if err != nil {
		t.Error(err)
	}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		t.Error(err)
	}
	if config.Project.ID == nil {
		t.Error("expected project.id to be not null")
	}
	if config.Project.IDs == nil {
		t.Error("expected project.ids to be not null")
	}
	expected := [3]int{2, 3, 1}
	if reflect.DeepEqual(config.Project.GetIDs(), expected) {
		t.Errorf("unexpected list of id returned from GetIDs. Expected %v got %v", config.Project.GetIDs(), expected)
	}
}
