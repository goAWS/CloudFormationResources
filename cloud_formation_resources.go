package cloudformationresources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	gocf "github.com/crewjam/go-cloudformation"
)

// CloudFormationLambdaEvent represents the event data sent during a
// Lambda invocation in the context of a CloudFormation operation.
// Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/crpg-ref-requests.html
type CloudFormationLambdaEvent struct {
	RequestType           string
	ResponseURL           string
	StackID               string `json:"StackId"`
	RequestID             string `json:"RequestId"`
	ResourceType          string
	LogicalResourceID     string `json:"LogicalResourceId"`
	PhysicalResourceID    string `json:"PhysicalResourceId"`
	ResourceProperties    map[string]interface{}
	OldResourceProperties map[string]interface{}
}

// CustomResourceRequest is the go representation of the CloudFormation resource
// request
type CustomResourceRequest struct {
	RequestType        string
	ResponseURL        string
	StackID            string `json:"StackId"`
	RequestID          string `json:"RequestId"`
	LogicalResourceID  string `json:"LogicalResourceId"`
	ResourceProperties map[string]interface{}
}

// GoAWSCustomResource does SpartaApplication-S3CustomResourced9468234fca3ffb5
type GoAWSCustomResource struct {
	gocf.CloudFormationCustomResource
	GoAWSType string
}

// CustomResourceCommand defines operations that a CustomResource must implement.
// The return values are either operation outputs or an error value that should
// be used in the response to the CloudFormation AWS Lambda response.
type CustomResourceCommand interface {
	create(session *session.Session,
		logger *logrus.Logger) (map[string]interface{}, error)

	update(session *session.Session,
		logger *logrus.Logger) (map[string]interface{}, error)

	delete(session *session.Session,
		logger *logrus.Logger) (map[string]interface{}, error)
}

const (
	// CreateOperation is a request to create a resource
	// @enum CloudFormationOperation
	CreateOperation = "Create"
	// DeleteOperation is a request to delete a resource
	// @enum CloudFormationOperation
	DeleteOperation = "Delete"
	// UpdateOperation is a request to update a resource
	// @enum CloudFormationOperation
	UpdateOperation = "Update"
)

var (
	// HelloWorld is the typename for HelloWorldResource
	HelloWorld = cloudFormationResourceType("HelloWorldResource")
)

func customCommandForTypeName(resourceTypeName string, properties *[]byte) (interface{}, error) {
	var unmarshalError error
	var customCommand interface{}
	// ---------------------------------------------------------------------------
	// BEGIN - RESOURCE TYPES
	switch resourceTypeName {
	case HelloWorld:
		command := HelloWorldResource{
			GoAWSCustomResource: GoAWSCustomResource{
				GoAWSType: resourceTypeName,
			},
		}
		if nil != properties {
			unmarshalError = json.Unmarshal([]byte(string(*properties)), &command)
		}
		customCommand = &command
	}
	// END - RESOURCE TYPES
	// ---------------------------------------------------------------------------

	if unmarshalError != nil {
		return nil, fmt.Errorf("Failed to unmarshal properties for type: %s", resourceTypeName)
	}
	if nil == customCommand {
		return nil, fmt.Errorf("Failed to create custom command for type: %s", resourceTypeName)
	}
	return customCommand, nil
}

func customTypeProvider(resourceType string) gocf.ResourceProperties {
	commandInstance, commandError := customCommandForTypeName(resourceType, nil)
	if nil != commandError {
		return nil
	}
	resProperties, ok := commandInstance.(gocf.ResourceProperties)
	if !ok {
		return nil
	}
	return resProperties
}

func init() {
	gocf.RegisterCustomResourceProvider(customTypeProvider)
}

// Returns an AWS Session (https://github.com/aws/aws-sdk-go/wiki/Getting-Started-Configuration)
// object that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func awsSession(logger *logrus.Logger) *session.Session {
	sess := session.New()
	sess.Handlers.Send.PushFront(func(r *request.Request) {
		logger.WithFields(logrus.Fields{
			"Service":   r.ClientInfo.ServiceName,
			"Operation": r.Operation.Name,
			"Method":    r.Operation.HTTPMethod,
			"Path":      r.Operation.HTTPPath,
			"Payload":   r.Params,
		}).Debug("AWS Request")
	})
	return sess
}

