package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// set by goreleaser during build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	region      = flag.String("region", "us-east-1", "AWS Region to make requests to")
	exitNoEks   = flag.Bool("exitNoEks", false, "if IAM Roles for Service Accounts environment variables are not detected, exit without error")
	sessionName = flag.String("sessionName", "", "if set will be used as role session name. Session Arn will be in format arn:aws:sts::<AccountNumber>:assumed-role/$AWS_ROLE_ARN/<sessionName>")
)

var (
	errInvalidEnv = errors.New("needed environment variable not set or with invalid value")
)

const (
	// always provided by a Pod running inside EKS with IAM Roles for Service Accounts enabled.
	AwsRoleArn              = "AWS_ROLE_ARN"
	AwsWebIdentityTokenFile = "AWS_WEB_IDENTITY_TOKEN_FILE"

	// can overwrite -sessionName flag.
	AwsSessionName = "AWS_SESSION_NAME"
)

func mainErr() error {
	logf("awstokengen version: %s, commit: %s, date: %s\n", version, commit, date)

	roleArn := os.Getenv(AwsRoleArn)
	if roleArn == "" {
		return errors.Wrapf(errInvalidEnv, "%s must be set", AwsRoleArn)
	}

	webIdentityTokenFile := os.Getenv(AwsWebIdentityTokenFile)
	if webIdentityTokenFile == "" {
		return errors.Wrapf(errInvalidEnv, "%s must be set", AwsWebIdentityTokenFile)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return errors.Wrap(err, "unable to load SDK config")
	}
	cfg.Region = *region

	newSts := sts.NewFromConfig(cfg)

	bts, err := ioutil.ReadFile(webIdentityTokenFile)
	if err != nil {
		return errors.Wrap(err, "could not read web-identity-token from file")
	}

	var sessName string
	if sessNameEnv := os.Getenv(AwsSessionName); sessNameEnv != "" {
		sessName = sessNameEnv
	} else if *sessionName != "" {
		sessName = *sessionName
	} else {
		sessName = uuid.New().String()
	}

	in := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleArn),
		RoleSessionName:  aws.String(sessName),
		WebIdentityToken: aws.String(string(bts)),
	}
	res, err := newSts.AssumeRoleWithWebIdentity(context.Background(), in)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stdout,
		"export AWS_ACCESS_KEY_ID=%s AWS_SECRET_ACCESS_KEY=%s AWS_SESSION_TOKEN=%s",
		*res.Credentials.AccessKeyId,
		*res.Credentials.SecretAccessKey,
		*res.Credentials.SessionToken,
	)
	if err != nil {
		return err
	}

	logf("assumed role arn: %s\n", *res.AssumedRoleUser.Arn)
	logf("valid until:      %v\n", *res.Credentials.Expiration)

	return nil
}

func main() {
	flag.Parse()

	err := mainErr()
	if errors.Is(err, errInvalidEnv) && *exitNoEks {
		logf("not running on EKS, exiting\n")
	} else if err != nil {
		logf("%#v\n", err)
		os.Exit(1)
	}
}

func logf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
