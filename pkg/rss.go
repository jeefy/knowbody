package knowbody

import (
	"fmt"
	"log"
	"time"

	"github.com/araddon/dateparse"
	"github.com/mmcdole/gofeed"
)

// rss processes RSS type streams and generates a message
func (stream *ContentStream) rss(c chan Message) {
	fp := gofeed.NewParser()

	feedURL := fmt.Sprintf("%s&t=%v", stream.Source, time.Now().Unix())

	feed, err := fp.ParseURL(feedURL)

	if err != nil {
		log.Printf("Error getting feed %s: %s", feedURL, err.Error())
	} else {
		log.Printf("Looping through feed from %s", feedURL)
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

						c <- Message{
							Title:   item.Title,
							Link:    item.Link,
							Channel: stream.Channel,
							Spoiler: stream.Spoiler,
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
