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
)

type Changer struct {
	EmojiPattern string `short:"p" long:"emoji-pattern" description:"Add an emoji pattern after the new name"`
	Continuously int64  `short:"c" long:"continuously" description:"Check and change name every N minutes" default:"0"`
	OldName      string `short:"o" long:"old" description:"Old name to match on."`
	First        string `short:"f" long:"first" description:"New first name to use" required:"true"`
	Last         string `short:"l" long:"last" description:"New last name to use" required:"true"`
	Token        string `short:"t" long:"token" description:"Slack Token" required:"true" env:"SLACK_TOKEN"`
}

var emojis = []string{
	"☁️",
	"💥",
	"🎉",
	"😎",
}

var changer Changer

func main() {
	rand.Seed(time.Now().Unix())

	_, err := flags.Parse(&changer)
	if err != nil {
		return
	}

	// TODO: figure out how to allow emojis to be specified from CLI args
	// fmt.Println(changer.Emojis)
	// fmt.Println(len(changer.Emojis))
	// for ind, val := range changer.Emojis {
	// 	fmt.Printf("%d: %s\n", ind, val)
	// }
	// for i, w := 0, 0; i < len(changer.Emojis); i += w {
	// 	runeValue, width := utf8.DecodeRuneInString(changer.Emojis[i:])
	// 	fmt.Printf("%#U starts at byte position %d\n", runeValue, i)
	// 	w = width
	// }

	for {
		err = changer.Process()
		if err != nil {
			fmt.Printf("error: %s\n", err)
			if changer.Continuously == 0 {
				break
			}
		}

		if changer.Continuously == 0 {
			break
		} else {
			fmt.Printf("Sleeping for %d minutes...\n", changer.Continuously)
			time.Sleep(time.Duration(changer.Continuously) * time.Minute)
		}
	}
}

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
		newLast := fmt.Sprintf("%s%s", c.Last, c.Emoji())

		fmt.Printf("Changing name to %s %s\n", c.First, newLast)

		_, err := c.post("users.profile.set", url.Values{"name": {"first_name"}, "value": {c.First}})
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

func (c Changer) Emoji() string {
	switch c.EmojiPattern {
	case "single":
		return fmt.Sprintf(" %s", emojis[rand.Intn(len(emojis))])
	case "3pal":
		in := rand.Intn(len(emojis))
		in2 := rand.Intn(len(emojis))
		return fmt.Sprintf(" %s %s %s", emojis[in], emojis[in2], emojis[in])
	default:
		return ""
	}
}

func (c Changer) post(method string, params url.Values) (*jsontree.JsonTree, error) {
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
