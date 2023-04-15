# Automating course exercises 

A `.env.toml` is required to run the program. It should contain the fields found in the example config.


```sh
❯ go run main.go help
NAME:
   main - A new cli application

USAGE:
   main [global options] command [command options] [arguments...]

COMMANDS:
   list-students, ls      List all the students in the organization
   create-repos, cr       Create repositories for each student
   delete-repos, del      Delete student repositories
   push, p                Push the code to the student repositories
   invite-students, inv   Invite students to the organization
   check-invitations, ci  Check if all students have accepted the invitation
   help, h                Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```