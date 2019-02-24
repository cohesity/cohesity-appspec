// Copyright 2019 Cohesity Inc.
//
// View browser web server that expose apis to get views from the app 
// and browse the files and folders inside the view.

package viewbrowserserver

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "path/filepath"
  "sync"

  CohesityAppSdk "github.com/cohesity/app-sdk-go/appsdk"
  appModels "github.com/cohesity/app-sdk-go/models"
  ManagementSdk "github.com/cohesity/management-sdk-go/managementsdk"
  managementModels "github.com/cohesity/management-sdk-go/models"

  "github.com/go-martini/martini"
  "github.com/golang/glog"
  "github.com/cohesity/cohesity-app-spec/viewbrowser/data"
)

var (
  appAuthenticationToken string
  apiEndpointIp          string
  apiEndpointPort        string
  appClient              CohesityAppSdk.COHESITYAPPSDK
  managementClient       ManagementSdk.COHESITYMANAGEMENTSDK
)

func init() {
  apiEndpointIp = os.Getenv("APPS_API_ENDPOINT_IP")
  apiEndpointPort = os.Getenv("APPS_API_ENDPOINT_PORT")
  appAuthenticationToken = os.Getenv("APP_AUTHENTICATION_TOKEN")
}

type FileBrowserServer struct {
  // Port on which the HTTP server will listen.
  httpPort int

  // The Martini server.
  martiniServer *martini.ClassicMartini
}

const (
  kBrowseFileHandlerApi string = "/views/:viewname/dir"
  kCohesityMountPath    string = "/cohesity/mount"
  kRetry                int    = 3

  // Constants for file types.
  kFile      string = "kFile"
  kDirectory string = "kDirectory"
  kSymlink   string = "kSymlink"
)

// NewFileBrowserServer creates new instance of filebrowser server.
func NewFileBrowserServer() *FileBrowserServer {
  ws := &FileBrowserServer{httpPort: 25695}

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

func (fs *FileBrowserServer) Start() {
  // Initializing appclient before starting the server.
  appClient = CohesityAppSdk.NewAppSdkClient(
    appAuthenticationToken, apiEndpointIp, apiEndpointPort)
  glog.Infoln("starting server.")
  var wg sync.WaitGroup
  fs.startServer()
  wg.Wait()
}

func (fs *FileBrowserServer) startServer() {
  endpt := fmt.Sprintf(":%v", fs.httpPort)
  glog.Infof("Listening on %v", endpt)
  if err := http.ListenAndServe(endpt, fs.martiniServer); err != nil {
    glog.Errorf(fmt.Sprint(err))
    panic(err)
  }
}

func dirExists(dirPath string) bool {
  if fi, err := os.Stat(dirPath); err == nil {
    // Check if the file at the given path is a directory or not.
    return fi.Mode().IsDir()
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
      dirEntry.Type = kDirectory
    } else if !file.Mode().IsRegular() {
      dirEntry.Type = kSymlink
    } else {
      dirEntry.Type = kFile
    }
    glog.Infoln("Entry under "+absoluteDirPath+": %+v\n", *dirEntry)
    readDirResult.Entries = append(readDirResult.Entries, dirEntry)
  }
  return readDirResult, nil
}

func (fs *FileBrowserServer) GetViewsHandler(resp http.ResponseWriter,
  req *http.Request) {
  
  // Get management token to make iris calls.Retry incase of failure.
  var managementAccessToken managementModels.AccessToken
 for i :=0 ;i<=kRetry ; i++ {
   managementTokenResponse, err := appClient.TokenManagement().CreateManagementAccessToken()
    if err ==nil {
      managementAccessToken.AccessToken = managementTokenResponse.AccessToken
      managementAccessToken.TokenType = managementTokenResponse.TokenType
      break
     }
    if i == kRetry {
      glog.Errorf(fmt.Sprint(err))
      resp.WriteHeader(http.StatusBadRequest)
    }
  }

  // Setting management token to initialize management client.
  managementClient = ManagementSdk.NewCohesityClientWithToken(apiEndpointIp, &managementAccessToken)

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
  Clusterviews := viewsResult.Views
  ClusterViewIDMap := make(map[int]string)

  // ClusterViewsInfo gives information about all the views in a cluster.
  var clusterViewsInfo data.ClusterViewsInfo
  var viewsInfo []*data.ViewInfo

  // Iterating over cluster views and storing viewname and id in a map.
  for _, view := range Clusterviews {
    ClusterViewIDMap[int(*view.ViewId)] = *view.Name
    var viewInfo data.ViewInfo
    viewInfo.ViewName = *view.Name
    viewInfo.ViewId = int(*view.ViewId)
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
  var appViewsInfo data.AppViewsInfo
  viewsInfo = viewsInfo[:0]

  appInstanceSettings := appSettings.AppInstanceSettings

  if appInstanceSettings.ReadViewPrivileges.PrivilegesType == appModels.PrivilegesType_KALL {
    appViewsInfo.ViewsInfo = clusterViewsInfo.ViewsInfo
  } else if appInstanceSettings.ReadViewPrivileges.PrivilegesType == appModels.PrivilegesType_KSPECIFIC {
    viewIds := appInstanceSettings.ReadViewPrivileges.ViewIds
    for _, viewId := range *viewIds {
      var viewInfo data.ViewInfo
      viewInfo.ViewName = ClusterViewIDMap[int(viewId)]
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

func (fs *FileBrowserServer) BrowseFilesHandler(pathParams martini.Params,
  resp http.ResponseWriter, req *http.Request) {

  viewName := pathParams["viewname"]

  if len(viewName) == 0 {
    errMsg := "View name not specified in request params."
    glog.Errorln(errMsg)
    http.Error(resp, errMsg, http.StatusBadRequest)
  }

  dirName := viewName + "_dir"
  var optPtr = new(string)
  *optPtr = "ro"
  mountOptions := appModels.MountOptions{
    ViewName:      viewName,
    DirName:       dirName,
    MountProtocol: appModels.MountProtocol_KNFS,
    MountOptions:  optPtr,
  }

  err := appClient.Mount().CreateMount(&mountOptions)

  if err != nil {
    glog.Errorf(fmt.Sprint(err))
    resp.WriteHeader(http.StatusInternalServerError)
    return
  }

  // Unmount after the successful/unsuccessful operation.
  defer func() {
    if err := appClient.Mount().DeleteUnmount(dirName); err != nil {
      errMsg := "Unmount failed with error: " + err.Error()
      glog.Errorln(errMsg)
      return
    }
  }()

  // Get the files under the directory.
  absoluteDirPath := filepath.Join(kCohesityMountPath, dirName)
  readDirResult, err := browseFiles(absoluteDirPath)
  if err != nil {
    errMsg := "Directory " + dirName + " does not exisit inside the " +
      "view " + viewName
    glog.Errorln(errMsg)
    http.Error(resp, errMsg, http.StatusBadRequest)
  }

  dataBuffer, err := json.MarshalIndent(readDirResult, "", " ")
  if err != nil {
    glog.Errorln("Error in marshalling read Dir result.")
    resp.WriteHeader(http.StatusInternalServerError)
  }
  resp.WriteHeader(http.StatusOK)
  resp.Write(dataBuffer)
  return
}
