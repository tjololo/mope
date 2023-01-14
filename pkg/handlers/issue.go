package handlers

import (
	"context"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v49/github"
	"github.com/shurcooL/githubv4"
	"github.com/tjololo/mope/pkg/structs"
	"github.com/tjololo/mope/pkg/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"strconv"
)

func HandleIssueCommentCreated(deliveryID string, eventName string, event *github.IssueCommentEvent) error {
	utils.Logger.Info("Issue commented", zap.Stringp("user", event.Sender.Login))
	appId, err := strconv.Atoi(os.Getenv("GITHUB_APP_ID"))
	if err != nil {
		utils.Logger.Error("coudl not parse $GITHUB_APP_ID envvar to int", zap.Error(err))
	}
	privateKeyFile := os.Getenv("GITHUB_PRIVATE_KEY_FILE")
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, int64(appId), *event.Installation.ID, privateKeyFile)

	if err != nil {
		return err
	}

	// Use installation transport with client.
	httpClient := &http.Client{Transport: itr}
	client := github.NewClient(httpClient)
	client2 := githubv4.NewClient(httpClient)

	ctx := context.Background()

	f, _, _, err := client.Repositories.GetContents(ctx, *event.Repo.Owner.Login, *event.Repo.Name, ".github/mope.yaml", nil)
	var config structs.Config
	s, err := f.GetContent()
	if err != nil {
		utils.Logger.Error("failed to get content", zap.Error(err))
	}
	err = yaml.Unmarshal([]byte(s), &config)
	if err != nil {
		utils.Logger.Error("failed to read config", zap.Error(err), zap.Stringp("content", f.Content))
	}

	var query struct {
		Organization struct {
			ProjectV2 struct {
				Id string
			} `graphql:"projectV2(number: $projectID)"`
		} `graphql:"organization(login:$login)"`
	}
	vars := map[string]interface{}{
		"projectID": githubv4.Int(config.Project.ID),
		"login":     githubv4.String(*event.Repo.Owner.Login),
	}
	err = client2.Query(ctx, &query, vars)
	if err != nil {
		utils.Logger.Error("fail", zap.Error(err))
		return err
	}
	utils.Logger.Info("project found", zap.String("id", query.Organization.ProjectV2.Id))
	var mutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				Id githubv4.String
			}
		} `graphql:"addProjectV2ItemById(input: $input)"`
	}
	input := githubv4.AddProjectV2ItemByIdInput{
		ProjectID: query.Organization.ProjectV2.Id,
		ContentID: *event.Issue.NodeID,
	}
	err = client2.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		utils.Logger.Error("update failed", zap.Error(err))
		return err
	}

	_, _, err = client.Issues.AddLabelsToIssue(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number, []string{"test"})

	if err != nil {
		return err
	}
	return nil
}
