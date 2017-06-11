package cloudformationresources

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	gocf "github.com/mweagle/go-cloudformation"
)

func TestUnzip(t *testing.T) {
	resUnzip := gocf.NewResourceByType(ZipToS3Bucket)
	zipResource := resUnzip.(*ZipToS3BucketResource)
	zipResource.DestBucket = gocf.String(os.Getenv("TEST_DEST_S3_BUCKET"))
	zipResource.SrcBucket = gocf.String(os.Getenv("TEST_SRC_S3_BUCKET"))
	zipResource.SrcKeyName = gocf.String(os.Getenv("TEST_SRC_S3_KEY"))
	zipResource.Manifest = map[string]interface{}{
		"Some": "Data",
	}
	// Put it
	logger := logrus.New()
	awsSession := awsSession(logger)
	createOutputs, createError := zipResource.create(awsSession, logger)
	if nil != createError {
		t.Errorf("Failed to create Unzip resource: %s", createError)
	}
	t.Logf("TestUnzip outputs: %#v", createOutputs)

	deleteOutputs, deleteError := zipResource.delete(awsSession, logger)
	if nil != deleteError {
		t.Errorf("Failed to create Unzip resource: %s", createError)
	}
	t.Logf("TestUnzip outputs: %#v", deleteOutputs)
}
