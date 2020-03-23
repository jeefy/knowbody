# Knowbody
"Knowledge Event" -> Slack bridge for Otakushirts only... for now.

tl;dr - Knowbody hits periodically hits an endpoint, and posts a link to the event/post to a designated channel.

If you know what this is for, feel free to PR changes in to conf.yaml

## Config Options
`SLACK_TOKEN` - Valid Slack Bot token that can post to channels

`TWITTER_CONSUMER_KEY`

`TWITTER_CONSUMER_SECRET`

`TWITTER_ACCESS_TOKEN`

`TWITTER_ACCESS_SECRET`

`SKIP_CONF_UPDATE` - Prevents knowbody from grabbing the latest config from GitHub. Useful for local testing.