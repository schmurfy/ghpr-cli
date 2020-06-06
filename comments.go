package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-github/v31/github"
	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"

	"github.com/schmurfy/ghpr-cli/gh"
)

type comments struct {
	ui cli.Ui

	issueNumber int
}

func (st *comments) Help() string {
	return showUsage("comments", st.flags())
}

func (st *comments) Synopsis() string {
	return "Manipulate comments"
}

func (st *comments) flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	fs.IntVarP(&st.issueNumber, "issue", "i", -1, "issue number")
	setRequiredFlags(fs, "issue")

	return fs
}

////////
/// comments list

type commentsList struct {
	*comments

	onlyMine bool
}

func (st *commentsList) Help() string {
	return showUsage("comments list", st.flags())
}

func (st *commentsList) flags() *pflag.FlagSet {
	fs := st.comments.flags()
	fs.BoolVarP(&st.onlyMine, "mine", "", false, " show only user comment")
	return fs
}

func (st *commentsList) Run(args []string) int {
	parseFlags(st.ui, st.flags(), args)
	cl := gh.New(ctx, githubToken)

	opts := &github.IssueListCommentsOptions{}
	comments, _, err := cl.Issues.ListComments(ctx, owner, repository, st.issueNumber, opts)
	if err != nil {
		panic(err)
	}
	for _, comment := range comments {
		show := true

		if st.onlyMine && (comment.GetUser().GetID() != cl.GetCurrentUser().GetID()) {
			show = false
		}

		if show {
			printJsonComment(comment)
		}
	}

	return 0
}

////////
/// comments create

type commentsCreate struct {
	*comments

	body string
}

func (st *commentsCreate) Help() string {
	return showUsage("comments create", st.flags())
}

func (st *commentsCreate) flags() *pflag.FlagSet {
	fs := st.comments.flags()
	fs.StringVarP(&st.body, "body", "", "", " comment body")
	setRequiredFlags(fs, "body")

	return fs
}

func (st *commentsCreate) Run(args []string) int {
	parseFlags(st.ui, st.flags(), args)
	cl := gh.New(ctx, githubToken)

	body := getBody(st.body)

	comment := &github.IssueComment{
		Body: &body,
	}
	comment, _, err := cl.Issues.CreateComment(ctx, owner, repository, st.issueNumber, comment)
	if err != nil {
		panic(err)
	}

	printJsonComment(comment)
	return 0
}

////////
/// comments update

type commentsUpdate struct {
	*comments

	body              string
	commentID         int64
	updateLastComment bool
}

func (st *commentsUpdate) Help() string {
	return showUsage("comments update", st.flags())
}

func (st *commentsUpdate) flags() *pflag.FlagSet {
	fs := st.comments.flags()
	fs.StringVarP(&st.body, "body", "", "", " comment body")
	fs.BoolVarP(&st.updateLastComment, "last", "", false, "update last of my comment")
	fs.Int64VarP(&st.commentID, "id", "", -1, "comment id")
	setRequiredFlags(fs, "body")

	return fs
}

func (st *commentsUpdate) Run(args []string) int {
	parseFlags(st.ui, st.flags(), args)
	cl := gh.New(ctx, githubToken)

	body := getBody(st.body)
	commentID := st.commentID

	if st.updateLastComment {
		// find last comment id
		sort := "created_at"
		opts := &github.IssueListCommentsOptions{
			Sort: &sort,
		}
		comments, _, err := cl.Issues.ListComments(ctx, owner, repository, st.issueNumber, opts)
		if err != nil {
			panic(err)
		}

		for i := len(comments) - 1; i >= 0; i-- {
			c := comments[i]
			if c.GetUser().GetID() == cl.GetCurrentUser().GetID() {
				commentID = c.GetID()
				break
			}
		}

		if commentID == -1 {
			panic("no comment found for user")
		}
	}

	comment := &github.IssueComment{
		Body: &body,
	}

	comment, _, err := cl.Issues.EditComment(ctx, owner, repository, commentID, comment)
	if err != nil {
		panic(err)
	}

	printJsonComment(comment)
	return 0
}

func getBody(commentBody string) string {
	if commentBody == "-" {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		commentBody = string(data)
	}
	return commentBody
}

func printJsonComment(comment *github.IssueComment) {
	buffer, err := json.Marshal(&comment)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(buffer))
}

func initComments(ui cli.Ui) {
	cm := &comments{ui: ui}

	Register("comments list", func() (cli.Command, error) {
		return &commentsList{comments: cm}, nil
	})

	Register("comments create", func() (cli.Command, error) {
		return &commentsCreate{comments: cm}, nil
	})

	Register("comments update", func() (cli.Command, error) {
		return &commentsUpdate{comments: cm}, nil
	})

	// Register("comments update", func() (cli.Command, error) {
	// 	st := &statuses{ui: ui}
	// 	return &statusesSet{statuses: st}, nil
	// })
}
