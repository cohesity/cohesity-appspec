// Copyright 2019 Cohesity Inc.
//
// This file defines the json objects and structs implementing the
// viewbrowser app's Restful API exposed to the apps.

package data

// Management token information obtained in response to the
// ManagementAccessToken API
type ManagementTokenResponse struct {
  ErrorCode       int    `json:"errorCode,omitempty"`
  ManagementToken string `json:"accessToken,omitempty"`
  TokenType       string `json:"tokenType,omitempty"`
  ErrorMessage    string `json:"message,omitempty"`
}

// View Identification information.
type ViewInfo struct {
  ViewName string `json:"name,omitempty"`
  ViewId   int    `json:"viewId,omitempty"`
}

// Represents the information of views.
type ViewsInformation struct {
  ViewsInfo []*ViewInfo `json:"views,omitempty"`
}

// Represents privileges of a view.
type ViewPrivileges struct {
  Type    *string `json:"privilegesType,omitempty"`
  ViewIds []*int  `json:"viewIds,omitempty"`
}

// Represents privileges of an app.
type AppInstanceSettings struct {
  ReadViewPrivileges      *ViewPrivileges `json:"readViewPrivileges,omitempty"`
  ReadWriteViewPrivileges *ViewPrivileges `json:"readWriteViewPrivileges,omitempty"`
}

// Represents settings of an app.
type AppSettings struct {
  AppInstanceSettings *AppInstanceSettings `json:"appInstanceSettings,omitempty"`
}

// ReadDirResult is the struct to return the result of read directory.
type ReadDirResult struct {
  // Entries is the array of files and folders that are immediate children
  // of the parent directory specified in the request.
  Entries []*DirEntry `json:"entries"`
}

// Represents a file, folder or symlink.
type DirEntry struct {
  // DirEntry Type is the type of entry i.e. file/folder/symlink.
  Type string ` json:"type,omitempty"`
  // Name is the name of the file or folder.Eg. for /test/file.txt, name will be
  // file.txt.
  Name string `json:"name,omitempty"`
  // FilePath is the path of the file/directory relative to the view.
  FilePath string `json:"filePath,omitempty"`
  // Size of the file or folder.
  Size int64 `json:"size,omitempty"`
}
