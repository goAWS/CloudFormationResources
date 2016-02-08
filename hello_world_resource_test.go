package cloudformationresources

import (
	"encoding/json"
	"testing"

	gocf "github.com/crewjam/go-cloudformation"
)

const mockEvent = `
{
  "RequestType": "Create",
  "ServiceToken": "arn:aws:lambda:us-west-2:123412341234:function:mock",
  "ResponseURL": "https://cloudformation-custom-resource-response-uswest2.s3-us-west-2.amazonaws.com/arn%3Aaws%3Acloudformation%3Aus-west-2%3A123412341234%3Astack/SpartaApplication/7b1329a0-cdbb-11e5-8739-50d5ca0184d2%7CConfigS3b53d1a94bfa7819abce144e99da6bb69becf0421%7Cf6f7ab4f-319d-4726-af8b-dbe56c1714a9?AWSAccessKeyId=AKIAI4KYMPPRGIACET5Q&Expires=1454871368&Signature=gt42bZajCWi1rLp5myu4X9MYbRg%3D",
  "StackId": "arn:aws:cloudformation:us-west-2:123412341234:stack/CloudFormationResources",
  "RequestId": "f6f7ab4f-319d-4726-af8b-dbe56c1714a9",
  "LogicalResourceId": "SomeLogicalIdb53d1a94bfa7819abce144e99da6bb69becf0421",
  "ResourceType": "Custom::goAWS::HelloWorld",
  "ResourceProperties": {
    "ServiceToken": "arn:aws:lambda:us-west-2:123412341234:function:SpartaApplication-S3CustomResourced9468234fca3ffb5-18V7808Y2VSHY",
    "Message": "World"
  }
}
`

func TestCreateHelloWorld(t *testing.T) {
	resHello := gocf.NewResourceByType(HelloWorld)
	customResource := resHello.(HelloWorldResource)
	customResource.Message = "Hello world"
}

func TestCreateHelloWorldNewInstances(t *testing.T) {
	resHello1 := gocf.NewResourceByType(HelloWorld)
	customResource1 := resHello1.(HelloWorldResource)

	resHello2 := gocf.NewResourceByType(HelloWorld)
	customResource2 := resHello2.(HelloWorldResource)

	if &customResource1 == &customResource2 {
		t.Errorf("gocf.NewResourceByType failed to make new instances")
	}
}

func TestProcessEvent(t *testing.T) {
	var request CustomResourceRequest
	err := json.Unmarshal([]byte(mockEvent), &request)
	if nil != err {
		t.Errorf("Failed to unmarshal JSON to CustomResourceRequest", err)
	}
	_, err = Handle(&request)
	if nil != err {
		t.Errorf("Failed to Process CustomResourceRequest: %s", err)
	}
}

func TestProcessUnknownEvent(t *testing.T) {
	var request CustomResourceRequest
	err := json.Unmarshal([]byte(mockEvent), &request)
	if nil != err {
		t.Errorf("Failed to unmarshal JSON to CustomResourceRequest", err)
	}
	request.ResourceType = "Custom::goAWS::62AA0C7B-2011-4FD8-9433-E4420552022B"
	_, err = Handle(&request)
	if nil == err {
		t.Errorf("Failed to reject unknown resource type: %s", err)
	}
}
