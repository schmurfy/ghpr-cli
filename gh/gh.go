package gh

import (
	"context"

	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	*github.Client

	currentUser *github.User
	owner       string
	repository  string
}

func New(ctx context.Context, token string) *GithubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	cl := github.NewClient(tc)

	user, _, err := cl.Users.Get(ctx, "")
	if err != nil {
		panic(err)
	}

	return &GithubClient{
		Client:      cl,
		currentUser: user,
	}
}

func (gh *GithubClient) GetCurrentUser() *github.User {
	return gh.currentUser
}

func (gh *GithubClient) GetPullRequestHead(ctx context.Context, owner string, repository string, prNumber int) *github.PullRequestBranch {
	pr, _, err := gh.PullRequests.Get(ctx, owner, repository, prNumber)
	if err != nil {
		panic(err)
	}

	return pr.GetHead()
}
