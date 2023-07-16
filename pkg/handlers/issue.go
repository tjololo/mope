package handlers

import (
	"context"
	"github.com/google/go-github/v50/github"
	mopegithub "github.com/tjololo/mope/pkg/github"
	"github.com/tjololo/mope/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"regexp"
)

func HandleIssueOpenEvent(deliveryID string, eventName string, event *github.IssuesEvent) error {
	utils.Logger.Info("Issue commented", zap.Stringp("user", event.Sender.Login))
	client, err := mopegithub.NewClient(*event.Installation.ID)
	if err != nil {
		utils.Logger.Error("Failed to get github clients", zap.Error(err))
		return err
	}
	ctx := context.Background()
	config, err := client.ReadConfigFromRepo(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *event.Repo.DefaultBranch)
	if err != nil {
		utils.Logger.Error("Failed to parse config", zap.Error(err))
		return err
	}
	projectIDs, err := client.GetProjectIDs(ctx, *event.Repo.Owner.Login, config.Project.GetIDs())
	if err != nil {
		utils.Logger.Error("fail", zap.Error(err))
		return err
	}

	utils.Logger.Info("project found", zap.Strings("id", projectIDs))
	return client.AddItemToProjects(ctx, projectIDs, *event.Issue.NodeID)
}

func HandleIssueCommentCreated(deliveryID string, eventName string, event *github.IssueCommentEvent) error {
	utils.Logger.Info("Issue commented", zap.Stringp("user", event.Sender.Login))
	commentBody := *event.Comment.Body
	regx := regexp.MustCompile(`^/label (.*?)(\s.*?)?$`)
	if !regx.MatchString(commentBody) {
		return nil
	}
	labelString := regx.FindStringSubmatch(commentBody)[1]
	client, err := mopegithub.NewClient(*event.Installation.ID)
	if err != nil {
		utils.Logger.Error("Failed to get github clients", zap.Error(err))
		return err
	}

	ctx := context.Background()
	config, err := client.ReadConfigFromRepo(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *event.Repo.DefaultBranch)
	if err != nil {
		utils.Logger.Error("Failed to parse config", zap.Error(err))
		return err
	}
	exists := false
	for _, v := range event.Issue.Labels {
		if *v.Name == labelString {
			exists = true
			break
		}
	}
	val, found := config.LabelOwners[labelString]
	if found && !exists {
		authorized := slices.Contains(val.Logins, *event.Sender.Login)
		if !authorized {
			members, err := client.GetMembersOfTeams(ctx, *event.Repo.Owner.Login, val.Teams)
			if err != nil {
				utils.Logger.Error("failed to get members of teams", zap.Error(err), zap.Strings("teams", val.Teams))
				return err
			}
			authorized = slices.Contains(members, *event.Sender.Login)
		}
		if authorized {
			return client.AddLabelToItem(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number, labelString)
		}
	}
	return nil
}
