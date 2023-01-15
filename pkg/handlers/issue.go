package handlers

import (
	"context"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v49/github"
	"github.com/shurcooL/githubv4"
	"github.com/tjololo/mope/pkg/structs"
	"github.com/tjololo/mope/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

func HandleIssueOpenEvent(deliveryID string, eventName string, event *github.IssuesEvent) error {
	utils.Logger.Info("Issue commented", zap.Stringp("user", event.Sender.Login))
	client, client2, err := getGithubClients(*event.Installation.ID)
	if err != nil {
		utils.Logger.Error("Failed to get github clients", zap.Error(err))
		return err
	}
	ctx := context.Background()
	config, err := readConfigFromRepo(ctx, client, *event.Repo.Owner.Login, *event.Repo.Name, *event.Repo.DefaultBranch)
	if err != nil {
		utils.Logger.Error("Failed to parse config", zap.Error(err))
		return err
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
	return nil
}

func HandleIssueCommentCreated(deliveryID string, eventName string, event *github.IssueCommentEvent) error {
	utils.Logger.Info("Issue commented", zap.Stringp("user", event.Sender.Login))
	commentBody := *event.Comment.Body
	regx := regexp.MustCompile(`^/label (.*?)(\s.*?)?$`)
	if !regx.MatchString(commentBody) {
		return nil
	}
	labelString := regx.FindStringSubmatch(commentBody)[1]
	client, _, err := getGithubClients(*event.Installation.ID)
	if err != nil {
		utils.Logger.Error("Failed to get github clients", zap.Error(err))
		return err
	}

	ctx := context.Background()

	config, err := readConfigFromRepo(ctx, client, *event.Repo.Owner.Login, *event.Repo.Name, *event.Repo.DefaultBranch)
	if err != nil {
		utils.Logger.Error("Failed to parse config", zap.Error(err))
		return err
	}

	val, found := config.LabelOwners[labelString]
	if found {
		slices.Contains(val.Logins, *event.Sender.Login)
		_, _, err = client.Issues.AddLabelsToIssue(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number, []string{labelString})
		if err != nil {
			return err
		}
	}
	return nil
}

func getGithubClients(installationId int64) (*github.Client, *githubv4.Client, error) {
	appId, err := strconv.Atoi(os.Getenv("GITHUB_APP_ID"))
	if err != nil {
		utils.Logger.Error("coudl not parse $GITHUB_APP_ID envvar to int", zap.Error(err))
	}
	privateKeyFile := os.Getenv("GITHUB_PRIVATE_KEY_FILE")
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, int64(appId), installationId, privateKeyFile)

	if err != nil {
		return nil, nil, err
	}

	// Use installation transport with client.
	httpClient := &http.Client{Transport: itr}
	client := github.NewClient(httpClient)
	clientv4 := githubv4.NewClient(httpClient)
	return client, clientv4, nil
}

func readConfigFromRepo(ctx context.Context, client *github.Client, owner, name, branch string) (*structs.Config, error) {
	f, _, _, err := client.Repositories.GetContents(ctx, owner, name, ".github/mope.yaml", &github.RepositoryContentGetOptions{
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
