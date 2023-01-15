package structs

type Config struct {
	Project     Project           `yaml:"project"`
	LabelOwners map[string]Owners `yaml:"labelOwners,omitempty"`
}

type Project struct {
	ID int `yaml:"id"`
}

type Owners struct {
	Logins []string `yaml:"logins"`
}
