package simpleslack

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	jsontree "github.com/bmatsuo/go-jsontree"
)

// Client holds the information needed to make API calls to Slack.  Right now,
// that is just the token.
type Client struct {
	Token string
}

// Post will send a POST request to the Slack API method with the specified
// form values and then return the result in a go-jsontree object, for easy
// introspection.
func (c Client) Post(method string, params url.Values) (*jsontree.JsonTree, error) {
	params["token"] = []string{c.Token}
	resp, err := http.PostForm(fmt.Sprintf("https://slack.com/api/%s", method), params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	tree := jsontree.New()
	err = tree.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}

	ok, err := tree.Get("ok").Boolean()
	if err != nil {
		return nil, err
	}

	if !ok {
		message, _ := tree.Get("error").String()
		return nil, fmt.Errorf("Error: %s", message)
	}

	return tree, nil
}
