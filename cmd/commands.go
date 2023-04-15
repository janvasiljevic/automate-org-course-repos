package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/go-github/v49/github"
	"github.com/urfave/cli/v2"
)

func ExecuteCommand(cliCtx *cli.Context, cmdFn func(*github.Client, context.Context) error) error {
	client, ctx := createClient()

	return cmdFn(client, ctx)
}

func ListStudents(client *github.Client, ctx context.Context) error {
	students, err := getStudents(client, ctx)

	if err != nil {
		return err

	}

	for _, student := range students {
		log.Printf("Student: %v \n", student)
	}

	return nil
}

func CreateRepos(client *github.Client, ctx context.Context) error {

	students, err := getStudents(client, ctx)

	if err != nil {
		return err
	}

	selectedStudents := []string{}

	multiPrompt := &survey.MultiSelect{
		Message: "For which students do you want to create repositories?",
		Options: students,
	}

	err = survey.AskOne(multiPrompt, &selectedStudents)

	if err != nil {
		return err
	}

	if !yesNoPrompt(fmt.Sprintf("Are you sure you want to create repositories for %v?", selectedStudents)) {
		return fmt.Errorf("Aborting...")
	}

	for _, student := range selectedStudents {

		fileBytes, err := os.ReadFile(Config.InitReadme)

		if err != nil {
			return err
		}

		_, _, err = client.Repositories.Create(ctx, Config.OrgName, &github.Repository{
			Name: github.String(student), Private: github.Bool(true),
		})

		if err != nil {
			fmt.Printf("Error creating repository for %v: %v\n", student, err)
		} else {
			fmt.Printf("Successfully created repository for %v\n", student)
		}

		_, _, err = client.Repositories.CreateFile(ctx, Config.OrgName, student, "README.md", &github.RepositoryContentFileOptions{
			Message: github.String("Initial commit"),
			Content: fileBytes,
		})

		if err != nil {
			fmt.Printf("Error creating README for %v: %v\n", student, err)
		} else {
			fmt.Printf("Successfully created README for %v\n", student)
		}
	}

	return nil
}

func DeleteStudentRepositories(client *github.Client, ctx context.Context) error {

	studentReposNames, err := getStudentRepositoriesNames(client, ctx)

	if err != nil {
		return err
	}

	if len(studentReposNames) == 0 {
		return fmt.Errorf("No student repositories found")
	}

	selectedRepos := []string{}

	multiPrompt := &survey.MultiSelect{
		Message: "Which repositories do you want to delete?",
		Options: studentReposNames,
	}

	err = survey.AskOne(multiPrompt, &selectedRepos)

	if err != nil {
		return err
	}

	if !yesNoPrompt(fmt.Sprintf("Are you sure you want to delete %v?", selectedRepos)) {
		return fmt.Errorf("Aborting...")
	}

	for _, repo := range selectedRepos {
		_, err := client.Repositories.Delete(ctx, Config.OrgName, repo)

		if err != nil {
			fmt.Printf("Error deleting repository for %v: %v\n", repo, err)
		} else {
			fmt.Printf("Successfully deleted repository for %v\n", repo)
		}
	}

	return nil
}

