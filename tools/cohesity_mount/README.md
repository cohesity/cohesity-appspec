Cohesity Mount Tool
=================

This tool helps App developers to mount Cohesity Views onto the App 
containers. 

## Installation

```bash
go get github.com/cohesity/cohesity-appspec/tools/cohesity_mount
```

## Build

```bash
cd $GOPATH/cohesity/cohesity-appspec/tools/cohesity_mount
go build cohesity_mount.go
```

## Run

NFS mount 
```bash
./cohesity_mount --view <view-name> --mountdir <mount> --options <options> 
--protocol nfs --namespace <view-namespace> 
```
SMB mount
```bash
./cohesity_mount --view <view-name> --options <options> --mountdir <mount> 
--protocol smb --namespace <view-namespace> --username  <smb-username> 
--password <smb-password>
```

Utility arguments:
```bash
Usage of ./cohesity_mount:
  
  --mountdir string
    	Directory on which the view is to be mounted.
  --namespace string
    	Namespace of the view that is to be mounted. (default "fs")
  --options string
    	Mount options. 
  --password string
    	Password for smb mount.
  --protocol string
    	mount protocol [nfs|smb] (default "nfs")
  --stderrthreshold value
    	logs at or above this threshold go to stderr.
  --username string
    	Username for smb mount.
  --alsologtostderr
    	log to standard error as well as files.
  --log_backtrace_at value
    	when logging hits line file:N, emit a stack trace.
  --log_dir string
    	If non-empty, write log files in this directory.
  --logtostderr
    	log to standard error instead of files.
```

## Questions & Feedback
We would love to hear from you. Please send your questions and feedback to: 
*apps@cohesity.com*
