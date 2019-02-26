// Copyright 2019 Cohesity Inc.
//
// This package provides function to parse and validate developer's appspec.

package appspecvalidator

import (
  "errors"
  "fmt"
  "os"
  
  "gopkg.in/yaml.v2"
)

type Pair struct {
  appSpecKind, appSpecName interface{}
}

var (
  uniqueAppSpecObject    map[interface{}]bool
  nodePortEnvVarMap      map[string]int
  binarySIMap            map[string]bool
  decimalSIMap           map[string]bool
  cleanupJobEncountered  bool
  uiNodePortEncountered  bool
  uiNodePortEnvVar       string
)

func init() {
  uiNodePortEncountered = false
  cleanupJobEncountered = false
  uniqueAppSpecObject = make(map[interface{}]bool)
  nodePortEnvVarMap = make(map[string]int)
  binarySIMap = map[string]bool{
    "ki": true,
    "Mi": true,
    "Gi": true,
    "Ti": true,
  }
  decimalSIMap = map[string]bool{
    "m": true,
    "" : true,
    "K": true,
    "M": true,
    "G": true,
    "T": true,
  }
}

const (
  kCohesityTagKeyWord    string = "cohesityTag"
  kCohesityCleanupTag    string = "cleanup"
  kCohesityUiNodePortTag string = "ui"
  kVolumeTypeStatic      string = "static"
  kVolumeTypeDynamic     string = "dynamic"
  kCohesityEnvKeyWord    string = "cohesityEnv"
  kBinarySIFormat        string = "BinarySI"
  kDecimalSIFormal       string = "DecimalSI"
)

type VolumeMounts struct {
  Name      *string `yaml:"name"`
  MountPath *string `yaml:"mountPath"`
}

type Requests struct {
  Cpu    *string `yaml:"cpu,omitempty"`
  Memory *string `yaml:"memory,omitempty"`
}

type Resources struct {
  Requests *Requests `yaml:"requests"`
}

type Env struct {
  Name  *string `yaml:"name"`
  Value *string `yaml:"value"`
}

type ContainerSpec struct {
  Name         *string         `yaml:"name"`
  Image        *string         `yaml:"image"`
  Resources    *Resources      `yaml:"resources,omitempty"`
  VolumeMounts []*VolumeMounts `yaml:"volumeMounts,omitempty"`
  Env          []*Env          `yaml:"env,omitempty"`
}

type VolumeSpec struct {
  Name       *string `yaml:"name"`
  FsType     *string `yaml:"fsType"`
  Type       *string `yaml:"volumeType"`
  VolumeName *string `yaml:"volumeName"`
}

type Labels struct {
  App *string `yaml:"app"`
}

type Metadata struct {
  Name        *string `yaml:"name"`
  Labels      *Labels `yaml:"labels,omitempty"`
  CohesityTag *string `yaml:"cohesityTag,omitempty"`
}

type Selector struct {
  MatchLabels *Labels `yaml:"matchLabels"`
}

type TemplateSpec struct {
  Containers []*ContainerSpec `yaml:"containers"`
  Volumes    []*VolumeSpec    `yaml:"volumes,omitempty"`
}

type Template struct {
  Metadata     *Metadata     `yaml:"metadata"`
  TemplateSpec *TemplateSpec `yaml:"spec"`
}

type Ports struct {
  Port        *int    `yaml:"port"`
  Protocol    *string `yaml:"protocol"`
  Name        *string `yaml:"name"`
  CohesityTag *string `yaml:"cohesityTag,omitempty"`
  CohesityEnv *string `yaml:"cohesityEnv,omitempty"`
}

type Replicas struct {
  Fixed *int `yaml:"fixed"`
  Share *int `yaml:"share,omitempty"`
  Min   *int `yaml:"min,omitempty"`
  Max   *int `yaml:"max,omitempty"`
}

type Spec struct {
  Replicas    *Replicas `yaml:"replicas,omitempty"`
  ServiceName *string   `yaml:"serviceName,omitempty"`
  Selector    *Selector `yaml:"selector,omitempty"`
  Type        *string   `yaml:"type,omitempty"`
  ClusterIp   *string   `yaml:"clusterIp,omitempty"`
  Ports       []*Ports  `yaml:"ports,omitempty"`
  Template    *Template `yaml:"template,omitempty"`
}

type AppSpec struct {
  ApiVersion *string   `yaml:"apiVersion"`
  Kind       *string   `yaml:"kind"`
  Metadata   *Metadata `yaml:"metadata"`
  Spec       *Spec     `yaml:"spec"`
}

