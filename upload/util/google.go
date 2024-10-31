package util

import (
	"io"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/jie123108/glog"
	"golang.org/x/net/context"
)

//  上传到google storage
func SaveFileToGoogleStorage(bucket, filename, contentType string, body_reader io.Reader) (err error) {
	glog.Info("begin SaveFileToGoogleStorage")

	// upload the image to Google Storage
	ctx := context.Background()
	client, err := storage.NewClient(ctx) // Creates a client
	if err != nil {
		glog.Error("Failed to create a client! err:", err)
		return
	}

	glog.Info("SaveFileToGoogleStorage bucket:", bucket, " filename:", filename)

	wc := client.Bucket(bucket).Object(filename).NewWriter(ctx)
	wc.ContentType = contentType
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleOwner}}
	body, _ := ioutil.ReadAll(body_reader)
	if _, err = wc.Write(body); err != nil {
		glog.Error("Read filename(", filename, ") failed! err:", err)
		return
	}
	if err = wc.Close(); err != nil {
		glog.Error("upload(", bucket, ",", filename, ") failed! err:", err)
		return
	}
	// defer body_reader.Close()
	// fmt.Printf("UploadID: %s file uploaded to, %s\n", result.UploadID, result.Location)

	return
}
