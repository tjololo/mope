package structs

type Config struct {
	Project          Project            `yaml:"project"`
	LabelOwners      map[string]*Owners `yaml:"labelOwners,omitempty"`
	ForkPullRequests *ForkPullRequests  `yaml:"forkPullRequests,omitempty"`
}

type Project struct {
	ID  *int  `yaml:"id,omitempty"`
	IDs []int `yaml:"ids,omitempty"`
}

func (p *Project) GetIDs() []int {
	var allIDs []int
	if p.IDs != nil {
		allIDs = p.IDs
	}
	if p.ID != nil {
		allIDs = append(allIDs, *p.ID)
	}
	return allIDs
}

type Owners struct {
	Logins []string `yaml:"logins"`
	Teams  []string `yaml:"teams"`
}

type ForkPullRequests struct {
	Labels       []string `yaml:"labels"`
	AddToProject bool     `yaml:"addToProject"`
}
