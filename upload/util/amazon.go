package util

import (	
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jie123108/glog"
	"github.com/jay-wlj/gobaselib/yf"
	"io"
	"errors"
)

var AmazonSession *session.Session
var smsCtx *sns.SNS
var endpoint_arn string


type AmazonCfg struct {
	AccessKey string
	SecretKey string
	Region string
}

func (t *AmazonCfg)Init() (err error) {
	if t.AccessKey == "" || t.SecretKey == "" || t.Region == "" {
		glog.Error("SmsInitFromConfig fail! conf args invalid!")
		err = errors.New(yf.ERR_ARGS_INVALID)
		return
	}

	// fmt.Println("creating session")
	awsConfig := &aws.Config{
		//need to put access key and secret key
		Credentials: credentials.NewStaticCredentials(t.AccessKey, t.SecretKey, ""),
		Region:      aws.String(t.Region),
	}

	AmazonSession = session.Must(session.NewSession(awsConfig))
	// fmt.Println("session created")

	smsCtx = sns.New(AmazonSession)
	// fmt.Println("service created")

	// paramendpoint := &sns.CreatePlatformEndpointInput{
	// 	//need to put ARN created in AWS SNS
	// 	PlatformApplicationArn: aws.String("arn:aws:sns:ap-northeast-1:353728732266:betterchain"),
	// }
	// endresp, enderr := smsCtx.CreatePlatformEndpoint(paramendpoint)
	// if enderr != nil {
	// 	// fmt.Println("CreatePlatformEndpointInput error=%s" , enderr)
	// 	err = enderr
	// 	return
	// }
	// endpoint_arn = *endresp.EndpointArn
	// fmt.Println("endpoint_arn: ", endpoint_arn)
	return
}

func SmsSend(tel, message string) (err error) {

	params := &sns.PublishInput{
		Message:     aws.String(message),
		PhoneNumber: aws.String(tel),
		// TargetArn:   aws.String(endpoint_arn),
	}
	_, err1 := smsCtx.Publish(params)

	if err1 != nil {
		err = err1
		// fmt.Println(err.Error())
		return
	}
	return
}


func s3CheckExist(bucket string, key string) (exist bool, err error) {
	input := &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)}
	s3inst := s3.New(AmazonSession)

	_, err = s3inst.GetObject(input)
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && (aerr.Code() == s3.ErrCodeNoSuchKey || aerr.Code() == s3.ErrCodeNoSuchBucket) {
			exist = false
			err = nil
		}
		return
	}
	exist = true
	return
}


// 上传到亚马逊存储服务
func SaveFileToS3(bucket, filename, contentType string, body_reader io.Reader) (err error) {
	// The session the S3 Uploader will use
	// sess := session.Must(session.NewSession())
	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(AmazonSession)

	var exist bool
	exist, err = s3CheckExist(bucket, filename)
	if err != nil {
		glog.Error("S3CheckExist(", bucket, ",", filename, ") failed! err:", err)
		return
	}

	if exist {
		glog.Info("*** file[", filename, "] is in s3 server !")
		return
	}

	// Upload the file to S3.
	result, e := uploader.Upload(&s3manager.UploadInput{
		ACL:         aws.String("public-read"),
		Bucket:      aws.String(bucket),
		Key:         aws.String(filename),
		Body:        body_reader,
		ContentType: aws.String(contentType),
	})
	if err= e; err != nil {
		glog.Error("upload(", bucket, ",", filename, ") failed! err:", err)
		return 
	}

	glog.Infof("UploadID: %s file uploaded to, %s\n", result.UploadID, result.Location)
	
	return
}
