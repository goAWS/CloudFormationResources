package cloudformationresources

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	gocf "github.com/crewjam/go-cloudformation"
)

// HelloWorldResource is a simple POC showing how to create custom resources
type HelloWorldResource struct {
	gocf.CloudFormationCustomResource
	Message string
}

func (command HelloWorldResource) create(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	logger.Info("Hello: ", command.Message)
	return nil, nil
}

func (command HelloWorldResource) update(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	logger.Info("Nice to see you again: ", command.Message)
	return nil, nil
}

func (command HelloWorldResource) delete(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	logger.Info("Goodbye: ", command.Message)
	return nil, nil
}
