package main

import (
	"flag"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var SignSession *session.Session
var re_url *regexp.Regexp

func init() {
	re_url = regexp.MustCompile(`(?m)https://([^.]*)[^/]*/(.*)`)
}

func SignInit(ACCESS_KEY, SECRET_KEY, region string) (err error) {
	// fmt.Println("creating session")
	awsConfig := &aws.Config{
		//need to put access key and secret key
		Credentials: credentials.NewStaticCredentials(ACCESS_KEY, SECRET_KEY, ""),
		Region:      aws.String(region),
	}
	SignSession = session.Must(session.NewSession(awsConfig))

	return
}

// https://betterchain.s3.ap-southeast-1.amazonaws.com/test/upload/img/8a/e3/8ae3a49ecc08efedc8cc4ad29e1acffa122f25aa.jpg
func S3UrlParse(url string) (bucket, path string, err error) {
	matches := re_url.FindStringSubmatch(url)
	if len(matches) == 3 {
		bucket = matches[1]
		path = matches[2]
		return
	}
	err = fmt.Errorf("ERR_INVALID_URL")
	return
}

func S3UrlSign(bucket, path string, expires time.Duration) (signed_url string, err error) {
	svc := s3.New(SignSession)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})
	signed_url, err = req.Presign(expires)
	return
}

func main() {
	var url string
	var expires time.Duration

	var ACCESS_KEY string
	var SECRET_KEY string
	var region string

	flag.StringVar(&url, "url", "", "Url to sign")
	flag.DurationVar(&expires, "expires", time.Hour*24, "expires time, like: 100s, 20m, 50h")

	flag.StringVar(&ACCESS_KEY, "ACCESS_KEY", "", "S3 ACCESS_KEY")
	flag.StringVar(&SECRET_KEY, "SECRET_KEY", "", "S3 SECRET_KEY")
	flag.StringVar(&region, "region", "ap-southeast-1", "s3 region")
	flag.Parse()

	err := SignInit(ACCESS_KEY, SECRET_KEY, region)
	if err != nil {
		fmt.Printf("SignInit(%s, %s, %s) failed! err: %s", ACCESS_KEY, SECRET_KEY, region)
		return
	}

	bucket, path, err := S3UrlParse(url)
	if err != nil {
		fmt.Printf("URL[%s] invalid! ", url)
		return
	}

	signed_url, err := S3UrlSign(bucket, path, expires)
	if err != nil {
		fmt.Printf("S3UrlSign(%s, %v) failed! err: %v", url, expires, err)
		return
	}
	fmt.Println("signed url: ", signed_url)
	return
}
