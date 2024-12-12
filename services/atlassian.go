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
		host  = os.Getenv("ATLASSIAN_HOST")
		mail  = os.Getenv("ATLASSIAN_EMAIL")
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
		host  = os.Getenv("ATLASSIAN_HOST")
		mail  = os.Getenv("ATLASSIAN_EMAIL")
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
