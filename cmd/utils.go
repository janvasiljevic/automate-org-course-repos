package cmd

import (
	"context"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/go-github/v49/github"
	"golang.org/x/oauth2"
)

func contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func createClient() (*github.Client, context.Context) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: Config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client, ctx
}

func getStudents(client *github.Client, ctx context.Context) ([]string, error) {
	members, _, err := client.Organizations.ListMembers(ctx, Config.OrgName, nil)

	if err != nil {
		return nil, err
	}

	var students []string

	for _, member := range members {
		if member.Login == nil {
			continue
		}

		if !contains(Config.WhiteListedMembers, *member.Login) {
			students = append(students, *member.Login)
		}
	}

	return students, nil
}

func getStudentRepositoriesNames(client *github.Client, ctx context.Context) ([]string, error) {
	students, err := getStudents(client, ctx)

	if err != nil {
		return nil, err
	}

	orgRepos, _, err := client.Repositories.ListByOrg(ctx, Config.OrgName, nil)

	if err != nil {
		return nil, err
	}

	var studentRepos []string

	for _, repo := range orgRepos {
		if repo.Name == nil {
			continue
		}

		if contains(students, *repo.Name) {
			studentRepos = append(studentRepos, *repo.Name)
		}
	}

	return studentRepos, nil
}

func yesNoPrompt(msg string) bool {
	sure := false

	yesPrompt := &survey.Confirm{
		Message: msg,
	}

	err := survey.AskOne(yesPrompt, &sure)

	if err != nil {
		return false
	}

	if !sure {
		return false
	}

	return sure
}
