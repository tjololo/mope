package structs

type Config struct {
	Project          Project            `yaml:"project"`
	LabelOwners      map[string]*Owners `yaml:"labelOwners,omitempty"`
	ForkPullRequests *ForkPullRequests  `yaml:"forkPullRequests,omitempty"`
}

type Project struct {
	ID int `yaml:"id"`
}

type Owners struct {
	Logins []string `yaml:"logins"`
}

type ForkPullRequests struct {
	Labels       []string `yaml:"labels"`
	AddToProject bool     `yaml:"addToProject"`
}
