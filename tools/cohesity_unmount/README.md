Cohesity Unmount Tool
=================

This tool helps App developers to mount Cohesity Views onto the App 
containers. 

## Installation

```bash
go get github.com/cohesity/cohesity-appspec/tools/cohesity_unmount
```

## Build

```bash
cd $GOPATH/cohesity/cohesity-appspec/tools/cohesity_unmount
go build cohesity_unmount.go
```

## Run

NFS mount 
```bash
./cohesity_unmount --unmountdir=<dir-name>
```

Utility arguments:
```bash
Usage of ./cohesity_unmount:
  
  --unmountdir string
    	Directory where the Cohesity view is to be mounted.
 
```

## Questions & Feedback
We would love to hear from you. Please send your questions and feedback to: 
*developer@cohesity.com*
