package controlflow

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/dk1027/go-questrade-api/api"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

type S3Config struct {
	Region *string `yaml:"region" validate:"required"`
	Bucket *string `yaml:"bucket" validate:"required"`
	Prefix *string `yaml:"prefix" validate:"required"`
}

type ControlFlow struct {
	Storage  *string `yaml:"storage" validate:"required"`
	Sessions *[]struct {
		Name string `yaml:"name" validate:"required"`
		Path string `yaml:"path" validate:"required"`
	} `yaml:"sessions,flow" validate:"required"`
	Balances *struct {
		SessionsRef []string `yaml:"sessions" validate:"required"`
	} `yaml:"balances" validate:"required"`
	Mappings  *map[string]string `yaml:"mappings" validate:"required"`
	Publisher *struct {
		Type     string `yaml:"type" validate:"required"`
		TopicArn string `yaml:"topic_arn"`
		Region   string `yaml:"region"`
	} `yaml:"publisher"`
	IgnoredAccounts  *[]string           `yaml:"ignored_accounts" validate:"required"`
	IgnoredSymbols   *[]string           `yaml:"ignored_symbols" validate:"required"`
	TargetAllocation *map[string]float64 `yaml:"target_allocation" validate:"required"`
	s3Config         *S3Config
	ioProvider       IOProvider
	publisher        Publisher
}

func (this *ControlFlow) String() string {
	return fmt.Sprintf("Storage: %v, Session: %v, Balances: %v, Mappings: %v", this.Storage, this.Sessions, this.Balances, this.Mappings)
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Parse(data []byte) *ControlFlow {
	validate := validator.New()
	cf := &ControlFlow{}

	err := yaml.Unmarshal(data, cf)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	err = validate.Struct(cf)
	if err != nil {
		log.Fatal(err)
	}
	if cf.Sessions == nil {
		log.Print("its nil")
	}

	switch *cf.Storage {
	case "file":
		log.Println("Using file io provider")
		cf.ioProvider = &FileIO{}
	case "s3":
		s3Config := &S3Config{}
		err = yaml.Unmarshal(data, s3Config)
		if err != nil {
			log.Fatal(err)
		}
		err = validate.Struct(s3Config)
		if err != nil {
			log.Fatal(err)
		}
		cf.s3Config = s3Config
		log.Println("Using s3 io provider")
		cf.ioProvider = NewS3IO(*cf.s3Config.Region, *cf.s3Config.Bucket, *cf.s3Config.Prefix)
	default:
		log.Fatalf("Unknown storage option: %s", *cf.Storage)
	}

	switch cf.Publisher.Type {
	case "sns":
		if cf.Publisher.TopicArn == "" {
			log.Fatal("Publisher type is sns: topic_arn is required.")
		}
		log.Println("Using sns publisher")
		cf.publisher = NewSNSPublisher(cf.Publisher.Region, cf.Publisher.TopicArn)
	default:
		cf.publisher = &NullPublisher{}
	}

	return cf
}

type Func func()
type SessionNode struct {
	Name string
	Fn   Func
}

func (this *ControlFlow) Execute() {
	sessions := make(map[string]*api.Session)
	// Load refresh token from file and then redeem refresh token
	for _, sessionSection := range *this.Sessions {
		log.Printf("Loading session %s\n", sessionSection.Path)
		refreshToken := this.loadRefreshToken(sessionSection.Path)
		sessions[sessionSection.Name] = this.redeem(refreshToken, sessionSection.Path)
	}
	// Pull data from accounts
	portfolio := Portfolio{}
	for _, session := range sessions {
		log.Printf("Checking portfolio balance..\n")
		checker := &Checker{session}
		portfolio = append(portfolio, checker.Get()...)
	}
	Must(this.ioProvider.Write(portfolio, "portfolio.json"))
	log.Print(portfolio)
	// Filter out ignored symbols
	Filter(this.IgnoredSymbols, this.IgnoredAccounts, &portfolio)
	aggregates := Aggregate(this.Mappings, &portfolio)
	log.Print(aggregates)
	bytes, err := json.Marshal(aggregates)
	if err != nil {
		log.Fatalf("failed marshaling aggregation")
	}

	Must(this.ioProvider.Write(bytes, "aggregated.json"))

	diff, percent := CalculatePercentBalance(aggregates, this.TargetAllocation)

	report := &Report{
		Aggregtae:        aggregates,
		Gap:              diff,
		PercentPortfolio: percent,
	}
	Must(this.publisher.Publish(report))
}

// redeem a refresh token for a new session, and save the new session to a file
func (this *ControlFlow) redeem(refreshToken, filename string) *api.Session {
	session, err := api.Redeem(refreshToken)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Redeemed refresh token successfully")
	j, _ := json.Marshal(session)
	Must(this.ioProvider.Write(j, filename))
	return session
}

// loadRefreshToken reads a previously saved session file and extracts the refresh token
func (this *ControlFlow) loadRefreshToken(filename string) string {
	session := &api.Session{}
	err := this.ioProvider.Read(filename, session)
	if err != nil {
		log.Fatal(err)
	}
	return session.RefreshToken
}

func Load(accessTokenFile string) string {
	jsonBytes, err := ioutil.ReadFile(accessTokenFile)
	if err != nil {
		log.Fatalln(err)
	}
	session := &api.Session{}
	Must(json.Unmarshal(jsonBytes, &session))
	return session.RefreshToken
}

func Redeem(refreshToken, output string) *api.Session {
	session, err := api.Redeem(refreshToken)
	if err != nil {
		log.Fatalln(err)
	}

	j, _ := json.Marshal(session)
	Must(ioutil.WriteFile(output, j, 0644))
	return session
}
