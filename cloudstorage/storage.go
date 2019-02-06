// The intention of this helper library is to extract any platform
// specific storage code. Libraries utilising this code should need
// no knowledge of the platform on which it is running. Equally, the APIs
// defined here should not include any arguments which are specific to
// any given platform (e.g. a key is required for GCP, however, AWS requires
// credentials to be stored at a specific location on disk).

// When adding a new platform, simply add the platform to the switch statements
// below and create a <platform>-storage.go file in which you implement the
// functions listed under CloudStorage. The libraries utilising this one should not
// need any alterations.

package cloudstorage

import (
	"analytics-common/utils"
)

type CloudStorage interface {
	Init(bucketNameEnvVar string, bucketNameDefault string)
	Upload(path string, data []byte)
	Download(object string, dest string, generation CloudGeneration) error
	GetObjectGeneration(object string) CloudGeneration
}

type CloudGeneration interface {
	Update(value interface{}) error
	Equals(rhs CloudGeneration) bool
}

func New(platform string) CloudStorage {
	switch platform {
	case "gcp":
		storage := GCPStorage{}
		return &storage
	case "aws":
		storage := AWSStorage{}
		return &storage
	}
	utils.Log("Unsupported platform type: %s", platform)
	return nil
}

func NewGeneration(platform string) CloudGeneration {
	switch platform {
	case "gcp":
		generation := GCPGeneration{}
		return &generation
	case "aws":
		generation := AWSGeneration{}
		return &generation
	}
	utils.Log("Unsupported platform type: %s", platform)
	return nil
}
