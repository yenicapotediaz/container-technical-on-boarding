/*	This package abstracts authentication with GitHub's API.
	Primarily this extract OAuth integration from the rest of the business logic,
	enabling better testing of business logic without dependency on GitHub's actual
	service.
*/

package onboarding

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/go-github/github"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

type (
	// AuthEnvironment provides a simple model for OAuth context abstraction.
	AuthEnvironment struct {
		Context        context.Context
		Config         *oauth2.Config
		StateString    string
		AccessToken    *oauth2.Token
		workflowClient *WorkflowClient
	}

	// Credentials for a GitHub application, integrating with GitHub API.
	Credentials struct {
		ClientID     string
		ClientSecret string
		Scopes       []string
	}
)

// NewAuthEnvironment prepares a new OAuth2 authenticated GitHub login environment.
func (creds *Credentials) NewAuthEnvironment() *AuthEnvironment {
	var (
		authContext = oauth2.NoContext
		config      = oauth2.Config{
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
			Scopes:       creds.Scopes,
			Endpoint:     githuboauth.Endpoint,
		}
		oauthStateString = uuid.NewV4().String()
	)
	return &AuthEnvironment{
		Context:     authContext,
		Config:      &config,
		StateString: oauthStateString,
	}
}

// AuthCodeURL gets and returns the oauth2 providers authorization URL
func (auth *AuthEnvironment) AuthCodeURL() string {
	oauthStateString := auth.StateString
	oauthConf := *auth.Config
	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	return url
}

func (auth *AuthEnvironment) newWorkflowClient() (*WorkflowClient, error) {
	if auth.workflowClient != nil {
		return auth.workflowClient, nil
	}

	if auth.AccessToken == nil {
		return nil, errors.New("AccessToken is not setup")
	}

	oauthClient := auth.Config.Client(auth.Context, auth.AccessToken)
	githubClient := github.NewClient(oauthClient)
	workflow := WorkflowClient{auth.Context, NewGitHubWrapper(githubClient)}
	return &workflow, nil
}

// SetupAccessToken gets and sets an oauth2 access token based on a recieved oauth2 code
func (auth *AuthEnvironment) SetupAccessToken(code string) (*oauth2.Token, error) {
	token, err := auth.Config.Exchange(auth.Context, code)
	if err != nil {
		msg := fmt.Sprintf("OAuth Exchange failed: %v", err)
		return nil, errors.New(msg)
	}
	auth.AccessToken = token

	return token, nil

}

// GithubUsername retrives the github username via that git api
func (auth *AuthEnvironment) GithubUsername() string {
	if auth.AccessToken == nil {
		return ""
	}

	oauthClient := auth.Config.Client(auth.Context, auth.AccessToken)
	githubClient := github.NewClient(oauthClient)
	githubUser, _, err := githubClient.Users.Get(auth.Context, "")
	if err != nil {
		log.Printf("Failed to get github user: %v", err)
	}
	return *githubUser.Login
}
