package knowbody

import (
	"log"
	"regexp"
	"time"

	"github.com/araddon/dateparse"
	"github.com/mmcdole/gofeed"
	"github.com/nlopes/slack"
)

// Config holds all info to run Knowbody
type Config struct {
	Streams    []ContentStream `yaml:"streams"`
	SlackToken string          `yaml:"slackToken"`
}

// ContentStream is a mapping of rss-feed to slack channel
type ContentStream struct {
	Name    string `yaml:"name"`    // name of the stream. preferably unique.
	URL     string `yaml:"url"`     // rss feed (https://twitrss.me/twitter_user_to_rss/?user=jeefy or https://danielmiessler.com/blog/rss-feed-youtube-channel/)
	Channel string `yaml:"channel"` // slack channel in the workspace
	Exclude string `yaml:"exclude"` // regex of content to exclude from this source cannot be set with include.
	Include string `yaml:"include"` // regex of content to include from this source. cannot be set with exclude.
	Spoiler bool   `yaml:"spoiler"` // boolean to indicate if it should post in a thread instead of in channel.

	excludeRegex *regexp.Regexp
	includeRegex *regexp.Regexp
}

// ContentState tracks the current state of a specific ContentStream
type ContentState struct {
	Stream  ContentStream `yaml:"stream"`
	RSSId   string        `yaml:"rssId"`
	RSSTime time.Time     `yaml:"rssTime"`
}

// CurrentState tracks the overall running state of Knowbody
type CurrentState struct {
	Streams  map[string]ContentState `yaml:"streams"`
	Channels map[string]string       `yaml:"channels"`
	LastRun  time.Time               `yaml:"lastRun"`

	slackClient *slack.Client
}

// Process will attempt to read and handle a specific ContentStream's RSS feed
func (stream *ContentStream) Process() {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(stream.URL)

	if err != nil {
		log.Printf("Error getting feed %s: %s", stream.URL, err.Error())
	} else {
		log.Printf("Looping through feed from %s", stream.URL)
		if feed != nil {
			for i := len(feed.Items) - 1; i >= 0; i-- {
				item := feed.Items[i]

				streamName := stream.Name
				_, ok := State.Streams[streamName]
				if !ok {
					State.Streams[streamName] = ContentState{
						Stream:  *stream,
						RSSId:   "0",
						RSSTime: State.LastRun,
					}
				}

				parsedDate, err := dateparse.ParseAny(item.Published)
				if err != nil {
					log.Fatalf("Error parsing date: %s", err.Error())
				}

				if parsedDate.After(State.Streams[streamName].RSSTime) {
					log.Printf("Checking %s", item.Title)
					if (stream.Exclude == "" && stream.Include == "") || (stream.Exclude != "" && !stream.excludeRegex.Match([]byte(item.Title))) || (stream.Include != "" && stream.includeRegex.Match([]byte(item.Title))) {
						log.Printf("Posting %s: %s to %s", item.Title, item.Link, stream.Channel)

						if _, ok := State.Channels[stream.Channel]; !ok {
							log.Printf("Channel '%s' does not exist on slack server.", stream.Channel)
						} else {
							if stream.Spoiler == true {
								_, ts, postErr := State.slackClient.PostMessage(State.Channels[stream.Channel], slack.MsgOptionText(item.Title, false))
								if postErr != nil {
									log.Printf("Error posting to spoiler reply thread in: %s", postErr.Error())
								}
								State.slackClient.PostMessage(State.Channels[stream.Channel], slack.MsgOptionText(item.Link, false), slack.MsgOptionTS(ts))
							} else {
								State.slackClient.PostMessage(State.Channels[stream.Channel], slack.MsgOptionText(item.Link, false))
							}
						}
					}
					state := State.Streams[streamName]
					state.RSSId = item.GUID
					state.RSSTime = parsedDate
					State.Streams[streamName] = state
				}
			}
		}
	}
}
