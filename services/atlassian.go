package services

import (
	"log"
	"os"
	"sync"

	"github.com/ctreminiom/go-atlassian/confluence"
	jira "github.com/ctreminiom/go-atlassian/jira/v2"
	"github.com/pkg/errors"
)

var ConfluenceClient = sync.OnceValue[*confluence.Client](func() *confluence.Client {
	var (
		//Host is the URL of the Confluence instance
		host = os.Getenv("ATLASSIAN_HOST")
		// Mail is the email of the user
		mail  = os.Getenv("ATLASSIAN_EMAIL")
		// Token is the API token of the user
		token = os.Getenv("ATLASSIAN_TOKEN")
	)

	if host == "" || mail == "" || token == "" {
		log.Fatal("ATLASSIAN_HOST, ATLASSIAN_EMAIL, ATLASSIAN_TOKEN are required")
	}

	instance, err := confluence.New(nil, host)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "failed to create confluence client"))
	}

	instance.Auth.SetBasicAuth(mail, token)

	return instance
})

var JiraClient = sync.OnceValue[*jira.Client](func() *jira.Client {
	var (
		//Host is the URL of the Jira instance
		host  = os.Getenv("ATLASSIAN_HOST")
		//Mail is the email of the user
		mail  = os.Getenv("ATLASSIAN_EMAIL")
		//Token is the API token of the user
		token = os.Getenv("ATLASSIAN_TOKEN")
	)

	if host == "" || mail == "" || token == "" {
		log.Fatal("ATLASSIAN_HOST, ATLASSIAN_EMAIL, ATLASSIAN_TOKEN are required")
	}

	instance, err := jira.New(nil, host)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "failed to create jira client"))
	}

	instance.Auth.SetBasicAuth(mail, token)

	return instance
})
