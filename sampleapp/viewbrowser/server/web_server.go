// Copyright 2019 Cohesity Inc.
//
// View browser web server that exposes apis to get views from the app,
// list the views and browse the files and folders inside the view.

package viewbrowser

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "path/filepath"
  "sync"

  "github.com/cohesity/app-sdk-go/appsdk"
  appModels "github.com/cohesity/app-sdk-go/models"
  "github.com/cohesity/management-sdk-go/managementsdk"
  managementModels "github.com/cohesity/management-sdk-go/models"
  "github.com/go-martini/martini"
  "github.com/golang/glog"
  "github.com/cohesity/cohesity-app-spec/sampleapp/viewbrowser/data"
)

var (
  // Authentication token for the app to authenticate with the service.
  appAuthenticationToken string

  // Ip address of the host the container is running on.
  hostIp string

  // Ip address of the  API endpoint.
  apiEndpointIp string

  // Port of the API endpoint
  apiEndpointPort string

  // Client to access cohesity AppSdk
  appClient CohesityAppSdk.COHESITYAPPSDK
)

func init() {
  hostIp = os.Getenv("HOST_IP")

  apiEndpointIp = os.Getenv("APPS_API_ENDPOINT_IP")

  apiEndpointPort = os.Getenv("APPS_API_ENDPOINT_PORT")

  appAuthenticationToken = os.Getenv("APP_AUTHENTICATION_TOKEN")
}

type ViewBrowserServer struct {
  // Port on which the HTTP server will listen.
  httpPort int

  // The Martini HTTP server.
  martiniServer *martini.ClassicMartini
}

const (
  kBrowseFileHandlerApi string = "/views/:viewname"
  kCohesityMountPath    string = "/cohesity/mount"
  kRetry                int    = 3
)

// NewFileBrowserServer creates new instance of filebrowser server.
func NewViewBrowserServer() *ViewBrowserServer {
  ws := &ViewBrowserServer{httpPort: 8080}

  // Set up the http handlers.
  wr := martini.NewRouter()

  // Handler to get views from iris.
  wr.Get("/views", ws.GetViewsHandler)

  wr.Get(kBrowseFileHandlerApi, ws.BrowseFilesHandler)

  // Create Martini.
  ms := martini.Classic()
  ms.Action(wr.Handle)

  ws.martiniServer = ms

  return ws
}

func (fs *ViewBrowserServer) Start() {
  // Initializing appclient before starting the server.
  appClient = CohesityAppSdk.NewAppSdkClient(
    appAuthenticationToken, apiEndpointIp, apiEndpointPort)
  glog.Infoln("starting server.")
  var wg sync.WaitGroup
  fs.startServer()
  wg.Wait()
}

func (fs *ViewBrowserServer) startServer() {
  endpt := fmt.Sprintf(":%v", fs.httpPort)
  glog.Infof("Listening on %v", endpt)
  if err := http.ListenAndServe(endpt, fs.martiniServer); err != nil {
    glog.Errorf(fmt.Sprint(err))
    panic(err)
  }
}

func dirExists(dirPath string) bool {
  if fileInfo, err := os.Stat(dirPath); err == nil {
    // Check if the file at the given path is a directory or not.
    return fileInfo.Mode().IsDir()
  }
  return false
}

// BrowseFiles takes absolute path of a directory and returns
// files/folders/symlinks that are immediate children of the directory.
func browseFiles(absoluteDirPath string) (*data.ReadDirResult, error) {
  if !dirExists(absoluteDirPath) {
    errMsg := "Directory " + absoluteDirPath + " does not exisit."
    glog.Errorln(errMsg)
    return nil, fmt.Errorf(errMsg)
  }
  // Get the info of files/folders inside the directory.
  readDirResult := new(data.ReadDirResult)
  files, _ := ioutil.ReadDir(absoluteDirPath)
  for _, file := range files {
    dirEntry := new(data.DirEntry)
    dirEntry.Name = file.Name()
    dirEntry.FilePath = filepath.Join(absoluteDirPath, file.Name())
    dirEntry.Size = file.Size()
    if file.IsDir() {
      dirEntry.Type = "kDirectory"
    } else if !file.Mode().IsRegular() {
      dirEntry.Type = "kSymlink"
    } else {
      dirEntry.Type = "kFile"
    }
    glog.Infoln("Adding entry to read directory under: "+absoluteDirPath+
      " with properties: %+v\n", *dirEntry)
    readDirResult.Entries = append(readDirResult.Entries, dirEntry)
  }
  return readDirResult, nil
}

