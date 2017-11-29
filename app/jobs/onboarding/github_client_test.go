package onboarding

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

/*
This module implements several mock methods to the interfaces
RepositoryAccess and ClientAccess from `workload.go`
*/

type (
	TestIssues struct {
		Cache *map[string](interface{})
	}

	TestRepositories struct {
		Cache *map[string](interface{})
	}

	TestProjects struct {
		Cache *map[string](interface{})
	}

	TestUsers struct {
		Cache *map[string](interface{})
	}

	TestGitHubClient struct {
		// implements GitHubClientInterface
		Cache map[string](interface{})
	}
)

func prepareGitHubClientTest() *WorkflowClient {
	cache := make(map[string](interface{}))
	context := oauth2.NoContext
	return &WorkflowClient{context, TestGitHubClient{cache}}
}

func (client TestGitHubClient) getIssuesService() iGitHubIssues {
	return &TestIssues{&client.Cache}
}

func (client TestGitHubClient) getRepositoriesService() iGitHubRepositories {
	return &TestRepositories{&client.Cache}
}

func (client TestGitHubClient) getProjectsService() iGitHubProjects {
	return &TestProjects{&client.Cache}
}

func (client TestGitHubClient) getUsersService() iGitHubUsers {
	return &TestUsers{&client.Cache}
}

func ClearCache(client *TestGitHubClient) {
	client.Cache = make(map[string](interface{}))
}

func prepareGitHubAPIResponse() *github.Response {
	return &github.Response{
		Response:  nil,
		NextPage:  0,
		PrevPage:  0,
		FirstPage: 0,
		LastPage:  0,
		Rate: github.Rate{
			Limit:     1,
			Remaining: 100,
			Reset:     github.Timestamp{Time: time.Now()},
		},
	}
}

func (users *TestUsers) Get(ctx context.Context, username string) (*github.User, *github.Response, error) {
	realname := "Test User"
	return &github.User{
		Login: &username,
		Name:  &realname,
	}, nil, nil
}

func (issues *TestIssues) ListMilestones(ctx context.Context, owner string, repo string, opts *github.MilestoneListOptions) ([]*github.Milestone, *github.Response, error) {
	return nil, prepareGitHubAPIResponse(), nil
}

func (issues *TestIssues) CreateMilestone(ctx context.Context, owner string, repo string, opts *github.Milestone) (*github.Milestone, *github.Response, error) {

	counter, ok := ((*issues.Cache)["milestoneCounter"]).(int)

	if !ok {
		counter = 1
	} else {
		counter = counter + 1
	}

	(*issues.Cache)["milestoneCounter"] = counter

	creationTime := time.Now()
	opts.CreatedAt = &creationTime
	opts.Number = &counter

	cacheKey := fmt.Sprintf("milestone/%d", opts.GetNumber())
	(*issues.Cache)[cacheKey] = opts

	milestones, ok := ((*issues.Cache)["milestones"]).([]*github.Milestone)

	if !ok {
		milestones = make([]*github.Milestone, 0)
	}

	milestones = append(milestones, opts)
	(*issues.Cache)["milestones"] = milestones

	return opts, nil, nil

}

func listContainsString(list []string, target string) {

}

func assigneeListContainsUsername(list []*github.User, username string) bool {
	for _, u := range list {
		if username == u.GetLogin() {
			return true
		}
	}
	return false
}