// Validates the metadata of the AppSpec.
func validateMetadata(appSpecMetadata *Metadata, kind *string) error {
  // If there's a cohesity tag in metadata, then it must say 'cleanup' and kind
  // can only be "Job".
  if appSpecMetadata.CohesityTag != nil {
    tagString := *appSpecMetadata.CohesityTag
    if tagString != kCohesityCleanupTag {
      errMsg := fmt.Sprintf("Invalid tag %s, expected %s.", tagString,
        kCohesityCleanupTag)
      return errors.New(errMsg)
    }
    if *kind != "Job" {
      errMsg := fmt.Sprintf("Invalid kind %s for tag %s, expected Job.", *kind,
        kCohesityCleanupTag)
      return errors.New(errMsg)
    }
    if cleanupJobEncountered {
      errMsg := "At most one cleanup job supported."
      return errors.New(errMsg)
    }
    cleanupJobEncountered = true
  }
  return nil
}

// Validates the volumeMounts spec of the AppSpec.
func validateVolumeMounts(volumeMounts []*VolumeMounts) error {
  for _, volumeMount := range volumeMounts {
    if volumeMount.Name == nil {
      return errors.New("VolumeMount name missing.")
    }
    if volumeMount.MountPath == nil {
      return errors.New("VolumeMount mountPath missing.")
    }
  }
  return nil
}

// Validates resources quantities like cpu and memory.
func validateResourceQuantity(quantityStr string) error {

  var suffix string
  if quantityStr == "" {
    return errors.New("Empty resource quantity.")
  }
  // Get the position of the suffix which is the first non digit.
  suffixPos := -1
  for i, char := range quantityStr {
    val := char - '0'
    if !(val >= 0 && val <= 9) {
      suffixPos = i
      break
    }
  }

  if suffixPos == -1 {
    // No suffix in this case, which is valid too. But rewrite suffix_pos to
    // the string length as it is used to get the substring which provides the
    // value.
    suffix = ""
    suffixPos = len(quantityStr)
  } else {
    suffix = quantityStr[suffixPos:len(quantityStr)]
  }

  // Search for the suffix in both Binary and Decimal multiplier maps
  _, isSuffixBinary := binarySIMap[suffix]
  _, isSuffixDecimal := decimalSIMap[suffix]

  if isSuffixBinary || isSuffixDecimal {
    return nil
  } else {
    return errors.New("Invalid resource quantity " + quantityStr)
  }
}

// Validates container resources.
func validateContainerResources(resources *Resources) error {
  // The caller must check for requests to be valid before calling this method.
  if resources.Requests.Cpu != nil {
    err := validateResourceQuantity(*resources.Requests.Cpu)
    if err != nil {
      return err
    }
  }
  if resources.Requests.Memory != nil {
    err := validateResourceQuantity(*resources.Requests.Memory)
    if err != nil {
      return err
    }
  }
  return nil
}

// Validates the volume specs of the AppSpec.
func validateVolumes(volumes []*VolumeSpec) error {

  for _, volume := range volumes {
    if volume.Name == nil {
      return errors.New("Volume name missing.")
    }
    if volume.FsType == nil {
      return errors.New("Volume fsType missing.")
    }
    if volume.Type == nil {
      return errors.New("Volume volumeType missing.")
    }
    if *volume.Type == kVolumeTypeStatic && volume.VolumeName == nil {
      return errors.New("Static volume volumeName missing.")
    }
  }
  return nil
}

//  Validates the containter specs of the AppSpec.
func validateContainers(containers []*ContainerSpec) error {
  var err error
  for _, container := range containers {
    if container.Name == nil {
      return errors.New("Container name missing.")
    }
    if container.Image == nil {
      return errors.New("Container image missing.")
    }

    // If containers have volume mounts, they need to be validated.
    if container.VolumeMounts != nil {
      err = validateVolumeMounts(container.VolumeMounts)
      if err != nil {
        return err
      }
    }
    if container.Resources != nil {
      err = validateContainerResources(container.Resources)
      if err != nil {
        return err
      }
    }
  }
  return nil
}

