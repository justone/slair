# slair (Slack + Flair)

This is a small program that helps me keep my Slack profile name in line.

# Install

```
go get github.com/justone/slair
```

# Example usage

Set Slack token via environment variable:

```
export SLACK_TOKEN=xoxp-zzzzzzzzzz-zzzzzzzzzz-zzzzzzzzzzz-zzzzzzzzzz
```

Alternatively, the slack token can be passed as an argument (`-t`).

Change your profile name (supports utf8 emoji):

```
slair -f Jim -l Bob
slair -f Jim -l "Bob ☁️"
```

Set name only if a certain old name is found:

```
slair -f Jim -l Bob -o James
```

Continuously correct the name every N minutes:

```
slair -f Jim -l Bob -o James -c 15
```

Append some flair at the end:

```
slair -f Jim -l Bob -c 15 -p single
slair -f Jim -l Bob -o James -c 15 -p 3pal
```

Possible values for `-p` are:

* 'single': a single emoji picked from a list
* '3pal': three emojis in a palindrome (e.g. 💥☁️💥)
