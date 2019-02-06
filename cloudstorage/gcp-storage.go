package cloudstorage

import (
	"analytics-common/utils"
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"
)

type GCPStorage struct {
	client                  *http.Client
	key                     string
	bucketName              string
	svc                     *storage.Service
	svcoutageretrysleeptime time.Duration
}

type GCPGeneration struct {
	Value int64
}

func (g *GCPStorage) Init(bucketNameEnvVar string, bucketNameDefault string) {
	// We get the key from environment variables since this is a GCP-specific requirement
	g.key = utils.Getenv("KEY", "private.json")
	g.bucketName = utils.Getenv(bucketNameEnvVar, bucketNameDefault)

	sleepduration, err := strconv.Atoi(utils.Getenv("SVC_OUTAGE_RETRYSLEEPTIME", "10"))
	if err != nil {
		utils.Log("Couldn't get SVC_OUTAGE_RETRYSLEEPTIME: %s,  setting to 10 seconds", err.Error())
		sleepduration = 10
	}
	g.svcoutageretrysleeptime = time.Second * time.Duration(int64(sleepduration))
	utils.Log("svcOutageRetrySleepTime set to: %s", g.svcoutageretrysleeptime)

	// We no longer create bucket here as it is created in the provisioning service
	// We create a service as and when needed since the scope requires specific permissions

}

func (g *GCPStorage) createService(scope string) error {
	key, err := ioutil.ReadFile(g.key)
	if err != nil {
		utils.Log("Couldn't read key file: %s", err.Error())
		return err
	}

	config, err := google.JWTConfigFromJSON(key)
	if err != nil {
		utils.Log("JWTConfigFromJSON: %s", err.Error())
		return err
	}

	config.Scopes = []string{scope}

	g.client = config.Client(oauth2.NoContext)

	g.svc, err = storage.New(g.client)
	if err != nil {
		utils.Log("Coulnd't create client: %s", err.Error())
		return err
	}

	return nil
}

// TODO: add checks for file size limits and recommendations for multipart uploads
func (g *GCPStorage) Upload(path string, data []byte) {
	g.createService(storage.DevstorageReadWriteScope)
	var object storage.Object // Google storage
	object.Name = path
	object.Kind = "storage#object"

	rdr := bytes.NewReader(data)

	for {
		_, err := g.svc.Objects.Insert(g.bucketName, &object).Media(rdr).Do()
		if err != nil {
			utils.Log("Couldn't insert in to bucket: %s", err.Error())
			time.Sleep(g.svcoutageretrysleeptime)
		} else {
			// exiting for, when there are no storage service connection issues
			break
		}
	}
}

func (g *GCPStorage) Download(object string, filepath string, generation CloudGeneration) error {
	g.createService(storage.DevstorageReadOnlyScope)

	resp, err := g.svc.Objects.Get(g.bucketName, object).Download()
	if err != nil {
		utils.Log("ERROR: Couldn't get object: %s", err.Error())
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.Log("ERROR: Couldn't read body: %s", err.Error())
		return err
	}

	err = ioutil.WriteFile(filepath, data, 0755)
	if err != nil {
		utils.Log("ERROR: Couldn't write file: %s", err.Error())
		return err
	}

	return nil
}

func (g *GCPStorage) GetObjectGeneration(objectName string) CloudGeneration {
	g.createService(storage.DevstorageReadOnlyScope)
	var object *storage.Object
	var err error
	for {
		object, err = g.svc.Objects.Get(g.bucketName, objectName).Do()
		if err != nil {
			utils.Log("Couldn't get object: %s", err.Error())
			time.Sleep(g.svcoutageretrysleeptime)
		} else {
			// exiting for, when there are no storage service connection issues
			break
		}
	}
	var generation GCPGeneration
	generation.Value = object.Generation

	return &generation
}

func (gg *GCPGeneration) Update(value interface{}) error {
	iVal, ok := value.(int64)
	if ! ok {
		errStr := "GCPGeneration only accepts int64 values to update"
		utils.Log("ERROR: " + errStr)
		return errors.New(errStr)
	}
	gg.Value = iVal
	return nil
}

func (gg *GCPGeneration) Equals(rhs CloudGeneration) bool {
	gGen, ok := rhs.(*GCPGeneration)
	// if the other value is not GCP Generation then its not equal
	if ! ok {
		return false
	}
	// compare values
	return gGen.Value == gg.Value
}
