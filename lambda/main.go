package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dk1027/go-questrade-api/controlflow"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-lambda-go/lambda"
)

var region = "us-west-2"
var bucket = "dk1027-go-questrade"
var s3_prefix = "dk1027"

type MyEvent struct {
	ConfigPath string `json:"config_path"`
}

func HandleRequest(_ context.Context, _ MyEvent) (string, error) {
	log.Println("starting lambda")
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	downloader := s3manager.NewDownloader(sess)
	buff := &aws.WriteAtBuffer{}
	key := fmt.Sprintf("%s/%s", s3_prefix, "config.yaml")
	log.Printf("Downloading %s", key)
	_, err := downloader.Download(buff,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		log.Fatalf("failed to download config file: %v", err)
	}

	cf := controlflow.Parse(buff.Bytes())
	cf.Execute()
	return "", nil
}

func main() {
	lambda.Start(HandleRequest)
}
