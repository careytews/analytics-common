package datatypes

// Objects which manager Gaffer-encoding of entities

// Design approach...
// Options:
// - Create objects and JSON objects for mapping e.g. Edge JSONEdge
//     Down-side is CPU for the mapping.
// - Create custom JSONMarshal interfaces on each object.  Seems fiddly for
//     what we're trying to do.
// - Create the Gaffer JSON directly and hide detail in API.  This seems like
//     the most direct way to get the right JSON in minimal mapping with a
//     clean API.

// Bundle is a JSON document represented internally as a Go map.  This can
// represent any JSON object.
type Bundle map[string]interface{}

// Edge is a Gaffer edge.
type Edge Bundle

// Create an Edge.  Need source, destination and group.
func NewEdge(source, destination, group string) *Edge {
	return &Edge{
		"class":       "uk.gov.gchq.gaffer.data.element.Edge",
		"group":       group,
		"source":      source,
		"destination": destination,
		"directed":    true,
		"properties":  &Bundle{},
	}
}

// Change Edge's directed flag.
func (e *Edge) IsDirected(is bool) *Edge {
	(*e)["directed"] = is
	return e
}

// Add a property to an edge.
func (e *Edge) SetProperty(key string, value interface{}) *Edge {
	props := (*e)["properties"].(*Bundle)
	(*props)[key] = value
	return e
}

// Entity is a Gaffer node/entity.
type Entity Bundle

// Create a new Entity
func NewEntity(vertex, group string) *Entity {
	return &Entity{
		"class":      "uk.gov.gchq.gaffer.data.element.Entity",
		"group":      group,
		"vertex":     vertex,
		"properties": &Bundle{},
	}
}

// Add a property to an Entity.
func (e *Entity) SetProperty(key string, value interface{}) *Entity {
	props := (*e)["properties"].(*Bundle)
	(*props)[key] = value
	return e
}

// TimestampSet is the RBBBackedTimestampSet type, intended to be used as a
// property value.
type TimestampSet Bundle

// Create new timtestamp set.
func NewTimestampSet(bucket string) *TimestampSet {
	return &TimestampSet{
		"uk.gov.gchq.gaffer.time.RBMBackedTimestampSet": &Bundle{
			"timeBucket": bucket,
			"timestamps": []uint64{},
		},
	}
}

// Add a time to a timestamp set.  We're look at UNIX time here.
func (e *TimestampSet) Add(t uint64) *TimestampSet {
	tss := (*e)["uk.gov.gchq.gaffer.time.RBMBackedTimestampSet"].(*Bundle)
	(*tss)["timestamps"] = append((*tss)["timestamps"].([]uint64), t)
	return e
}
