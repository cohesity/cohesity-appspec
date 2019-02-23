package utils

import (
  "errors"
  "fmt"
  "path"
  "strconv"
  "strings"
  "syscall"
)

// All constants here must have unique names.
const (
  // These mount command options are self contained in the option keyword.
  kMountOptionNolock  string = "nolock"
  kMountOptionHard    string = "hard"
  kMountOptionSoft    string = "soft"
  kMountOptionIntr    string = "intr"
  kMountOptionSync    string = "sync"
  kMountOptionNoac    string = "noac"
  kMountOptionNoatime string = "noatime"
  kMountOptionRo      string = "ro"

  // These mount command options are of the form: option=value.
  kMountOptionRsize    string = "rsize"
  kMountOptionWsize    string = "wsize"
  kMountOptionNfsVers  string = "nfsvers"
  kMountOptionSmbVers  string = "vers"
  kMountOptionRetrans  string = "retrans"
  kMountOptionTimeo    string = "timeo"
  kMountOptionRetry    string = "retry"
  kMountOptionUid      string = "uid"
  kMountOptionGid      string = "gid"
  kMountOptionUsername string = "username"
  kMountOptionPassword string = "password"

  // These are used to track mutually exclusive options.
  kMountOptionType string = "type" // hard vs soft
)

// Helper to check if mount directory is valid.
// Examples of valid mount directories:
//  1. "/a"
//  2. "/abc/def"
//  3. "/abc/def/" (ok to have trailing "/").
// Examples of invalid mount directories:
//  1. "/" (can't mount to root).
//  2. "a", "abc/def", "../abc"(must specify absolutepath).

func isValidMountDir(mountDir string) bool {
  return len(mountDir) > 1 && path.IsAbs(mountDir)
}

//-------------------------------------------------------------------

// Helper to validate the supported NFS options.
func NfsValidation(optSlice []string, optMap map[string]bool) error {
  // Check various options based on slice length.
  if len(optSlice) == 1 {
    switch optSlice[0] {
    case kMountOptionNolock:
      break
    case kMountOptionIntr:
      break
    case kMountOptionSync:
      break
    case kMountOptionNoac:
      break
    case kMountOptionNoatime:
      break
    case kMountOptionRo:
      break
    case kMountOptionHard:
      fallthrough
    case kMountOptionSoft:
      // Check and record for mutally exclusive options.
      if optMap[kMountOptionType] == true {
        return errors.New("Mutually exclusive options set")
      }
      optMap[kMountOptionType] = true
      break
    default:
      return errors.New("Invalid option")
    }
  } else if len(optSlice) == 2 {
    switch optSlice[0] {
    case kMountOptionRsize:
      fallthrough
    case kMountOptionWsize:
      fallthrough
    case kMountOptionRetrans:
      fallthrough
    case kMountOptionTimeo:
      fallthrough
    case kMountOptionRetry:
      fallthrough
    case kMountOptionNfsVers:
      // Check for validity of value (an integer)
      if _, err := strconv.ParseInt(optSlice[1], 10, 64); err != nil {
        return errors.New("Invalid value (expected integer)")
      }
      break
    case kMountOptionUid:
      break
    case kMountOptionGid:
      break
    default:
      return errors.New("Invalid option")
    }
  }

  // Remember the option that was just checked.
  optMap[optSlice[0]] = true
  return nil
}

//------------------------------------------------------------------

// Helper to validate the supported SMB options.
func SmbValidation(optSlice []string, optMap map[string]bool) error {
  // Check various options based on slice length.
  if len(optSlice) == 1 {
    switch optSlice[0] {
    case kMountOptionHard:
      fallthrough
    case kMountOptionSoft:
      // Check and record for mutally exclusive options.
      if optMap[kMountOptionType] == true {
        return errors.New("Mutually exclusive options set")
      }
      optMap[kMountOptionType] = true
      break
    default:
      return errors.New("Invalid option")
    }
  } else if len(optSlice) == 2 {
    switch optSlice[0] {
    case kMountOptionSmbVers:
      // Check for validity of value (an integer)
      if _, err := strconv.ParseInt(optSlice[1], 10, 64); err != nil {
        return errors.New("Invalid value (expected integer)")
      }
      break
    case kMountOptionUsername:
      break
    case kMountOptionPassword:
      break
    case kMountOptionUid:
      break
    case kMountOptionGid:
      break
    default:
      return errors.New("Invalid option")
    }
  }

  // Remember the option that was just checked.
  optMap[optSlice[0]] = true
  return nil
}

//-------------------------------------------------------------------

// Helper to validate the supported mount options.
func ValidateMountOptions(options string,
  protoValidation func([]string, map[string]bool) error) error {

  // Split the list of options passed by their comma seperated values
  optList := strings.Split(options, ",")
  optMap := make(map[string]bool)

  // Iterate over options which are of two types:
  // 1. option (OR)
  // 2. option=value
  for _, optStr := range optList {

    // Split the option based on the "=" sign and check the slice length.
    optSlice := strings.Split(optStr, "=")
    if len(optSlice) < 1 || len(optSlice) > 2 {
      return errors.New("Invalid option")
    }

    // Is option repeated ?
    if val, ok := optMap[optSlice[0]]; ok {
      if val == true {
        return errors.New("Repeated option")
      }
    }

    // Run the protocol based validation function.
    if err := protoValidation(optSlice, optMap); err != nil {
      return err
    }

  }
  return nil
}

//-----------------------------------------------------------------------

// Helper to get the filesystem id for a given directory. If error is set
// fsid is unspecified.
func GetDirFsid(dir string) (syscall.Fsid, error) {
  var statfs syscall.Statfs_t
  var fsid syscall.Fsid
  err := syscall.Statfs(dir, &statfs)
  if err == nil {
    fsid = statfs.Fsid
  }
  return fsid, err
}

//-----------------------------------------------------------------------

// Helper to check if something has been mounted under mountDir.
func IsMounted(mountDir string) bool {
  // Check if it is a valid mount directory name.
  if !isValidMountDir(mountDir) {
    panic(fmt.Sprintf("Invalid mount dir %v", mountDir))
  }

  // Get the parent directory. Trim the trailing "/" when getting the parent
  // directory. E.g. "/a/b/c/" and "/a/b/c" are both valid mount directories
  // and should yield the same parent "/a/b".
  // Note: It is safe to use len(mountPath)-2 as isValidMountDir checks for
  // length > 1.
  mountPath := mountDir
  if string(mountPath[len(mountPath)-1]) == "/" {
    mountPath = mountPath[:len(mountPath)-2]
  }
  parentDir := path.Dir(mountPath)

  // Get the filesystem id for both mount directory and its parent directory.
  var mountDirFsid, parentDirFsid syscall.Fsid
  var err error
  mountDirFsid, err = GetDirFsid(mountDir)
  if err != nil {
    return false
  }
  parentDirFsid, err = GetDirFsid(parentDir)
  if err != nil {
    return false
  }

  // If both filesystem ids are the same, there is nothing mounted in mountDir.
  if mountDirFsid == parentDirFsid {
    return false
  }

  // All checks passed.
  return true
}

//------------------------------------------------------------------------
