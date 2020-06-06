package main

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v31/github"
	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"

	"github.com/schmurfy/ghpr-cli/gh"
)

type statuses struct {
	ui cli.Ui

	prNumber int
}

func (st *statuses) Help() string {
	return showUsage("statuses", st.flags())
}

func (st *statuses) Synopsis() string {
	return "Manipulate statuses"
}

func (st *statuses) flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	fs.IntVarP(&st.prNumber, "pr", "p", -1, "pull-request number")
	setRequiredFlags(fs, "pr")

	return fs
}

func (st *statuses) Errorf(format string, args ...interface{}) {
	st.ui.Error(fmt.Sprintf(format, args...))
}

////////
/// status list

type statusesList struct {
	*statuses
}

func (st *statusesList) Help() string {
	return showUsage("statuses list", st.flags())
}

func (st *statusesList) Run(args []string) int {
	parseFlags(st.ui, st.flags(), args)
	cl := gh.New(ctx, githubToken)

	ref := cl.GetPullRequestHead(ctx, owner, repository, st.prNumber).GetRef()
	opts := &github.ListOptions{}
	statuses, _, err := cl.Repositories.ListStatuses(ctx, owner, repository, ref, opts)
	if err != nil {
		panic(err)
	}

	for _, st := range statuses {
		printJsonStatus(st)
	}

	return 0
}

////////
/// status set

type statusesSet struct {
	*statuses

	state     string
	desc      string
	context   string
	targetURL string
}

func (st *statusesSet) Help() string {
	return showUsage("statuses set", st.flags())
}

func (st *statusesSet) flags() *pflag.FlagSet {
	fs := st.statuses.flags()

	fs.StringVarP(&st.state, "state", "s", "", "set state")
	setAllowedValues(fs, "state", []string{"error", "failure", "pending", "success"})

	fs.StringVarP(&st.desc, "description", "d", "", "set description")
	fs.StringVarP(&st.context, "context", "c", "", "set context")
	fs.StringVarP(&st.targetURL, "url", "u", "", "set target url")

	setRequiredFlags(fs, "state", "description", "context")

	return fs
}

func (st *statusesSet) Run(args []string) int {
	parseFlags(st.ui, st.flags(), args)
	cl := gh.New(ctx, githubToken)

	sha1 := cl.GetPullRequestHead(ctx, owner, repository, st.prNumber).GetSHA()
	status := &github.RepoStatus{
		State:       &st.state,
		Description: &st.desc,
		Context:     &st.context,
	}

	if st.targetURL != "" {
		status.TargetURL = &st.targetURL
	}

	fmt.Printf("CreateStatus(%s, %s, %s, %+v)\n", owner, repository, sha1, status)
	status, _, err := cl.Repositories.CreateStatus(ctx, owner, repository, sha1, status)
	if err != nil {
		panic(err)
	}

	printJsonStatus(status)
	return 0
}

func printJsonStatus(status *github.RepoStatus) {
	buffer, err := json.Marshal(&status)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(buffer))
}

func initStatuses(ui cli.Ui) {
	Register("statuses list", func() (cli.Command, error) {
		st := &statuses{ui: ui}
		return &statusesList{st}, nil
	})

	Register("statuses set", func() (cli.Command, error) {
		st := &statuses{ui: ui}
		return &statusesSet{statuses: st}, nil
	})
}
