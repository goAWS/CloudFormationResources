# CloudFormationResources
Catalog of golang-based CloudFormation CustomResources


## Adding a New Resource

  1. Create a new struct that implements `CustomResourceCommand`
    - This struct **MUST** embed `gocf.CloudFormationCustomResource` as in:

    ```
    type HelloWorldResource struct {
    	gocf.CloudFormationCustomResource
    	Message string
    }
    ```
  2. Add a const representing the resource type to the _CustomResourceType_ enum eg, `HelloWorld`
    - The literal **MUST** start with [Custom::goAWS](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-cfn-customresource.html)
  3. Add the new type to the `customResources` map in the `init` function in <i>cloud_formation_resources.go</i>
  4. Add a new case to the `switch/case` in `Process()` that initializes the custom resource and assigns the new value to the `commandInstance` as in:

    ```
    case HelloWorld:
  		customCommand := HelloWorldResource{}
  		if err := json.Unmarshal([]byte(string(marshaledProperties)), &customCommand); nil != err {
  			return nil, fmt.Errorf("Failed to unmarshal ResourceProperties for %s", request.ResourceType)
  		}
  		commandInstance = customCommand
    ```
