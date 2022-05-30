package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/tabwriter"
	"time"
)

var configFile = "config/config.json"
var logsDir = "logs"

// input config file
type Config struct {
	Timings   Timings    `json:"timings"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Timings struct {
	IntervalSeconds  int `json:"intervalSeconds"`
	RunDurationHours int `json:"runDurationHours"`
}

type Endpoint struct {
	Method    string    `json:"method"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	BasicAuth BasicAuth `json:"basicAuth"`
	Headers   []Header  `json:"headers"`
}

type BasicAuth struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// output to save / display
type LatencyMeasure struct {
	Name       string
	StatusCode int
	TimeTaken  int64
}

//functions
func main() {

	//read and load config for interval, run duration and endpoints
	var config Config
	fmt.Printf("Reading config from %s\n", configFile)
	jsonFile, err := os.Open(configFile)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Unable to read config file in current directory. Exiting!")
		return
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &config)

	//create log file for current app instance
	logFile := fmt.Sprintf("%s/latencies-%d-%s-PID_%d.log", logsDir, time.Now().Day(), time.Now().Month(), os.Getpid())
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Unable to create log file in current directory. Exiting!")
		return
	}
	defer f.Close()
	log.SetOutput(f)
	fmt.Printf("Logging to file - %s\n", logFile)

	//verify if atleast 1 endpoint was read, timings are present and print list of endpoints
	if len(config.Endpoints) < 1 && (config.Timings.IntervalSeconds == 0 || config.Timings.RunDurationHours == 0) {
		printToAllLoggers(fmt.Sprintf("Unable to parse config file or no endpoints present. Please check %s. Exiting!\n", configFile))
		return
	}
	for i, endpoint := range config.Endpoints {
		printToAllLoggers(fmt.Sprintf("Endpoint %d \"%s\"\t %s", i+1, endpoint.Name, endpoint.URL))
	}
	printToAllLoggers(fmt.Sprintf("Using config as %d seconds interval and %d hour run duration", config.Timings.IntervalSeconds, config.Timings.RunDurationHours))
	printToAllLoggers("Starting up ... ")
	writer := tabwriter.NewWriter(os.Stdout, 2, 8, 1, '\t', tabwriter.AlignRight)

	//configure the ticker
	ticker := time.NewTicker(time.Duration(config.Timings.IntervalSeconds) * time.Second)
	stopTicker := make(chan bool)

	//start goroutine
	go func() {
		for {
			select {
			case <-stopTicker:
				return
			case <-ticker.C:
				{
					fmt.Fprintln(writer, "")
					writer.Flush()
					for i, endpoint := range config.Endpoints {
						go fmt.Fprintln(writer, measureURL(i, endpoint))
					}
				}
			}
		}
	}()

	printToAllLoggers(fmt.Sprintf("Started up %d pollers ...", len(config.Endpoints)))

	//configure the ticker stop
	time.Sleep(time.Duration(config.Timings.RunDurationHours) * time.Hour)
	ticker.Stop()
	stopTicker <- true

	printToAllLoggers("Shutting down ... ")
}

func printToAllLoggers(message string) {
	log.Printf(message)
	fmt.Println(message)
}

func measureURL(index int, endpoint Endpoint) string {
	var resultDisplayText string
	var resultLoggerText string
	var latencyEntry LatencyMeasure
	latencyEntry.Name = endpoint.Name

	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	req, err := http.NewRequest(endpoint.Method, endpoint.URL, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	if endpoint.BasicAuth.UserName != "" || endpoint.BasicAuth.Password != "" {
		req.SetBasicAuth(endpoint.BasicAuth.UserName, endpoint.BasicAuth.Password)
	}

	for _, header := range endpoint.Headers {
		req.Header.Add(header.Name, header.Value)
	}

	startTimestamp := time.Now()
	response, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer response.Body.Close()
	latencyEntry.TimeTaken = time.Now().Sub(startTimestamp).Milliseconds()

	if err != nil {
		fmt.Println(err.Error())
		latencyEntry.StatusCode = 0

		resultLoggerText = fmt.Sprintf("System: %s, ConnectionError, Time Taken : %d ms", latencyEntry.Name, latencyEntry.TimeTaken)
		if latencyEntry.TimeTaken < 1000 {
			resultDisplayText = fmt.Sprintf("System: %s \t ConnectionError \t %d ms", latencyEntry.Name, latencyEntry.TimeTaken)
		} else {
			resultDisplayText = fmt.Sprintf("System: %s \t ConnectionError \t %.2f seconds", latencyEntry.Name, float64(latencyEntry.TimeTaken)/1000)
		}
	} else {
		latencyEntry.StatusCode = response.StatusCode

		resultLoggerText = fmt.Sprintf("System: %s, HTTP Status %d, Time Taken : %d ms", latencyEntry.Name, latencyEntry.StatusCode, latencyEntry.TimeTaken)
		if latencyEntry.TimeTaken < 1000 {
			resultDisplayText = fmt.Sprintf("System: %s \t HTTP Status %d \t %d ms", latencyEntry.Name, latencyEntry.StatusCode, latencyEntry.TimeTaken)
		} else {
			resultDisplayText = fmt.Sprintf("System: %s \t HTTP Status %d \t %.2f seconds", latencyEntry.Name, latencyEntry.StatusCode, float64(latencyEntry.TimeTaken)/1000)
		}
	}

	log.Println(resultLoggerText)
	return resultDisplayText
}