// Validate all components of a given object spec.
func validateSpec(appSpecObject *AppSpec) error {
  var err error
  replicaSpec := appSpecObject.Spec.Replicas
  if replicaSpec != nil {
    if replicaSpec.Fixed == nil && replicaSpec.Share == nil {
      return errors.New("Replica specification incorrect.")
    }
    if replicaSpec.Fixed != nil {
      if replicaSpec.Max != nil || replicaSpec.Min != nil ||
        replicaSpec.Share != nil {
        return errors.New("Replica specification incorrect.")
      }
    }
  }

  if appSpecObject.Spec.Template == nil {
    return errors.New("Spec Template missing.")
  }

  if appSpecObject.Spec.Template.TemplateSpec == nil {
    return errors.New("Template Specification missing.")
  }

  if appSpecObject.Spec.Template.TemplateSpec.Containers == nil {
    return errors.New("Template Containers missing.")
  }

  containers := appSpecObject.Spec.Template.TemplateSpec.Containers

  err = validateContainers(containers)
  if err != nil {
    return err
  }

  if appSpecObject.Spec.Template.TemplateSpec.Volumes != nil {
    err = validateVolumes(appSpecObject.Spec.Template.TemplateSpec.Volumes)
    if err != nil {
      return err
    }
  }
  return nil
}

// Validates the service spec of the AppSpec.
func validateService(appSpecObject *AppSpec) error {

  if appSpecObject.Metadata == nil {
    return errors.New("Service metadata missing.")
  }

  if appSpecObject.Metadata.Name == nil {
    return errors.New("Service metadata name missing.")
  }

  if appSpecObject.Metadata.Labels == nil {
    return errors.New("Service metadata labels missing.")
  }

  if appSpecObject.Spec == nil {
    return errors.New("Service spec missing.")
  }

  if appSpecObject.Spec.Type != nil {
    if *appSpecObject.Spec.Type != "NodePort" &&
      *appSpecObject.Spec.Type != "ClusterIP" {
      return errors.New("Service Spec Type invalid." +
        "Only NodePort and ClusterIP are allowed.")
    }
  } else {
    return errors.New("Service Spec Type is missing.")
  }

  if *appSpecObject.Spec.Type == "NodePort" {

    if appSpecObject.Spec.Ports == nil {
      errMsg := "Port must be specified if the service is of type NodePort."
      return errors.New(errMsg)
    }

    var hasUiTag bool = false

    for _, entry := range appSpecObject.Spec.Ports {
      // Check whether the nodeports in this service have the UI tag.  Not that
      // the UI node port could also be tagged to be passed as an environment
      // variable.
      if entry.CohesityTag != nil {
        // We only support the 'ui' tag at present.
        tagStr := entry.CohesityTag
        if *tagStr != kCohesityUiNodePortTag {
          errMsg := fmt.Sprintf("Invalid nodeport tag: expected %s, "+
            "got %s.", kCohesityUiNodePortTag, *tagStr)
          return errors.New(errMsg)
        }

        if hasUiTag {
          return errors.New("Only one ui tag is supported.")
        }
        if uiNodePortEncountered {
          return errors.New("At most one UI node port supported.")
        }
        uiNodePortEncountered = true

        hasUiTag = true
      }
      // If a nodePort has cohesityEnv tag, that means the value of that tag
      // is an environment variable that needs to be passed to all the pods.
      // If there are multiple nodePorts which have cohesityEnv tag, then the
      // environment variables specified in the tag must be unique across all
      // nodePorts.
      if entry.CohesityEnv != nil {
        // Check the validity and uniqueness of tag value which is to be used
        // as the environment variable.
        envStr := entry.CohesityEnv
        if envStr == nil {
          return errors.New("CohesityEnv empty.")
        }
        _, ok := nodePortEnvVarMap[*envStr]

        if *envStr == "" || ok || uiNodePortEnvVar == *envStr {
          errMsg := fmt.Sprintf("CohesityEnv: " + *envStr +
            " is not unique in the appspec.")
          return errors.New(errMsg)
        }

        if hasUiTag {
          uiNodePortEnvVar = *envStr
        } else {
          nodePortEnvVarMap[*envStr] = 0
        }
      }
    }
  }

  if *appSpecObject.Spec.Type == "ClusterIp" {
    if appSpecObject.Spec.ClusterIp != nil {
      if *appSpecObject.Spec.ClusterIp != "none" {
        return errors.New("ClusterIp if specified, can only be set to none.")
      }
    }
  }

  if appSpecObject.Spec.Selector == nil {
    return errors.New("Service spec selector missing.")
  }
  return nil
}

