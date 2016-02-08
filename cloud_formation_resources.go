package cloudformationresources

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	gocf "github.com/crewjam/go-cloudformation"
)

// CustomResourceRequest is the go representation of the CloudFormation resource
// request
type CustomResourceRequest struct {
	RequestType        string
	ServiceToken       string
	ResponseURL        string
	StackID            string `json:"StackId"`
	RequestID          string `json:"RequestId"`
	LogicalResourceID  string `json:"LogicalResourceId"`
	ResourceType       string
	ResourceProperties map[string]interface{}
}

// CustomResourceCommand defines a CloudFormation resource
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

const (
	// HelloWorld is a simple Hello World resource that accepts a single "Message" resource property
	// @enum CustomResourceType
	HelloWorld = "Custom::goAWS::HelloWorld"
)

func customTypeProvider(resourceType string) gocf.ResourceProperties {
	typeRef, ok := customResources[resourceType]
	if !ok {
		return nil
	}
	customResourceElem := reflect.New(typeRef).Elem().Interface()
	resProperties, ok := customResourceElem.(gocf.ResourceProperties)
	if !ok {
		return nil
	}
	return resProperties
}

var customResources map[string]reflect.Type

func init() {
	// Setup the map
	customResources = make(map[string]reflect.Type, 8)
	customResources[HelloWorld] = reflect.TypeOf(HelloWorldResource{})
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

// Handle processes the given CustomResourceRequest value
func Handle(request *CustomResourceRequest) (map[string]interface{}, error) {
	logger := logrus.New()
	logger.Formatter = new(logrus.JSONFormatter)
	session := awsSession(logger)

	if _, exists := customResources[request.ResourceType]; !exists {
		types := make([]string, len(customResources))
		for eachKey := range customResources {
			types = append(types, eachKey)
		}
		return nil, fmt.Errorf("Unregistered CloudFormation CustomResource type requested: <%s>. Registered types: %s", request.ResourceType, types)
	}
	marshaledProperties, err := json.Marshal(request.ResourceProperties)
	if nil != err {
		return nil, err
	}
	var commandInstance CustomResourceCommand

	//
	// INSERT NEW RESOURCES HERE
	///
	// Create the appropriate type
	switch request.ResourceType {
	case HelloWorld:
		customCommand := HelloWorldResource{}
		if err := json.Unmarshal([]byte(string(marshaledProperties)), &customCommand); nil != err {
			return nil, fmt.Errorf("Failed to unmarshal ResourceProperties for %s", request.ResourceType)
		}
		commandInstance = customCommand
	default:
		return nil, fmt.Errorf("Unsupported resource type: %s", request.ResourceType)
	}
	if nil == commandInstance {
		return nil, fmt.Errorf("Failed to create commandInstance for type: %s", request.ResourceType)
	}

	switch request.RequestType {
	case CreateOperation:
		return commandInstance.create(session, logger)
	case DeleteOperation:
		return commandInstance.delete(session, logger)
	case UpdateOperation:
		return commandInstance.update(session, logger)
	default:
		return nil, fmt.Errorf("Unsupported operation: %s", request.RequestType)
	}
}
