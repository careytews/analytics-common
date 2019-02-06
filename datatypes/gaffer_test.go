
package datatypes

import (
	"testing"
	"encoding/json"
//	"fmt"
)

func TestBundle(t *testing.T) {
	v1 := &Bundle{
		"abc": "def",
		"ghi": "jkl",
	}
	s1, _ := json.Marshal(v1)
	if string(s1) != "{\"abc\":\"def\",\"ghi\":\"jkl\"}" {
		t.Error("JSON marshalling Bundle failed")
	}
}

func TestEdge(t *testing.T) {

	v1 := NewEdge("source", "dest", "group")
	s1, _ := json.Marshal(v1)
	t1 := "{\"class\":\"uk.gov.gchq.gaffer.data.element.Edge\",\"destination\":\"dest\",\"directed\":true,\"group\":\"group\",\"properties\":{},\"source\":\"source\"}"
	if string(s1) != t1 {
		t.Error("%s is not %s", string(s1), t1)
	}

	v2 := NewEdge("source", "dest", "group").IsDirected(false).
		SetProperty("hello", "world")
	s2, _ := json.Marshal(v2)
	t2 := "{\"class\":\"uk.gov.gchq.gaffer.data.element.Edge\",\"destination\":\"dest\",\"directed\":false,\"group\":\"group\",\"properties\":{\"hello\":\"world\"},\"source\":\"source\"}"
	if string(s2) != t2 {
		t.Error("%s is not %s", string(s2), t2)
	}

}

func TestEntity(t *testing.T) {

	v1 := NewEntity("node", "group")
	s1, _ := json.Marshal(v1)
	t1 := "{\"class\":\"uk.gov.gchq.gaffer.data.element.Entity\",\"group\":\"group\",\"properties\":{},\"vertex\":\"node\"}"
	if string(s1) != t1 {
		t.Error("%s is not %s", string(s1), t1)
	}

	v2 := NewEntity("node", "group").
		SetProperty("hello", "world")
	s2, _ := json.Marshal(v2)
	t2 := "{\"class\":\"uk.gov.gchq.gaffer.data.element.Entity\",\"group\":\"group\",\"properties\":{\"hello\":\"world\"},\"vertex\":\"node\"}"
	if string(s2) != t2 {
		t.Error("%s is not %s", string(s2), t2)
	}

}

func TestTimestampSet(t *testing.T) {

	v1 := NewEntity("node", "group").
		SetProperty("time", NewTimestampSet("MINUTE").Add(1234))
	s1, _ := json.Marshal(v1)
	t1 := "{\"class\":\"uk.gov.gchq.gaffer.data.element.Entity\",\"group\":\"group\",\"properties\":{\"time\":{\"uk.gov.gchq.gaffer.time.RBMBackedTimestampSet\":{\"timeBucket\":\"MINUTE\",\"timestamps\":[1234]}}},\"vertex\":\"node\"}"
	if string(s1) != t1 {
		t.Error("%s is not %s", string(s1), t1)
	}

}


