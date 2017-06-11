package cloudformationresources

import (
	"encoding/json"
	"testing"

	"github.com/Sirupsen/logrus"
	gocf "github.com/mweagle/go-cloudformation"
)

const mockEvent = `
{
  "RequestType": "Create",
  "ServiceToken": "arn:aws:lambda:us-west-2:123412341234:function:mock",
  "ResponseURL": "",
  "StackId": "arn:aws:cloudformation:us-west-2:123412341234:stack/CloudFormationResources",
  "RequestId": "f6f7ab4f-319d-4726-af8b-dbe56c1714a9",
  "LogicalResourceId": "SomeLogicalIdb53d1a94bfa7819abce144e99da6bb69becf0421",
  "ResourceProperties": {
    "ServiceToken": "arn:aws:lambda:us-west-2:123412341234:function:SpartaApplication-S3CustomResourced9468234fca3ffb5-18V7808Y2VSHY",
    "Message": "World",
    "GoAWSType" : "Custom::goAWS::HelloWorldResource"
  }
}
`

func TestCreateHelloWorld(t *testing.T) {
	resHello := gocf.NewResourceByType(HelloWorld)
	customResource := resHello.(*HelloWorldResource)
	customResource.Message = "Hello world"
}

func TestCreateHelloWorldNewInstances(t *testing.T) {
	resHello1 := gocf.NewResourceByType(HelloWorld)
	customResource1 := resHello1.(*HelloWorldResource)

	resHello2 := gocf.NewResourceByType(HelloWorld)
	customResource2 := resHello2.(*HelloWorldResource)

	if &customResource1 == &customResource2 {
		t.Errorf("gocf.NewResourceByType failed to make new instances")
	}
}

func TestExecuteCreateHelloWorld(t *testing.T) {
	resHello1 := gocf.NewResourceByType(HelloWorld)
	customResource1 := resHello1.(*HelloWorldResource)
	customResource1.Message = "Create resource here"

	logger := logrus.New()
	awsSession := awsSession(logger)
	createOutputs, createError := customResource1.create(awsSession, logger)
	if nil != createError {
		t.Errorf("Failed to create HelloWorldResource: %s", createError)
	}
	t.Logf("HelloWorldResource outputs: %s", createOutputs)
}

func TestProcessEvent(t *testing.T) {
	var request CustomResourceRequest
	err := json.Unmarshal([]byte(mockEvent), &request)
	if nil != err {
		t.Errorf("Failed to unmarshal JSON to CustomResourceRequest: %s", err)
	}
	err = Handle(&request, logrus.New())
	if nil != err {
		t.Errorf("Failed to Process CustomResourceRequest: %s", err)
	}
}

func TestProcessUnknownEvent(t *testing.T) {
	var request CustomResourceRequest
	err := json.Unmarshal([]byte(mockEvent), &request)
	if nil != err {
		t.Errorf("Failed to unmarshal JSON to CustomResourceRequest: %s", err)
	}
	request.ResourceProperties["GoAWSType"] = "Custom::goAWS::62AA0C7B-2011-4FD8-9433-E4420552022B"
	err = Handle(&request, logrus.New())
	if nil == err {
		t.Errorf("Failed to reject unknown resource type: %s", err)
	}
}
