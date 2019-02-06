/*
	These are the datatypes for the IoC Alert REST interface
*/
package datatypes

import (
	"net"
	"time"

	"github.com/google/uuid"
)

// IoC types
const (
	IoCDnsCat2 = 0
)

// IoCAlert describes the information contained in a generic IoC alert message
//
// The timestamp and ID will be set by the REST, so there's no need to set
// them when calling it.
//
// A valid DNS IoC alert for the REST interface would be:
// {
//		"type": "0",
//		"data": "{\"deviceName\": \"minesweepers-mac\", \"domainName\": \"microsoft.com\", \"sourceIp\": \"192.168.18.43\", \"startTime\": \"2017-09-13T17:22:00Z\"}"
// }
//
// Please note: The JSON standard requires the field names to be in camelCase
type IoCAlert struct {
	Type      int       `json:"type"`
	Data      string    `json:"data"`
	ID        uuid.UUID `json:"id,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// IoCAlerts is a slice of IoC structs
type IoCAlerts []IoCAlert

// DNSIoCAlert is the data specific to a DNS tunnelling IoC
//
// This will be sent as the "data" field of an IoC alert. See the example for // IoCAlert to see what this would look like.
//
// Test with:
// curl -H "Content-Type: application/json" -d '{"type":0, "data": "{\"deviceName\": \"minesweepers-mac\", \"domainName\": \"microsoft.com\", \"sourceIp\": \"192.168.18.43\", \"startTime\": \"2017-09-13T17:22:00Z\"}"}' http://localhost:8080/iocalerts
//
// Please note: The JSON standard requires the field names to be in camelCase
type DNSIoCAlert struct {
	ID         uuid.UUID `json:"id,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
	DeviceName string    `json:"deviceName"`
	DomainName string    `json:"domainName"`
	SourceIP   net.IP    `json:"sourceIp"`
	StartTime  time.Time `json:"startTime"`
}
