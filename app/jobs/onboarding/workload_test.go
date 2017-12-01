package onboarding

/*
This module's tests focus on exercising the `workload.go` module.
It requires the GitHub Client mock/fixtures implemented in `github_client_test.go`
*/

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/go-github/github"
	"github.com/samsung-cnct/container-technical-on-boarding/app/jobs"
)

// NOTE: This embodies an assumption of business process.
// In particular, a 21-day onboarding cycle that ends on the Friday of the third full week following
// a new-hire's start date.
// TODO: convert these tests to Go's table-style.
func TestMilestoneDueTimeCalculator(t *testing.T) {

	// First test the default case (measuring from "now")
	now := time.Now()
	fromToday := getMilestoneDueTime(nil)
	duration := fromToday.Sub(now)

	assert(t, ((duration.Hours() / 24) >= 21), "Milestone duration was less than 21 days; (actual %f hours)", duration.Hours())

	// Seed cases with specific known outcomes.
	startMonday := time.Date(2017, 06, 19, 0, 0, 0, 0, time.Local)
	startTuesday := time.Date(2017, 06, 20, 0, 0, 0, 0, time.Local)
	startWednesday := time.Date(2017, 06, 21, 0, 0, 0, 0, time.Local)
	startThursday := time.Date(2017, 06, 22, 0, 0, 0, 0, time.Local)
	startFriday := time.Date(2017, 06, 23, 0, 0, 0, 0, time.Local)
	startSaturday := time.Date(2017, 06, 24, 0, 0, 0, 0, time.Local)
	startSunday := time.Date(2017, 06, 25, 0, 0, 0, 0, time.Local)

	resultMonday := getMilestoneDueTime(&startMonday)
	resultTuesday := getMilestoneDueTime(&startTuesday)
	resultWednesday := getMilestoneDueTime(&startWednesday)
	resultThursday := getMilestoneDueTime(&startThursday)
	resultFriday := getMilestoneDueTime(&startFriday)
	resultSaturday := getMilestoneDueTime(&startSaturday)
	resultSunday := getMilestoneDueTime(&startSunday)

	assertEqual(t, int(resultMonday.Sub(startMonday).Hours()), 25*24, "Duration from Monday start, actual %d, expected %d")
	assertEqual(t, int(resultTuesday.Sub(startTuesday).Hours()), 24*24, "Duration from Tuesday start, actual %d, expected %d")
	assertEqual(t, int(resultWednesday.Sub(startWednesday).Hours()), 23*24, "Duration from Wednesday start, actual %d, expected %d")
	assertEqual(t, int(resultThursday.Sub(startThursday).Hours()), 22*24, "Duration from Thursday start, actual %d, expected %d")
	assertEqual(t, int(resultFriday.Sub(startFriday).Hours()), 21*24, "Duration from Friday start, actual %d, expected %d")

	// Employees don't start on weekends. Test anyway.
	assertEqual(t, int(resultSaturday.Sub(startSaturday).Hours()), 20*24, "Duration from Saturday start, actual %d, expected %d")
	assertEqual(t, int(resultSunday.Sub(startSunday).Hours()), 26*24, "Duration from Sunday start, actual %d, expected %d")
}

func nvl(values ...string) *string {
	for _, v := range values {
		if len(v) > 0 {
			return &v
		}
	}
	return nil
}

func TestCreateClientAndResolveUser(t *testing.T) {
	client := prepareGitHubClientTest()
	testUsername := "testingisawesome"
	user := client.resolveUser(&testUsername)

	verify := []struct {
		expect string
		actual string
	}{
		{testUsername, user.GetLogin()},
		{"Test User", user.GetName()},
	}

	for _, testCase := range verify {
		if testCase.expect != testCase.actual {
			t.Errorf("User Setup Failed, expected: %v, actual: %v", testCase.expect, testCase.actual)
		}
	}
}

func TestProjectCreateAndUpdate(t *testing.T) {
	client := prepareGitHubClientTest()

	repo, err := client.GetRepository("testowner", "testrepo")

	if err != nil {
		t.Errorf("GetRepository produced an error?! %v", err)
	}

	projectName := "testproject"
	projectDescription := "this is a test project. super awesome."
	projectColumns := []string{"test column 1", "test column 2"}
	mapColumnIndices := make(map[string]int)

	for index, lbl := range projectColumns {
		mapColumnIndices[lbl] = index
	}

	project, err := repo.CreateOrUpdateProject(&projectName, &projectDescription, projectColumns)

	if err != nil {
		t.Errorf("CreateOrUpdateProject produced an error?! %v", err)
	}

	columnsFound, err := repo.FetchMappedProjectColumns(project)

	for lbl, col := range columnsFound {
		index := mapColumnIndices[lbl]
		if lbl != projectColumns[index] {
			t.Errorf("Column name did not match; expected: %v, actual %v", projectColumns[index], col.GetName())
		}
	}

}