func PushContent(client *github.Client, ctx context.Context) error {

	studentReposNames, err := getStudentRepositoriesNames(client, ctx)

	if err != nil {
		return err
	}

	if len(studentReposNames) == 0 {
		return fmt.Errorf("No student repositories to push content to")
	}

	selectedRepos := []string{}

	multiPrompt := &survey.MultiSelect{
		Message: "Which repositories do you want to push content to?",
		Options: studentReposNames,
	}

	err = survey.AskOne(multiPrompt, &selectedRepos)

	if err != nil {
		return err
	}

	files, err := os.ReadDir(Config.ContentDir)

	if err != nil {
		return err
	}

	fileNames := []string{}

	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	if len(fileNames) == 0 {
		return fmt.Errorf("No files found in content directory")
	}

	selectedFiles := []string{}

	multiPrompt = &survey.MultiSelect{
		Message: "Which files do you want to push?",
		Options: fileNames,
	}

	err = survey.AskOne(multiPrompt, &selectedFiles)

	if err != nil {
		return err
	}

	commitMessage := ""

	commitPrompt := &survey.Input{
		Message: "Enter commit message:",
	}

	err = survey.AskOne(commitPrompt, &commitMessage)

	if err != nil {
		return err
	}

	if commitMessage == "" {
		return fmt.Errorf("Commit message cannot be empty")
	}

	if !yesNoPrompt(fmt.Sprintf("Are you sure you want to push %v to %v? with commit message: %v", selectedFiles, selectedRepos, commitMessage)) {
		return fmt.Errorf("Aborting...")
	}

	for _, repo := range selectedRepos {

		log.Printf("------ Pushing content to %v -------", repo)

		treeEntries := []*github.TreeEntry{}

		for _, rootFile := range selectedFiles {
			rootFilePathFullname := filepath.Join(Config.ContentDir, rootFile)
			fileInfo, err := os.Stat(rootFilePathFullname)

			if err != nil {
				return err
			}

			if !fileInfo.IsDir() {
				fmt.Println("Specified file must be a directory")

				return nil
			}

			err = filepath.Walk(rootFilePathFullname, func(filePath string, info os.FileInfo, err error) error {

				if err != nil {
					return err
				}

				if !info.IsDir() {
					fileBytes, err := os.ReadFile(filePath)

					if err != nil {
						return err
					}

					blob, _, err := client.Git.CreateBlob(ctx, Config.OrgName, repo, &github.Blob{
						Content:  github.String(string(fileBytes)),
						Encoding: github.String("utf-8"),
					})

					if err != nil {
						return err
					}

					// Strip the content directory from the path
					filePath = strings.Replace(filePath, Config.ContentDir+"/", "", 1)

					treeEntries = append(treeEntries, &github.TreeEntry{
						Path: github.String(filePath),
						Mode: github.String("100644"),
						Type: github.String("blob"),
						SHA:  blob.SHA,
					})
				}

				return nil
			})

			if err != nil {
				return err
			}
		}

		// Print all the tree entries

		for _, entry := range treeEntries {
			log.Printf("-> Repo: %v, Path: %v", repo, *entry.Path)
		}

		commit, _, err := client.Repositories.GetCommit(ctx, Config.OrgName, repo, "main", nil)

		if err != nil {
			return err
		}

		treeSHA := commit.Commit.Tree.SHA

		if treeSHA == nil {
			return fmt.Errorf("No tree SHA found for commit")
		}

		// Create tree
		treeRes, _, err := client.Git.CreateTree(ctx, Config.OrgName, repo, *treeSHA, treeEntries)

		if err != nil {
			return err
		}

		if treeRes.SHA == nil {
			return fmt.Errorf("No tree SHA found for new tree")
		}

		log.Printf("Created tree with SHA: %v", *treeRes.SHA)

		// I Have no clue why this is necessary, but it is
		tmpCommit := &github.Commit{
			SHA: commit.SHA,
		}

		// Create commit
		commitRes, _, err := client.Git.CreateCommit(ctx, Config.OrgName, repo, &github.Commit{
			Message: github.String(commitMessage),
			Tree: &github.Tree{
				SHA: treeRes.SHA,
			},
			Parents: []*github.Commit{tmpCommit},
		})

		if err != nil {
			return err
		}

		log.Printf("Created commit with SHA: %v and message: %v and author %v", *commitRes.SHA, *commitRes.Message, commitRes.Author.Name)

		if commitRes.SHA == nil {
			return fmt.Errorf("No commit SHA found for new commit")
		}

		// Update ref
		_, _, err = client.Git.UpdateRef(ctx, Config.OrgName, repo, &github.Reference{
			Ref: github.String("refs/heads/main"),
			Object: &github.GitObject{
				SHA: commitRes.SHA,
			},
		}, false)

		if err != nil {
			return err
		}

		log.Printf("Updated ref for repository: %v", repo)
	}

	return nil
}

func InviteStudents(client *github.Client, ctx context.Context) error {

	if !yesNoPrompt(fmt.Sprintf("Are you sure you want to invite students %v to the %v organization?", Config.InviteTo, Config.OrgName)) {
		return fmt.Errorf("Aborting...")
	}

	for _, student := range Config.InviteTo {

		user, _, err := client.Users.Get(ctx, student)

		log.Printf("Got id %v for user %v", user.GetID(), student)

		if err != nil {
			return err
		}

		_, _, err = client.Organizations.CreateOrgInvitation(ctx, Config.OrgName, &github.CreateOrgInvitationOptions{
			InviteeID: github.Int64(user.GetID()),
			Role:      github.String("direct_member"),
			TeamID:    []int64{},
		})

		if err != nil {
			log.Printf("Error inviting %v to %v: %v", student, Config.OrgName, err)
			continue
		}

		log.Printf("Invited %v to %v", student, Config.OrgName)
	}

	return nil
}

func CheckForInvites(client *github.Client, ctx context.Context) error {

	invitations, _, err := client.Organizations.ListPendingOrgInvitations(ctx, Config.OrgName, nil)

	if err != nil {
		return err
	}

	if len(invitations) == 0 {
		log.Printf("No pending invitations found for %v", Config.OrgName)
		return nil
	}

	for _, invitation := range invitations {
		log.Printf("Found pending invitation for %v", invitation.GetLogin())
	}

	return nil
}
