package github

import "github.com/shurcooL/githubv4"

type ProjectQuery struct {
	Organization struct {
		ProjectV2 struct {
			Id string
		} `graphql:"projectV2(number: $projectID)"`
	} `graphql:"organization(login:$login)"`
}

type member struct {
	Login string
}

type TeamMembersQuery struct {
	Organization struct {
		Team struct {
			Name    string
			Members struct {
				Nodes []member
			} `graphql:"members(membership:ALL)"`
		} `graphql:"team(slug:$slug)"`
	} `graphql:"organization(login:$login)"`
}

type AddToProjectMutation struct {
	AddProjectV2ItemById struct {
		Item struct {
			Id githubv4.String
		}
	} `graphql:"addProjectV2ItemById(input: $input)"`
}