func (issues *TestIssues) ListByRepo(ctx context.Context, owner string, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error) {
	var resultIssues []*github.Issue
	var matchingIssues []*github.Issue

	resultIssues, ok := (*issues.Cache)["issues"].([]*github.Issue)

	if !ok {
		resultIssues = make([]*github.Issue, 0)
		(*issues.Cache)["issues"] = resultIssues
	}

	// Save the cache
	(*issues.Cache)["issues"] = resultIssues

	for _, issue := range resultIssues {
		matched := false

		if opts.Assignee == "*" {
			matched = true // initially anyway
		} else if opts.Assignee == "none" {
			matched = (issue.Assignee == nil)
		} else if len(opts.Assignee) > 0 && opts.Assignee != "none" {
			matched = assigneeListContainsUsername(issue.Assignees, opts.Assignee)
		}

		if !matched {
			// log.Printf("(GitHub API) Issue #%d did not match assignee; search: %v, current: %v", issue.GetNumber(), opts.Assignee, issue.Assignees)
			continue
		}

		if opts.Milestone == "none" {
			matched = matched && (issue.Milestone == nil)
		} else if len(opts.Milestone) > 0 && issue.Milestone != nil {
			targetNumber, err := strconv.Atoi(opts.Milestone)
			matched = matched && (issue.Milestone.GetNumber() == targetNumber && err == nil)
		} else if opts.Milestone == "*" {
			matched = true
		}

		if matched {
			matchingIssues = append(matchingIssues, issue)
		}
	}

	return matchingIssues, prepareGitHubAPIResponse(), nil
}

func (issues *TestIssues) Create(ctx context.Context, owner string, repo string, req *github.IssueRequest) (*github.Issue, *github.Response, error) {

	resultIssues, ok := (*issues.Cache)["issues"].([]*github.Issue)

	if !ok {
		resultIssues = make([]*github.Issue, 0)
		(*issues.Cache)["issues"] = resultIssues
	}

	issueNumber := len(resultIssues) + 1

	if req.Assignees == nil && req.Assignee != nil {
		// GitHub is deprecating the single-assignee attribute, as Issues can properly support multiple.
		// As such we coerce this model accordingly.
		assigneeList := []string{*req.Assignee}
		req.Assignees = &assigneeList
		req.Assignee = nil
	}

	userList := []*github.User{}

	for _, name := range *req.Assignees {
		userName := name
		userList = append(userList, &github.User{
			Login: &userName,
		})
	}

	thisIssue := github.Issue{
		ID:        &issueNumber,
		Number:    &issueNumber,
		Title:     req.Title,
		Body:      req.Body,
		Assignee:  nil,
		Assignees: userList,
		Milestone: &github.Milestone{
			Number: req.Milestone,
		},
	}

	// log.Printf("(GitHub API) Created issue: #%d: %v", thisIssue.GetNumber(), thisIssue.GetTitle())

	resultIssues = append(resultIssues, &thisIssue)

	// Save the cache
	(*issues.Cache)["issues"] = resultIssues

	return &thisIssue, nil, nil
}

func (issues *TestIssues) Edit(ctx context.Context, owner string, repo string, issueID int, req *github.IssueRequest) (*github.Issue, *github.Response, error) {

	thisIssue := github.Issue{
		ID:    &issueID,
		Title: req.Title,
		Body:  req.Body,
		Milestone: &github.Milestone{
			Number: req.Milestone,
		},
	}

	return &thisIssue, nil, nil

}

func (repos *TestRepositories) CreateProject(ctx context.Context, owner string, repo string, opts *github.ProjectOptions) (*github.Project, *github.Response, error) {
	projectNumber := 1001
	createTimestamp := github.Timestamp{Time: time.Now()}
	thisProject := github.Project{
		Name:      &opts.Name,
		Body:      &opts.Body,
		Number:    &projectNumber,
		CreatedAt: &createTimestamp,
		UpdatedAt: &createTimestamp,
	}

	return &thisProject, prepareGitHubAPIResponse(), nil
}

// TODO: migrate this out to a separate seeding function for the cache.
func (repos *TestRepositories) ListProjects(ctx context.Context, owner string, repo string, opts *github.ProjectListOptions) ([]*github.Project, *github.Response, error) {

	var resultProjects = make([]*github.Project, 3)

	names := []string{"Project 1", "Project 2", "Project 3"}
	bodies := []string{"This is a project", "This is a project", "This is a project"}

	for i, v := range []int{1, 2, 3} {
		resultProjects[i] = &github.Project{
			Number: &v,
			Name:   &names[i],
			Body:   &bodies[i],
		}
	}

	return resultProjects, prepareGitHubAPIResponse(), nil

}

