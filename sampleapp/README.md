View Browser App
=================

This App runs a webserver and exposes two endpoints: 

`http://<app-ip>:<node-port>/views`  

`http://<app-ip>:<node-port>/views/<view-name>`.
 
This is to demonstrate how to write an App for Cohesity's App
Platform. It also servers as an example for consumption of Cohesity's [App](https://github.com/cohesity/app-sdk-go)
and [Management](https://github.com/cohesity/management-sdk-go) Golang SDKs.

## Installation
In order to successfully build and run Sample App, you are required to 
have the following setup in your system: [Golang](https://golang.org/doc/install)

After installing GO, use `go get` to install the sample app: viewbrowser:

```go get github.com/cohesity/cohesity-appspec/viewbrowser```

This will also get all the dependencies including Cohesity App and 
Management Go SDKs.

## Container Environment Parameters
The App Environment Container has the following parameters initialized by 
Cohesity App Server.
```
HOST_IP  # The Host IP on which the container is running.
APPS_API_ENDPOINT_IP # Cohesity App Server IP.
APPS_API_ENDPOINT_PORT # Cohesity App Server Port.
APP_AUTHENTICATION_TOKEN # Authetication Token to make Cohesity App API calls. 
```
We use the above variables in various use cases to initialize and make call to  App server.

## Using App & Management SDK
The Sample App uses
Cohesity provides [App](https://github.com/cohesity/app-sdk-go)
and [Management](https://github.com/cohesity/management-sdk-go) SDKs to 
make it easy to write Apps onto Cohesity Management Platform.

Importing the packages:
```
import (
    "github.com/cohesity/app-sdk-go/appsdk",
    "github.com/cohesity/management-sdk-go/managementsdk"
)
```

Init the App Client:
```
apiEndpointIp = os.Getenv("APPS_API_ENDPOINT_IP")
apiEndpointPort = os.Getenv("APPS_API_ENDPOINT_PORT")
appAuthenticationToken = os.Getenv("APP_AUTHENTICATION_TOKEN")
appClient = CohesityAppSdk.NewAppSdkClient(appAuthenticationToken,
                                           apiEndpointIp,
                                           apiEndpointPort)
```
App Client in action:
```
//Getting App Settings.
appSettings, err := appClient.Settings().GetAppSettings()
```

Getting  ManagementAccessToken:
```
managementTokenResponse, err := appClient.TokenManagement().CreateManagementAccessToken() 
managementAccessToken.AccessToken = managementTokenResponse.AccessToken
managementAccessToken.TokenType = managementTokenResponse.TokenType
```
Initializing
```
managementClient = ManagementSdk.NewCohesityClientWithToken(apiEndpointIp, 
                                                            managementAccessToken)
```               

Using Management Client to Cohesity Views:
```
 viewsResult, err := managementClient.Views().GetViews(viewNames, viewBoxNames,
            matchPartialNames, maxCount, maxViewId, includeInactive, tenantIds,
            allUnderHierarchy, viewBoxIds, jobIds, sortByLogicalUsage, matchAliasNames)
```

Using Management Client to Mount View:
```
err := appClient.Mount().CreateMount(&mountOptions)
```

Note: 

All the models for Management Structs & Variables are defined [here](https://github.com/cohesity/management-sdk-go/models)

All the models for App Structs & Variables are defined [here](https://github.com/cohesity/app-sdk-go/models)


## Packaging the App
Please refer to [this](README-Container.md) section to containerize this 
application.
## Questions & Feedback
We would love to hear from you. Please send your questions and feedback to: 
*developer@cohesity.com*
