package knowbody

import (
	"fmt"
	"log"
	"os"

	"github.com/araddon/dateparse"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// twitter processes twitter-specific feeds and generates a Message
func (stream *ContentStream) twitter(c chan Message) {
	twitterConsumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	twitterConsumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	twitterAccessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	twitterAccessSecret := os.Getenv("TWITTER_ACCESS_SECRET")
	config := oauth1.NewConfig(twitterConsumerKey, twitterConsumerSecret)
	token := oauth1.NewToken(twitterAccessToken, twitterAccessSecret)
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	Twitter := twitter.NewClient(httpClient)

	/*feedURL := fmt.Sprintf("%s&t=%v", stream.Source, time.Now().Unix())

	feed, err := fp.ParseURL(feedURL)*/

	tweets, _, err := Twitter.Timelines.UserTimeline(&twitter.UserTimelineParams{
		ScreenName: stream.Source,
	})

	if err != nil {
		log.Printf("Error getting tweets from %s: %s", stream.Source, err.Error())
	} else {
		log.Printf("Looping through tweets from %s", stream.Source)
		if len(tweets) > 0 {
			for i := len(tweets) - 1; i >= 0; i-- {
				item := tweets[i]

				streamName := stream.Name
				_, ok := State.Streams[streamName]
				if !ok {
					State.Streams[streamName] = ContentState{
						Stream:  *stream,
						RSSId:   "0",
						RSSTime: State.LastRun,
					}
				}

				parsedDate, err := dateparse.ParseAny(item.CreatedAt)
				if err != nil {
					log.Fatalf("Error parsing tweet date: %s", err.Error())
				}

				if parsedDate.After(State.Streams[streamName].RSSTime) {
					log.Printf("Checking %s", item.FullText)
					if (stream.Exclude == "" && stream.Include == "") || (stream.Exclude != "" && !stream.excludeRegex.Match([]byte(item.FullText))) || (stream.Include != "" && stream.includeRegex.Match([]byte(item.FullText))) || (item.RetweetedStatus != nil && ((stream.Exclude != "" && !stream.excludeRegex.Match([]byte(item.RetweetedStatus.Text))) || (stream.Include != "" && stream.includeRegex.Match([]byte(item.RetweetedStatus.Text))))) {
						link := fmt.Sprintf("https://twitter.com/%v/status/%v", stream.Source, item.IDStr)
						log.Printf("Posting %s: %s to %s", item.FullText, link, stream.Channel)

						c <- Message{
							Title:   item.FullText,
							Link:    link,
							Channel: stream.Channel,
							Spoiler: stream.Spoiler,
						}
					}
					state := State.Streams[streamName]
					state.RSSId = item.IDStr
					state.RSSTime = parsedDate
					State.Streams[streamName] = state
				}
			}
		}
	}
}
