package main

import (
	"log"
	"os"
	"github.com/fsnotify"
	"context"
	"io"
	"net/http"
	"path/filepath"
	"cloud.google.com/go/storage"
)

func GetFileContentType(out *os.File) (string, error){
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer)

	return contentType, nil
}
//
func write(client *storage.Client, bucket, object string) error {
	ctx := context.Background()
	// [START upload_file]
	//path = filepath.FromSlash(object)
	f, err := os.Open(object)
	if err != nil {
		return err
	}
	defer f.Close()

	_, path := filepath.Split(object)

	contentType, err := GetFileContentType(f)
	if err != nil{
		panic(err)
	}
	if contentType == "image/jpeg" || contentType == "image/png"{
		wc := client.Bucket(bucket).Object(path).NewWriter(ctx)
		if _, err = io.Copy(wc, f); err != nil {
			return err
		}
		if err := wc.Close(); err != nil {
			return err
		}
		log.Println("Uploaded to Cloud")
	}else{
		return nil
	}
	// [END upload_file]
	return nil
}

func main() {
	ctx := context.Background()

	//Sets your Google Cloud Platform
	//projectID := "My practice project"

	//Creates a client
	client, err := storage.NewClient(ctx)
	if err != nil{
		log.Fatalf("Failed to create client: %v", err)
	}
	bucket := "staging.my-practice-project-217021.appspot.com"
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event.Name)
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("New File Detected:", event.Name)
					//Implement syncing to google cloud
					if err:= write(client, bucket, event.Name); err != nil{
						log.Fatal(err)
					}
					log.Println("Uploaded on Cloud")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("C://Users//Joon//Documents//XponentialWorks//Watched")
	log.Println("Directory is being watched for new images...")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}