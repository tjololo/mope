package handlers

import (
	"context"
	"github.com/google/go-github/v60/github"
	mopegithub "github.com/tjololo/mope/pkg/github"
	"github.com/tjololo/mope/pkg/utils"
	"go.uber.org/zap"
)

func HandlePullReuqestOpened(_ string, _ string, event *github.PullRequestEvent) error {
	client, err := mopegithub.NewClient(*event.Installation.ID)
	if err != nil {
		utils.Logger.Error("Failed to get github clients", zap.Error(err))
		return err
	}
	ctx := context.Background()
	if *event.PullRequest.Draft || !*event.GetPullRequest().Head.Repo.Fork {
		return nil
	}
	config, err := client.ReadConfigFromRepo(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *event.Repo.DefaultBranch)
	if err != nil {
		utils.Logger.Error("Failed to parse config", zap.Error(err))
		return err
	}
	if config.ForkPullRequests == nil {
		return nil
	}
	if len(config.ForkPullRequests.Labels) > 0 {
		err = client.AddLabelToItem(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *event.PullRequest.Number, config.ForkPullRequests.Labels...)
		if err != nil {
			utils.Logger.Error("Failed to label pullrequest", zap.Error(err))
			return err
		}
	}
	if config.ForkPullRequests.AddToProject {
		projectIDs, err := client.GetProjectIDs(ctx, *event.Repo.Owner.Login, config.Project.GetIDs())
		if err != nil {
			utils.Logger.Error("Failed to fetch projectIDs", zap.Error(err))
			return err
		}
		return client.AddItemToProjects(ctx, projectIDs, *event.PullRequest.NodeID)
	}
	return nil
}
