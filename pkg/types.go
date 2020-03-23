package knowbody

import (
	"log"
	"regexp"
	"time"

	"github.com/nlopes/slack"
)

// Config holds all info to run Knowbody
type Config struct {
	Streams    []ContentStream `yaml:"streams"`
	SlackToken string          `yaml:"slackToken"`
}

// Message represents slack messages to send
type Message struct {
	Channel string
	Link    string
	Title   string
	Spoiler bool
}

// ContentStream is a mapping of rss-feed to slack channel
type ContentStream struct {
	Name    string `yaml:"name"`   // name of the stream. preferably unique.
	Source  string `yaml:"source"` // rss feed (https://twitrss.me/twitter_user_to_rss/?user=jeefy or https://danielmiessler.com/blog/rss-feed-youtube-channel/)
	Type    string `yaml:"type"`
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
func (stream *ContentStream) Process(c chan Message) {
	switch stream.Type {
	case "rss":
		stream.rss(c)
	case "twitter":
		stream.twitter(c)
	default:
		log.Printf("invalid stream type: %s", stream.Type)
	}
}