func (repos *TestRepositories) Get(ctx context.Context, owner string, repo string) (*github.Repository, *github.Response, error) {

	thisRepo := &github.Repository{ // Set up a fake GitHub repo model
		Owner: &github.User{
			Login: &owner,
		},
		Name: &repo,
	}

	return thisRepo, nil, nil
}

func (proj *TestProjects) UpdateProject(ctx context.Context, projectID int, opts *github.ProjectOptions) (*github.Project, *github.Response, error) {

	projectNumberUpdate := 1002 // Normally this should not be needed, if target project already has a number.
	updateTimestamp := github.Timestamp{Time: time.Now()}

	thisProject := github.Project{
		ID:        &projectID,
		Name:      &opts.Name,
		Body:      &opts.Body,
		Number:    &projectNumberUpdate,
		CreatedAt: &updateTimestamp, // Not really proper. Hopefully no one is testing against CreatedAt's specific value.
		UpdatedAt: &updateTimestamp,
	}

	return &thisProject, prepareGitHubAPIResponse(), nil
}

func (proj *TestProjects) CreateProjectColumn(ctx context.Context, projectID int, opts *github.ProjectColumnOptions) (*github.ProjectColumn, *github.Response, error) {

	cache := *proj.Cache

	counter, ok := (cache["columnCount"]).(int)

	if !ok {
		counter = 1
	}

	if counter < 1 {
		counter = 1000
	}

	thisColumn := github.ProjectColumn{
		ID:        &counter,
		Name:      &opts.Name,
		CreatedAt: &github.Timestamp{Time: time.Now()},
	}

	cache["columnCount"] = counter + 1

	cacheKey := fmt.Sprintf("column/%d", *thisColumn.ID)
	cache[cacheKey] = &thisColumn

	cacheParentKey := fmt.Sprintf("project/%d/columns", projectID)

	parent, ok := (cache[cacheParentKey]).([]*github.ProjectColumn)
	if !ok {
		parent = make([]*github.ProjectColumn, 0)
	}
	parent = append(parent, &thisColumn)
	cache[cacheParentKey] = parent

	return &thisColumn, nil, nil
}

func (proj *TestProjects) ListProjectColumns(ctx context.Context, projectID int, opts *github.ListOptions) ([]*github.ProjectColumn, *github.Response, error) {

	cacheKey := fmt.Sprintf("project/%d/columns", projectID)
	ptrCache, ok := ((*proj.Cache)[cacheKey]).([]*github.ProjectColumn)

	if !ok {
		(*proj.Cache)[cacheKey] = make([]*github.ProjectColumn, 0)
	}

	return ptrCache, prepareGitHubAPIResponse(), nil
}

func (proj *TestProjects) CreateProjectCard(ctx context.Context, columnID int, opt *github.ProjectCardOptions) (*github.ProjectCard, *github.Response, error) {

	cacheKey := fmt.Sprintf("cards/column/%d", columnID)
	ptrCache, ok := ((*proj.Cache)[cacheKey]).([]*github.ProjectCard)

	if !ok {
		(*proj.Cache)[cacheKey] = make([]*github.ProjectCard, 0)
		ptrCache = ((*proj.Cache)[cacheKey]).([]*github.ProjectCard)
	}

	count := len(ptrCache) + 1

	card := github.ProjectCard{
		ID:        &count,
		CreatedAt: &github.Timestamp{Time: time.Now()},
	}

	// Save to cache
	ptrCache = append(ptrCache, &card)
	(*proj.Cache)[cacheKey] = ptrCache

	return &card, nil, nil
}

func (proj *TestProjects) ListProjectCards(ctx context.Context, columnID int, opt *github.ListOptions) ([]*github.ProjectCard, *github.Response, error) {
	cacheKey := fmt.Sprintf("cards/column/%d", columnID)
	ptrCache, ok := ((*proj.Cache)[cacheKey]).([]*github.ProjectCard)

	if !ok {
		(*proj.Cache)[cacheKey] = make([]*github.ProjectCard, 0)
	}

	return ptrCache, prepareGitHubAPIResponse(), nil
}