// Validates an individual AppSpec object.
func validateAppSpec(appSpecObject *AppSpec) error {
  var err error
  appSpecMetadata := appSpecObject.Metadata

  if appSpecMetadata == nil {
    return errors.New("AppSpecObject metadata missing.")
  }

  if appSpecMetadata.Name == nil {
    return errors.New("AppSpecObject metadata name missing.")
  }

  appSpecName := *appSpecMetadata.Name

  if appSpecObject.Kind == nil {
    errMsg := fmt.Sprintf("AppSpecObject kind is missing. Name %s",appSpecName)
    return errors.New(errMsg)
  }

  appSpecKind := *appSpecObject.Kind
  apiVersion  := *appSpecObject.ApiVersion

  if appSpecObject.ApiVersion == nil {
    errMsg := fmt.Sprintf("Apiversion missing.Kind: %s. Name: %s", appSpecKind, appSpecName)
    return errors.New(errMsg)
  }

  if appSpecKind == "StatefulSet" || appSpecKind == "ReplicaSet" {
    if apiVersion != "apps/v1" {
      errMsg := fmt.Sprintf("Incorrect api version.Kind: %s. Name: %s", appSpecKind, appSpecName)
      return errors.New(errMsg)
    }
  }

  if appSpecKind == "Service" && apiVersion != "v1" {
    errMsg := fmt.Sprintf("Incorrect api version. Object kind: %s. Name: %s", appSpecKind, appSpecName)
    return errors.New(errMsg)
  }

  if appSpecKind == "Job" && apiVersion != "batch/v1" {
    errMsg := fmt.Sprintf("Incorrect api version. Object kind: %s. Name: %s", appSpecKind, appSpecName)
    return errors.New(errMsg)
  }

  appSpecObj := Pair{appSpecKind, appSpecName}

  if _, ok := uniqueAppSpecObject[appSpecObj]; ok {
    errMsg := fmt.Sprintf("No two AppSpecObjects of same kind "+
      "can have same name. kind: %s. Name: %s", appSpecKind, appSpecName)
    return errors.New(errMsg)
  }

  uniqueAppSpecObject[appSpecObj] = true

  if appSpecKind == "StatefulSet" || appSpecKind == "Job" ||
    appSpecKind == "ReplicaSet" {
    err = validateMetadata(appSpecMetadata, &appSpecKind)
    if err != nil {
      errMsg := fmt.Sprint(err)
      errMsg = errMsg + fmt.Sprintf("Kind: %s. Name %s",appSpecKind,appSpecName)
      return err
    }
    err = validateSpec(appSpecObject)
    if err != nil {
      errMsg := fmt.Sprintf("Kind: %s. Name %s",appSpecKind,appSpecName)
      errMsg = fmt.Sprint(err) + errMsg
      return err
    }
  } else if appSpecKind == "Service" {
    err = validateService(appSpecObject)
    if err != nil {
      errMsg := fmt.Sprintf("Kind: %s. Name %s",appSpecKind,appSpecName)
      errMsg = fmt.Sprint(err) + errMsg
      return err
    }
  } else {
    errMsg := "Object kind is not one of StatefulSet, Job, " +
      "ReplicaSet, Service." + fmt.Sprintf("Kind: %s. Name %s",appSpecKind,appSpecName)
    return errors.New(errMsg)
  }
  return nil
}

// ParseAndValidateAppSpec takes the user input appspec, parses and validates it.

func ParseAndValidateAppSpec(InputAppSpecFile string) error {

  appSpecFile, err := os.Open(InputAppSpecFile)
  if err != nil {
    return err
  }
  dec := yaml.NewDecoder(appSpecFile)

  appSpecObjects := make(map[interface{}]interface{})

  for dec.Decode(&appSpecObjects) == nil {
    var appSpec AppSpec
    appSpecObject, err := yaml.Marshal(appSpecObjects)
    if err != nil {
      errMsg := "Error in marshalling appspec." + fmt.Sprint(err)
      return errors.New(errMsg)
    }
    fmt.Printf(string(appSpecObject))
    err = yaml.Unmarshal([]byte(string(appSpecObject)), &appSpec)
    if err != nil {
      errMsg := "Error in unmarshalling appspec." + fmt.Sprint(err)
      return errors.New(errMsg)
    }

    err = validateAppSpec(&appSpec)
    if err != nil {
      errMsg := "Error in validating appspec." + fmt.Sprint(err)
      return errors.New(errMsg)
    }

    for key := range appSpecObjects {
      delete(appSpecObjects, key)
    }
  }
  return nil
}
