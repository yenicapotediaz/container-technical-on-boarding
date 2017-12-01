package onboarding

import (
	"testing"
)

const testYamlDataFixture = `
githubOrganization: testOrg
githubRepository: testRepo
clientId: testApplicationID
clientSecret: testApplicationSecret
task_owners:
    testOwner1: &owner_one
        github_username: "{{ index .Environ "GITHUB_USER" }}"
    testOwner2: &owner_two
        github_username: testUsername2
tasks:
    - title: "This is a Test Task"
      assignee: *owner_one
      description: test task one.
    - title: second test task
      assignee: *owner_two
      description: |
        this is a test.
        all of this text should be included.
        the description will end with the word "Acorn".
        Yes.
        Acorn
`

func TestConfigYAMLParser(t *testing.T) {
	scheme := SetupScheme{}
	environ := map[string]string{
		"GITHUB_USER": "testUsername1",
	}
	err := scheme.ingest([]byte(testYamlDataFixture), &environ)

	if err != nil {
		t.Fatalf("Loading sample YAML failed with error: %v", err)
	}

	if len(scheme.TaskOwners) != 2 {
		t.Errorf("Tasks counted actual: %d, expected: %d", len(scheme.TaskOwners), 2)
	}

	testCases := []struct {
		expect  string
		actual  string
		message string
	}{
		{scheme.GithubOrganization, "testOrg", "YAML Organization mismatch, actual: %v, expected %v"},
		{scheme.GithubRepository, "testRepo", "YAML Repository mismatch, actual: %v, expected %v"},
		{scheme.ClientID, "testApplicationID", "YAML ClientID mismatch, actual: %v, expected %v"},
		{scheme.ClientSecret, "testApplicationSecret", "YAML ClientID mismatch, actual: %v, expected %v"},
		{scheme.Tasks[0].Assignee.String(), "testUsername1", "YAML Task Owner mismatch, actual %v, expected %v"},
		{scheme.Tasks[1].Title, "second test task", "YAML Task Title mismatch, actual %v, expected %v"},
	}

	for _, thisCase := range testCases {
		if thisCase.expect != thisCase.actual {
			t.Errorf(thisCase.message, thisCase.actual, thisCase.expect)
		}
	}

}