func TestCreateIssuesMilestones(t *testing.T) {
	client := prepareGitHubClientTest()

	var resultIssues []*github.Issue

	repo, _ := client.GetRepository("testowner", "testrepo")
	duedate := getMilestoneDueTime(nil)
	milestoneName := "Test Milestone"
	milestoneDescription := "This stone is a mile high."

	milestone, _ := repo.CreateOrUpdateMilestone(&milestoneName, &milestoneDescription, &duedate)

	issues := []struct {
		title       string
		description string
		assignee    string
		index       int
	}{
		{"Issue #1", "First Issue", "testuser1", 1},
		{"Issue #2", "Second Issue", "testuser2", 2},
		{"Issue #3", "Third Issue", "testuser3", 3},

		// We intentionally duplicate title here.
		// This should alter the result of the first issue, rather than producing a fourth issue.
		{"Issue #1", "Awesome Issue", "testuser1", 1},
	}

	for _, i := range issues {
		title := i.title
		assignee := i.assignee
		description := i.description
		thisIssue, _ := repo.CreateOrUpdateIssue(&assignee, &title, &description, milestone.GetNumber())
		resultIssues = append(resultIssues, thisIssue)
	}

	issuesFound, _ := repo.GetIssuesByRequest(&github.IssueRequest{
		Milestone: milestone.Number,
	})

	if len(issuesFound) != 3 {
		t.Errorf("Expected %d issues, found %d", 3, len(issuesFound))
	}

	for _, issue := range issuesFound {
		if issue.Milestone == nil {
			t.Errorf("Expected milestone for Issue #%d", issue.GetNumber())
			break
		}
		if issue.Milestone.GetNumber() != milestone.GetNumber() {
			issueNumber := issue.GetNumber()
			milestoneNumberFound := issue.Milestone.GetNumber()
			milestoneNum := milestone.GetNumber()
			t.Errorf("Expected milestone for Issue #%d to be %d, found %d", issueNumber, milestoneNum, milestoneNumberFound)
		}
	}

	cachedClient, ok := client.Client.(TestGitHubClient)

	if ok {
		ClearCache(&cachedClient)
	}

}

func TestCreateIssuesCards(t *testing.T) {
	client := prepareGitHubClientTest()

	var resultIssues []*github.Issue
	var resultCards []*github.ProjectCard

	repo, _ := client.GetRepository("testowner", "testrepo")

	projectName := "testproject"
	projectDescription := "this is a test project. super awesome."
	projectColumns := []string{"backlog", "in progress", "review", "done"}
	mapColumnIndices := make(map[string]int)

	for index, lbl := range projectColumns {
		mapColumnIndices[lbl] = index
	}

	project, _ := repo.CreateOrUpdateProject(&projectName, &projectDescription, projectColumns)

	columns, _ := repo.FetchMappedProjectColumns(project)

	issues := []struct {
		title       string
		description string
		assignee    string
		index       int
	}{
		{"Issue #1", "First Issue", "testuser1", 1},
		{"Issue #2", "Second Issue", "testuser2", 2},
		{"Issue #3", "Third Issue", "testuser3", 3},

		// We intentionally duplicate title here.
		// This should alter the result of the first issue, rather than producing a fourth issue.
		{"Issue #1", "Awesome Issue", "testuser1", 1},
	}

	for _, i := range issues {
		title := i.title
		assignee := i.assignee
		description := i.description
		thisIssue, _ := repo.CreateOrUpdateIssue(&assignee, &title, &description, 0)

		// cache them...
		resultIssues = append(resultIssues, thisIssue)

		card, _ := repo.CreateCardForIssue(thisIssue, columns["backlog"])
		resultCards = append(resultCards, card)
	}

	for _, card := range resultCards {
		if card.GetID() < 1 {
			t.Errorf("Invalid card? %#v", card)
		}
	}

}

func TestFullWorkload(t *testing.T) {
	client := prepareGitHubClientTest()
	creds := Credentials{
		ClientID:     "TEST_CLIENT_ID",
		ClientSecret: "TEST_CLIENT_SECRET",
		Scopes:       []string{"user", "repo", "issues", "milestones"},
	}
	setup := SetupScheme{
		ClientID:           creds.ClientID,
		ClientSecret:       creds.ClientSecret,
		GithubOrganization: "testOrganization",
		GithubRepository:   "testRepository",
		Tasks: []TaskEntry{
			{Title: "test1", Description: "test", Assignee: indirectAssignee{GithubUsername: "test"}},
			{Title: "test2", Description: "test", Assignee: indirectAssignee{GithubUsername: "test"}},
			{Title: "test3", Description: "test", Assignee: indirectAssignee{GithubUsername: "test"}},
		},
		TaskOwners: map[string]indirectAssignee{
			"new_hire": {GithubUsername: "test"},
		},
	}

	events := make(chan jobs.Event)
	job := GenerateProject{
		ID:      42,
		Setup:   &setup,
		AuthEnv: &AuthEnvironment{workflowClient: client},
		New:     events,
	}

	go job.Run()
	for event := range events {
		switch event.Text {
		case "error":
			log.Fatalf("Something broke in full workload. %v", event.Error)
		default:
			fmt.Println(event.Text)
		}
	}
}
