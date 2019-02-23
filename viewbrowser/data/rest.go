// Copyright 2019 Cohesity Inc.
//
// Author: Abhilash velalam

// This file defines the json objects and structs implementing the
// filebrowser's Restful API exposed to the apps.

package data

// ManagementTokenResponse is the struct to represent
// management api result.
type ManagementTokenResponse struct {
  ErrorCode       int    `json:"errorCode,omitempty"`
  ManagementToken string `json:"accessToken,omitempty"`
  TokenType       string `json:"tokenType,omitempty"`
  ErrorMessage    string `json:"message,omitempty"`
}

// ViewInfo is the struct to represent the information
// of the views.
type ViewInfo struct {
  ViewName string `json:"name,omitempty"`
  ViewId   int    `json:"viewId,omitempty"`
}

// ClusterViewsInfo is the struct to represent the
// information of the views in the cluster.
type ClusterViewsInfo struct {
  ViewsInfo []*ViewInfo `json:"views,omitempty"`
}

// AppViewsInfo is the struct to represent the
// information of the views accessible by the app.
type AppViewsInfo struct {
  ViewsInfo []*ViewInfo `json:"views,omitempty"`
}

// ViewPrivileges is the struct to represent privileges
// of the view.
type ViewPrivileges struct {
  Type    *string `json:"privilegesType,omitempty"`
  ViewIds []*int  `json:"viewIds,omitempty"`
}

// AppInstanceSettings is the struct to represent
// appinstance settings.
type AppInstanceSettings struct {
  ReadViewPrivileges      *ViewPrivileges `json:"readViewPrivileges,omitempty"`
  ReadWriteViewPrivileges *ViewPrivileges `json:"readWriteViewPrivileges,omitempty"`
}

// AppSettings is the sturct to represent settings of the app.
type AppSettings struct {
  AppInstanceSettings *AppInstanceSettings `json:"appInstanceSettings,omitempty"`
}

// ReadDirResult is the struct to return the result of read directory.
type ReadDirResult struct {
  // Entries is the array of files and folders that are immediate children
  // of the parent directory specified in the request.
  Entries []*DirEntry `json:"entries"`
}

// DirEntry is the struct to represent a file, folder or symlink.
type DirEntry struct {
  // DirEntryType is the type of entry i.e. file/folder/symlink.
  Type string ` json:"type,omitempty"`
  // Name is the name of the file or folder. For /test/file.txt, name will be
  // file.txt.
  Name string `json:"name,omitempty"`
  // FilePath is the path of the file/directory relative to the view.
  FilePath string `json:"filePath,omitempty"`
  // Size of the file of folder.
  Size int64 `json:"size,omitempty"`
}