// cloudFormationResourceType a string for the resource name that represents a
// custom CloudFormation resource typename
func cloudFormationResourceType(resType string) string {
	return fmt.Sprintf("Custom::goAWS::%s", resType)
}

func sendCloudFormationResponse(customResourceRequest *CustomResourceRequest,
	results map[string]interface{},
	responseErr error, logger *logrus.Logger) error {

	parsedURL, parsedURLErr := url.ParseRequestURI(customResourceRequest.ResponseURL)
	if nil != parsedURLErr {
		return parsedURLErr
	}

	status := "FAILED"
	if nil == responseErr {
		status = "SUCCESS"
	}
	responseData := map[string]interface{}{
		"Status":             status,
		"Reason":             fmt.Sprintf("See the details in the CloudWatch Log Stream"),
		"PhysicalResourceId": customResourceRequest.LogicalResourceID,
		"StackId":            customResourceRequest.StackID,
		"RequestId":          customResourceRequest.RequestID,
		"LogicalResourceId":  customResourceRequest.LogicalResourceID,
		"Data":               results,
	}
	jsonData, jsonError := json.Marshal(responseData)
	if nil != jsonError {
		return jsonError
	}
	responseBuffer := bytes.NewBuffer(jsonData)
	req, httpErr := http.NewRequest("PUT", customResourceRequest.ResponseURL, responseBuffer)
	if nil != httpErr {
		return httpErr
	}
	// Need to use the Opaque field b/c Go will parse inline encoded values
	// which are supposed to be roundtripped to AWS.
	// Ref: https://tools.ietf.org/html/rfc3986#section-2.2
	// Ref: https://golang.org/pkg/net/url/#URL
	req.URL = &url.URL{
		Scheme:   parsedURL.Scheme,
		Host:     parsedURL.Host,
		Opaque:   parsedURL.RawPath,
		RawQuery: parsedURL.RawQuery,
	}
	// Although it seems reasonable to set the Content-Type to "application/json" - don't.
	// The Content-Type must be an empty string in order for the
	// AWS Signature checker to pass.
	// Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html
	req.Header.Set("Content-Type", "")
	req.Header.Set("Content-Length", strconv.Itoa(responseBuffer.Len()))

	logger.WithFields(logrus.Fields{
		"ResponseURL": req.URL,
	}).Debug("CloudFormation ResponseURL")

	client := &http.Client{}
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return httpErr
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Error sending response: %d. Data: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Handle processes the given CustomResourceRequest value
func Handle(request *CustomResourceRequest, logger *logrus.Logger) error {
	session := awsSession(logger)

	logger.WithFields(logrus.Fields{
		"Request": request,
	}).Debug("Incoming request")

	marshaledProperties, marshalError := json.Marshal(request.ResourceProperties)
	if nil != marshalError {
		return marshalError
	}

	commandTypeName := request.ResourceProperties["GoAWSType"].(string)
	commandInstance, commandError := customCommandForTypeName(commandTypeName, &marshaledProperties)
	if nil != commandError {
		return commandError
	}

	// TODO - lift this into a backoff/retry loop
	customCommandHandler := commandInstance.(CustomResourceCommand)
	var operationOutputs map[string]interface{}
	var operationError error
	switch request.RequestType {
	case CreateOperation:
		operationOutputs, operationError = customCommandHandler.create(session, logger)
	case DeleteOperation:
		operationOutputs, operationError = customCommandHandler.delete(session, logger)
	case UpdateOperation:
		operationOutputs, operationError = customCommandHandler.update(session, logger)
	default:
		operationError = fmt.Errorf("Unsupported operation: %s", request.RequestType)
	}
	if "" != request.ResponseURL {
		sendErr := sendCloudFormationResponse(request, operationOutputs, operationError, logger)
		if nil != sendErr {
			logger.WithFields(logrus.Fields{
				"Error": sendErr,
			}).Error("Failed to notify CloudFormation of result.")
		}
	}
	return operationError
}
