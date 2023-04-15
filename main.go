package main

import (
	"krozek-automate/cmd"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "list-students",
				Usage:   "List all the students in the organization",
				Aliases: []string{"ls"},

				Action: func(cliCtx *cli.Context) error {
					return cmd.ExecuteCommand(cliCtx, cmd.ListStudents)
				},
			},
			{
				Name:    "create-repos",
				Usage:   "Create repositories for each student",
				Aliases: []string{"cr"},

				Action: func(cliCtx *cli.Context) error {
					return cmd.ExecuteCommand(cliCtx, cmd.CreateRepos)
				},
			},
			{
				Name:    "delete-repos",
				Usage:   "Delete student repositories",
				Aliases: []string{"del"},
				Action: func(cliCtx *cli.Context) error {
					return cmd.ExecuteCommand(cliCtx, cmd.DeleteStudentRepositories)
				},
			},
			{
				Name:    "push",
				Usage:   "Push the code to the student repositories",
				Aliases: []string{"p"},
				Action: func(cliCtx *cli.Context) error {
					return cmd.ExecuteCommand(cliCtx, cmd.PushContent)
				},
			},
			{
				Name:    "invite-students",
				Usage:   "Invite students to the organization",
				Aliases: []string{"inv"},
				Action: func(cliCtx *cli.Context) error {
					return cmd.ExecuteCommand(cliCtx, cmd.InviteStudents)
				},
			},

			{
				Name:    "check-invitations",
				Usage:   "Check if all students have accepted the invitation",
				Aliases: []string{"ci"},
				Action: func(cliCtx *cli.Context) error {
					return cmd.ExecuteCommand(cliCtx, cmd.CheckForInvites)
				},
			},
		},
	}

	if err := cmd.InitConfig(); err != nil {
		log.Fatal(err)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
