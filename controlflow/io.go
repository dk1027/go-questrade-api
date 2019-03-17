package controlflow

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/aws"

	sess "github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type IOProvider interface {
	Write(data interface{}, filename string) error
	ReadPortfolio(filename string) (*Portfolio, error)
	Read(filename string, out interface{}) error
}

type FileIO struct {
}

func (*FileIO) Write(data interface{}, filename string) error {
	log.Printf("writing %s\n", filename)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("failed marshaling data")
	}
	err = ioutil.WriteFile(filename, jsonBytes, 0644)
	if err != nil {
		log.Fatalf("unable to write data file")
	}
	return nil
}

func (*FileIO) ReadPortfolio(filename string) (*Portfolio, error) {
	portfolio := &Portfolio{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to open portfolio file: %v", err)
	}
	if err = json.Unmarshal(data, &portfolio); err != nil {
		log.Fatalf("unable to unmarshal portfolio")
	}
	return portfolio, nil
}

func (*FileIO) ReadJson(filename string) (*map[string]interface{}, error) {
	out := &map[string]interface{}{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to open json file: %v", err)
	}

	if err = json.Unmarshal(data, &out); err != nil {
		log.Fatalf("unable to unmarshal json")
	}
	return out, nil
}

func (*FileIO) Read(filename string, out interface{}) error {
	jsonBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln(err)
	}
	b64 := string(jsonBytes)
	b64 = b64[1 : len(b64)-1]
	log.Print(b64)
	jsonString, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(jsonString, &out); err != nil {
		log.Fatal("unable to unmarshal json")
	}
	return nil
}

type S3IO struct {
	Region     string
	BucketName string
	Prefix     string
	session    *sess.Session
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

func NewS3IO(region, bucketName, prefix string) *S3IO {
	s := sess.Must(sess.NewSession(&aws.Config{Region: aws.String(region)}))
	return &S3IO{
		Region:     region,
		BucketName: bucketName,
		Prefix:     prefix,
		session:    s,
		uploader:   s3manager.NewUploader(s),
		downloader: s3manager.NewDownloader(s),
	}
}

func (io *S3IO) Write(data interface{}, filename string) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("failed marshaling data")
	}
	key := fmt.Sprintf("%s/%s", io.Prefix, filename)
	res, err := io.uploader.Upload(&s3manager.UploadInput{
		Bucket: &io.BucketName,
		Key:    &key,
		Body:   bytes.NewReader(jsonBytes),
	})

	if err != nil {
		log.Fatalf("unable to upload data file: %v", err)
	}
	log.Printf("upload successfully %s", res.Location)
	return nil
}

func (io *S3IO) S3Download(thing interface{}, filename string) error {
	buff := &aws.WriteAtBuffer{}
	key := fmt.Sprintf("%s/%s", io.Prefix, filename)
	_, err := io.downloader.Download(buff,
		&s3.GetObjectInput{
			Bucket: aws.String(io.BucketName),
			Key:    aws.String(key),
		})
	if err != nil {
		log.Fatalf("failed to download portfolio file: %v", err)
	}
	b64 := string(buff.Bytes())
	b64 = b64[1 : len(b64)-1]
	jsonBytes, err := base64.StdEncoding.DecodeString(b64)
	if err = json.Unmarshal(jsonBytes, thing); err != nil {
		log.Fatalf("unable to unmarshal data")
	}
	return nil
}

func (io *S3IO) ReadPortfolio(filename string) (*Portfolio, error) {
	portfolio := &Portfolio{}
	err := io.S3Download(portfolio, filename)
	return portfolio, err
}

func (io *S3IO) Read(filename string, out interface{}) error {
	err := io.S3Download(out, filename)
	return err
}
