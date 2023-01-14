package structs

type Config struct {
	Project Project `yaml:"project"`
}

type Project struct {
	ID int `yaml:"id"`
}
