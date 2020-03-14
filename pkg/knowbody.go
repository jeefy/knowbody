package knowbody

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/nlopes/slack"
	"gopkg.in/yaml.v2"
)

var (
	// CurrentConfig is what Knowbody currently sees as the config
	// Think "State" but for the config
	CurrentConfig Config
	// State is... the current state.
	State CurrentState
)

// Start starts the loop to look at all the things
func Start() {
	CurrentConfig.SlackToken = os.Getenv("SLACK_TOKEN")

	State.Streams = make(map[string]ContentState)
	State.Channels = make(map[string]string)

	// Recover state from previous run
	ReadState()

	// Assume if LastRun is over a year old to not bother and just set it to now
	// Prevents spamming
	if State.LastRun.Before(time.Now().AddDate(-1, 0, 0)) {
		State.LastRun = time.Now()
	}

	for {
		// Allow config changes between runs
		ReadConfig()

		State.slackClient = slack.New(CurrentConfig.SlackToken)

		channels, err := State.slackClient.GetChannels(true)
		if err != nil {
			log.Fatalf("error getting slack channels: %s", err)
		}

		for _, channel := range channels {
			State.Channels[channel.Name] = channel.ID
		}

		for _, contentStream := range CurrentConfig.Streams {
			contentStream.Process()
		}

		State.LastRun = time.Now()

		WriteState()

		log.Print("Time to sleep for 60")

		time.Sleep(60 * time.Second)
	}
}

// Lint simply reads the two primary files and ensures they can be parsed.
func Lint() {
	readYamlIntoConfig("conf.yaml", &CurrentConfig)
	readYamlIntoConfig("/tmp/knowbody.lock", &State)
}

// ReadConfig will attempt to download the master config from GitHub.
// It then reads the current conf.yaml file on the file system and updates
// all the internal data structures
func ReadConfig() {
	err := DownloadFile("conf.yaml", "https://raw.githubusercontent.com/jeefy/knowbody/master/conf.yaml")
	if err != nil {
		log.Printf("Error downloading updated config from Github: %s", err.Error())
	}

	readYamlIntoConfig("conf.yaml", &CurrentConfig)

	for key, stream := range CurrentConfig.Streams {
		comp, err := regexp.Compile(stream.Include)
		if err != nil {
			log.Fatalf("Error compiling regex '%s': %s", stream.Include, err.Error())
		}
		CurrentConfig.Streams[key].includeRegex = comp

		comp, err = regexp.Compile(stream.Exclude)
		if err != nil {
			log.Fatalf("Error compiling regex `%s`: %s", stream.Exclude, err.Error())
		}
		CurrentConfig.Streams[key].excludeRegex = comp
	}
}

// WriteState takes the state and dumps it to /tmp/knowbody.lock
// We like keeping track of state. This allows Knowbody to crash and recover without issue.
// Because we're lazy goddammit.
func WriteState() {
	d, err := yaml.Marshal(&State)
	if err != nil {
		log.Fatalf("Error marshalling YAML: %s", err.Error())
	}

	err = ioutil.WriteFile("/tmp/knowbody.lock", d, 0644)
	if err != nil {
		log.Fatalf("Error writing state file: %s", err.Error())
	}
}

// ReadState looks reads /tmp/knowbody.lock and pushes it into the State variable
func ReadState() {
	yamlFile, err := ioutil.ReadFile("/tmp/knowbody.lock")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &State)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}

// DownloadFile downloads files. Pretty simple.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error downloading file: Status Code %d", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func readYamlIntoConfig(file string, obj interface{}) {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, obj)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}
