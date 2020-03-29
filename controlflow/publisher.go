package controlflow

import (
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"

	sess "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type Report struct {
	Aggregtae        *Table
	Gap              *Table
	PercentPortfolio *Table
}

type Publisher interface {
	Publish(report *Report) error
}

type SNSPublisher struct {
	topicArn string
	region   string
	sns      *sns.SNS
}

func NewSNSPublisher(region, topicArn string) *SNSPublisher {
	svc := sns.New(sess.Must(sess.NewSession(&aws.Config{Region: aws.String(region)})))
	return &SNSPublisher{
		topicArn: topicArn,
		region:   region,
		sns:      svc,
	}
}

func (p *SNSPublisher) Publish(report *Report) error {
	var headers []string
	for k := range *report.Aggregtae {
		headers = append(headers, k)
	}
	sort.Strings(headers)
	s := ToText(headers, []Table{*report.Aggregtae, *report.Gap, *report.PercentPortfolio})
	log.Println(s)
	input := &sns.PublishInput{}
	input.SetTopicArn(p.topicArn)
	input.SetMessage(s)
	log.Println(input)
	output, err := p.sns.Publish(input)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(output)
	return err
}

type NullPublisher struct{}

func (n *NullPublisher) Publish(report *Report) error {
	var headers []string
	for k := range *report.Aggregtae {
		headers = append(headers, k)
	}
	s := ToText(headers, []Table{*report.Aggregtae, *report.Gap, *report.PercentPortfolio})
	log.Println(s)
	return nil
}