func (fs *ViewBrowserServer) GetViewsHandler(resp http.ResponseWriter,
  req *http.Request) {

  // Get management token to make iris calls.Retry incase of failure.
  // management token is only valid for 24 hours.
  var managementAccessToken managementModels.AccessToken
  for i := 0; i <= kRetry; i++ {
    managementTokenResponse, err := appClient.TokenManagement().CreateManagementAccessToken()
    if err != nil {
      if i == kRetry {
        glog.Errorf(fmt.Sprint(err))
        resp.WriteHeader(http.StatusBadRequest)
        return
      }
    } else {
      managementAccessToken = managementModels.AccessToken{
        AccessToken: managementTokenResponse.AccessToken,
        TokenType:   managementTokenResponse.TokenType,
      }
      break
    }
  }

  // Setting management token to initialize management client.
  managementClient := CohesityManagementSdk.NewCohesityClientWithToken(hostIp, &managementAccessToken)

  var viewsResult *managementModels.GetViewsResult
  var viewNames, viewBoxNames, tenantIds []string
  var matchPartialNames, includeInactive, allUnderHierarchy,
    sortByLogicalUsage, matchAliasNames *bool
  var maxCount, maxViewId *int64
  var viewBoxIds, jobIds []int64

  viewsResult, err := managementClient.Views().GetViews(viewNames, viewBoxNames,
    matchPartialNames, maxCount, maxViewId, includeInactive, tenantIds,
    allUnderHierarchy, viewBoxIds, jobIds, sortByLogicalUsage, matchAliasNames)
  if err != nil {
    glog.Errorf(fmt.Sprint(err))
    resp.WriteHeader(http.StatusInternalServerError)
    return
  }
  clusterViews := viewsResult.Views
  clusterViewIDMap := make(map[int]string)

  // ClusterViewsInfo gives information about all the views in a cluster.
  var clusterViewsInfo data.ViewsInformation
  var viewsInfo []*data.ViewInfo

  // Iterating over cluster views and storing viewname and id in a map.
  for _, view := range clusterViews {
    clusterViewIDMap[int(*view.ViewId)] = *view.Name
    viewInfo := data.ViewInfo{
      ViewName: *view.Name,
      ViewId:   int(*view.ViewId),
    }
    viewsInfo = append(viewsInfo, &viewInfo)
  }

  clusterViewsInfo.ViewsInfo = viewsInfo

  appSettings, err := appClient.Settings().GetAppSettings()

  if err != nil {
    glog.Errorf(fmt.Sprint(err))
    resp.WriteHeader(http.StatusInternalServerError)
    return
  }

  // appViewsInfo gives information about views acecssible by the app.
  var appViewsInfo data.ViewsInformation
  viewsInfo = viewsInfo[:0]

  // appInstanceSettings give information about the views accessible by the app
  // and their privileges.

  appInstanceSettings := appSettings.AppInstanceSettings

  if appInstanceSettings.ReadViewPrivileges.PrivilegesType == appModels.PrivilegesType_KALL {
    appViewsInfo.ViewsInfo = clusterViewsInfo.ViewsInfo
  } else if appInstanceSettings.ReadViewPrivileges.PrivilegesType == appModels.PrivilegesType_KSPECIFIC {
    viewIds := appInstanceSettings.ReadViewPrivileges.ViewIds
    for _, viewId := range *viewIds {
      var viewInfo data.ViewInfo
      viewInfo.ViewName = clusterViewIDMap[int(viewId)]
      viewInfo.ViewId = int(viewId)
      viewsInfo = append(viewsInfo, &viewInfo)
    }
    appViewsInfo.ViewsInfo = viewsInfo
  }

  dataBuffer, err := json.MarshalIndent(appViewsInfo, "", " ")
  glog.Infoln("Response Json:" + string(dataBuffer))
  resp.WriteHeader(http.StatusOK)
  resp.Write(dataBuffer)
  return
}

// Gets the view name from the url, creates a directory, mount the view
// on the directory created, get the directories information inside the view
// and unmounts the view.
func (fs *ViewBrowserServer) BrowseFilesHandler(pathParams martini.Params,
  resp http.ResponseWriter, req *http.Request) {

  viewName := pathParams["viewname"]

  if len(viewName) == 0 {
    errMsg := "View name not specified in request params."
    glog.Errorln(errMsg)
    http.Error(resp, errMsg, http.StatusBadRequest)
  }

  // Name of the directory that is to be created and to be mounted the view on.
  dirName := viewName + "_dir"
  options := "ro"

  // Options to be specified for the mount api.
  mountOptions := appModels.MountOptions{
    ViewName:      viewName,
    DirName:       dirName,
    MountProtocol: appModels.MountProtocol_KNFS,
    MountOptions:  &options,
  }

  // Api to mount the view.
  err := appClient.Mount().CreateMount(&mountOptions)
  if err != nil {
    glog.Errorf(fmt.Sprint(err))
    resp.WriteHeader(http.StatusBadRequest)
    return
  }

  // Try to unmount after the handler exists irrespective of whether
  // the rest of the op was successful.
  defer func() {
    // Api to unmount the view.
    if err := appClient.Mount().DeleteUnmount(dirName); err != nil {
      errMsg := "Unmount failed with error: " + err.Error()
      glog.Errorln(errMsg)
    }
  }()

  // Get the files under the directory.
  absoluteDirPath := filepath.Join(kCohesityMountPath, dirName)
  readDirResult, err := browseFiles(absoluteDirPath)
  if err != nil {
    errMsg := "Directory " + dirName + " does not exist inside the " +
      "view " + viewName
    glog.Errorln(errMsg)
    http.Error(resp, errMsg, http.StatusBadRequest)
  }

  dataBuffer, err := json.MarshalIndent(readDirResult, "", " ")
  if err != nil {
    glog.Errorln("Error in marshalling readDirresult.")
    resp.WriteHeader(http.StatusInternalServerError)
    return
  }
  resp.WriteHeader(http.StatusOK)
  resp.Write(dataBuffer)
  return
}
