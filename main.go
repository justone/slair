package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	jsontree "github.com/bmatsuo/go-jsontree"
	flags "github.com/jessevdk/go-flags"
	"gopkg.in/kyokomi/emoji.v1"
)

var opts struct {
	EmojiPattern string `short:"p" long:"emoji-pattern" description:"Add an emoji pattern after the new name"`
	Emojis       string `short:"e" long:"emojis" description:"List of emojis to use when randomly selecting" default:":cloud:,:boom:,:tada:,:sunglasses:"`
	ListEmojis   bool   `long:"list-emojis" description:"List all emojis"`
	Continuously int64  `short:"c" long:"continuously" description:"Check and change name every N minutes" default:"0"`
	OldName      string `short:"o" long:"old" description:"Old name to match on."`
	First        string `short:"f" long:"first" description:"New first name to use"`
	Last         string `short:"l" long:"last" description:"New last name to use"`
	User         string `short:"u" long:"user" description:"User ID to query/update (optional, defaults to current user)"`
	LookupUser   string `long:"user-lookup" description:"Look up User ID by username."`
	Token        string `short:"t" long:"token" description:"Slack Token" env:"SLACK_TOKEN"`
}

// Changer handles everything around changing the Slack profile.
type Changer struct {
	Emojis                                          []string
	User, OldName, First, Last, Token, EmojiPattern string
}

// ParseEmojis takes a comma-delimited list of emoji identifiers and returns an
// array of real emoji characters.
func ParseEmojis(list string) ([]string, error) {
	result := []string{}

	rawList := strings.Split(list, ",")
	for _, raw := range rawList {
		raw = strings.TrimSpace(raw)
		if raw[0:1] != ":" {
			raw = fmt.Sprintf(":%s:", raw)
		}
		if _, ok := emoji.CodeMap()[raw]; !ok {
			return []string{}, fmt.Errorf("Unable to resolve emoji for %s", raw)
		}
		result = append(result, emoji.Sprint(raw))
	}

	return result, nil
}

func main() {
	rand.Seed(time.Now().Unix())

	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}

	if opts.ListEmojis {
		for k := range emoji.CodeMap() {
			fmt.Printf("%s => %s\n", k, emoji.Sprint(k))
		}
		return
	}

	if len(opts.Token) == 0 {
		fmt.Println("error: no slack token found, use either env var or cli arg")
		return
	}

	emojis, err := ParseEmojis(opts.Emojis)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	changer := Changer{
		emojis,
		opts.User,
		opts.OldName,
		opts.First,
		opts.Last,
		opts.Token,
		opts.EmojiPattern,
	}

	if len(opts.LookupUser) > 0 {
		userID, err := lookupUserID(changer, opts.LookupUser)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("User id:", userID)
		return
	}

	for {
		err = changer.Process()
		if err != nil {
			fmt.Printf("error: %s\n", err)
			if opts.Continuously == 0 {
				break
			}
		}

		if opts.Continuously == 0 {
			break
		} else {
			fmt.Printf("Sleeping for %d minutes...\n", opts.Continuously)
			time.Sleep(time.Duration(opts.Continuously) * time.Minute)
		}
	}
}

// lookupUserID will return the user id for the given user.
func lookupUserID(c Changer, lookup string) (string, error) {
	userList, err := c.post("users.list", url.Values{})
	if err != nil {
		return "", err
	}

	users := userList.Get("members")
	count, err := users.Len()
	if err != nil {
		return "", err
	}

	for i := 0; i < count; i++ {
		user := users.GetIndex(i)
		username, err := user.Get("name").String()
		if err != nil {
			return "", err
		}

		if username == opts.LookupUser {
			id, err := user.Get("id").String()
			if err != nil {
				return "", err
			}
			return id, nil
		}
	}

	return "", fmt.Errorf("Not found")
}

// Process takes care of updating the Slack profile.  If an old name is
// specified, it must be present in the old profile name for the update to
// occur.  If EmojiPattern is present, then flair will be added.
func (c Changer) Process() error {

	profile, err := c.post("users.profile.get", url.Values{})

	first, err := profile.Get("profile").Get("first_name").String()
	if err != nil {
		return err
	}
	last, err := profile.Get("profile").Get("last_name").String()
	if err != nil {
		return err
	}

	fmt.Printf("Name found: %s %s\n", first, last)
	if len(c.OldName) == 0 || (strings.Contains(first, c.OldName) || strings.Contains(last, c.OldName)) {
		newLast := fmt.Sprintf("%s%s", emoji.Sprint(c.Last), c.Flair())

		fmt.Printf("Changing name to %s %s\n", c.First, newLast)

		_, err := c.post("users.profile.set", url.Values{"name": {"first_name"}, "value": {emoji.Sprint(c.First)}})
		if err != nil {
			return err
		}
		_, err = c.post("users.profile.set", url.Values{"name": {"last_name"}, "value": {newLast}})
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Skipping name change")
	}

	return nil
}

// Flair generates the pieces of flair.
func (c Changer) Flair() string {
	switch c.EmojiPattern {
	case "single":
		return fmt.Sprintf(" %s", c.Emojis[rand.Intn(len(c.Emojis))])
	case "3pal":
		in := rand.Intn(len(c.Emojis))
		in2 := rand.Intn(len(c.Emojis))
		return fmt.Sprintf(" %s %s %s", c.Emojis[in], c.Emojis[in2], c.Emojis[in])
	default:
		return ""
	}
}

// post makes the actual requests to the Slack API.
func (c Changer) post(method string, params url.Values) (*jsontree.JsonTree, error) {
	params["token"] = []string{c.Token}
	if len(c.User) > 0 {
		params["user"] = []string{c.User}
	}
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
