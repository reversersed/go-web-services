package rest

import (
	"errors"
	"net/http"
	"testing"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

var urlBilderCases = []struct {
	Name     string
	Url      string
	Path     string
	Filters  []FilterOptions
	Err      error
	Excepted string
}{
	{
		Name:     "Empty url test",
		Url:      "http://localhost:0000",
		Path:     "",
		Filters:  []FilterOptions{},
		Excepted: "http://localhost:0000",
	},
	{
		Name:     "Empty filter test",
		Url:      "http://localhost:0000",
		Path:     "/testing",
		Filters:  []FilterOptions{},
		Excepted: "http://localhost:0000/testing",
	},
	{
		Name:     "Empty filter test without slash",
		Url:      "http://localhost:0000",
		Path:     "testing",
		Filters:  []FilterOptions{},
		Excepted: "http://localhost:0000/testing",
	},
	{
		Name: "Single filter test",
		Url:  "http://localhost:0000",
		Path: "/testing",
		Filters: []FilterOptions{
			{
				Field:  "id",
				Values: []string{"test"},
			},
		},
		Excepted: "http://localhost:0000/testing?id=test",
	},
	{
		Name: "Single filter test with multiple values",
		Url:  "http://localhost:0000",
		Path: "/testing",
		Filters: []FilterOptions{
			{
				Field:  "id",
				Values: []string{"test", "second", "any"},
			},
		},
		Excepted: "http://localhost:0000/testing?id=test%2Csecond%2Cany",
	},
	{
		Name: "Multiple filter test with multiple values",
		Url:  "http://localhost:0000",
		Path: "/testing",
		Filters: []FilterOptions{
			{
				Field:  "id",
				Values: []string{"test", "second", "any"},
			},
			{
				Field:  "name",
				Values: []string{"Alice", "Gray"},
			},
		},
		Excepted: "http://localhost:0000/testing?id=test%2Csecond%2Cany&name=Alice%2CGray",
	},
	{
		Name:    "Wrong http url",
		Url:     "wrongurl",
		Path:    "testing",
		Filters: []FilterOptions{},
		Err:     errors.New("failed to parse url: parse \"wrongurl\": invalid URI for request"),
	},
}

func TestUrlBuilder(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}

	for _, urlCase := range urlBilderCases {
		t.Run(urlCase.Name, func(t *testing.T) {
			client := &RestClient{
				BaseURL: urlCase.Url,
				Logger:  logger,
			}
			url, err := client.BuildURL(urlCase.Path, urlCase.Filters)
			if err != nil {
				if urlCase.Err != nil {
					if urlCase.Err.Error() != err.Error() {
						t.Fatalf("excepted error %s but got %s", urlCase.Err.Error(), err.Error())
					}
				} else {
					t.Fatalf("excepted url but got error %v", err)
				}
			}
			if urlCase.Err != nil && err == nil {
				t.Fatalf("excepted error %s but got nil", urlCase.Err.Error())
			}
			if url != urlCase.Excepted && urlCase.Err == nil {
				t.Errorf("excepted %s but got %s", urlCase.Excepted, url)
			}
		})
	}
}

func TestClientClose(t *testing.T) {
	client := &RestClient{
		HttpClient: &http.Client{},
	}
	client.Close()

	if client.HttpClient != nil {
		t.Errorf("excepted http nil but got %v", client.HttpClient)
	}
}
