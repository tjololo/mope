package github

import (
	"context"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v60/github"
	"github.com/shurcooL/githubv4"
	"github.com/tjololo/mope/pkg/structs"
	"github.com/tjololo/mope/pkg/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"strconv"
)

type GithubClient struct {
	client   *github.Client
	clientV4 *githubv4.Client
}

func NewClient(installationId int64) (*GithubClient, error) {
	appId, err := strconv.Atoi(os.Getenv("GITHUB_APP_ID"))
	if err != nil {
		utils.Logger.Error("coudl not parse $GITHUB_APP_ID envvar to int", zap.Error(err))
		return nil, err
	}
	privateKeyFile := os.Getenv("GITHUB_PRIVATE_KEY_FILE")
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, int64(appId), installationId, privateKeyFile)

	if err != nil {
		return nil, err
	}

	// Use installation transport with client.
	httpClient := &http.Client{Transport: itr}
	client := github.NewClient(httpClient)
	clientV4 := githubv4.NewClient(httpClient)
	return &GithubClient{
		client:   client,
		clientV4: clientV4,
	}, nil
}

func (g *GithubClient) ReadConfigFromRepo(ctx context.Context, owner, name, branch string) (*structs.Config, error) {
	f, _, _, err := g.client.Repositories.GetContents(ctx, owner, name, ".github/mope.yaml", &github.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err != nil {
		return nil, err
	}
	var config *structs.Config
	s, err := f.GetContent()
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(s), &config)
	return config, err
}

func (g *GithubClient) GetProjectIDs(ctx context.Context, organization string, projectNumbers []int) (projectIDs []string, err error) {
	var s string
	for _, projectNumber := range projectNumbers {
		s, err = g.GetProjectID(ctx, organization, projectNumber)
		if err != nil {
			return nil, err
		}
		projectIDs = append(projectIDs, s)
	}
	return
}

func (g *GithubClient) GetProjectID(ctx context.Context, organization string, projectNumber int) (string, error) {
	var query ProjectQuery
	vars := map[string]interface{}{
		"projectID": githubv4.Int(projectNumber),
		"login":     githubv4.String(organization),
	}
	err := g.clientV4.Query(ctx, &query, vars)
	if err != nil {
		return "", err
	}
	return query.Organization.ProjectV2.Id, nil
}

func (g *GithubClient) GetMembersOfTeam(ctx context.Context, organization, teamslug string) ([]string, error) {
	var query TeamMembersQuery
	vars := map[string]interface{}{
		"login": githubv4.String(organization),
		"slug":  githubv4.String(teamslug),
	}
	err := g.clientV4.Query(ctx, &query, vars)
	if err != nil {
		return nil, err
	}
	var members []string
	for _, node := range query.Organization.Team.Members.Nodes {
		members = append(members, node.Login)
	}
	return members, nil
}

func (g *GithubClient) GetMembersOfTeams(ctx context.Context, organization string, teams []string) ([]string, error) {
	var members []string
	if teams == nil {
		return members, nil
	}
	for _, team := range teams {
		m, err := g.GetMembersOfTeam(ctx, organization, team)
		if err != nil {
			return nil, err
		}
		members = append(members, m...)
	}
	return members, nil
}

func (g *GithubClient) AddItemToProjects(ctx context.Context, projectIds []string, itemNodeID string) error {
	for _, projectId := range projectIds {
		if err := g.AddItemToProject(ctx, projectId, itemNodeID); err != nil {
			return err
		}
	}
	return nil
}

func (g *GithubClient) AddItemToProject(ctx context.Context, projectId, itemNodeID string) error {
	var mutation AddToProjectMutation
	input := githubv4.AddProjectV2ItemByIdInput{
		ProjectID: projectId,
		ContentID: itemNodeID,
	}
	return g.clientV4.Mutate(ctx, &mutation, input, nil)
}

func (g *GithubClient) AddLabelToItem(ctx context.Context, organization, repository string, itemnumber int, label ...string) error {
	_, _, err := g.client.Issues.AddLabelsToIssue(ctx, organization, repository, itemnumber, label)
	return err
}
