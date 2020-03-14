# Knowbody
Simple RSS -> Slack bridge for Otakushirts only... for now.

If you know what this is for, feel free to PR changes in to conf.yaml

If you're running `make run` you need a `.env` file that contains:

```
export SLACK_TOKEN=xxx
```

If you're running it any other way, it will expect the `SLACK_TOKEN`
environment variable to exist and, obviously, contain a valid Slack
token.