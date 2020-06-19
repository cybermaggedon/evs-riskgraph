package main

import (
	pb "github.com/cybermaggedon/evs-golang-api/protos"
	"github.com/golang/protobuf/ptypes"
	"time"
)

// Gaffer object types etc.
const (
	ENTITY        = "uk.gov.gchq.gaffer.data.element.Entity"
	EDGE          = "uk.gov.gchq.gaffer.data.element.Edge"
	TIMESTAMP_SET = "uk.gov.gchq.gaffer.time.RBMBackedTimestampSet"
	TIME_BUCKET   = "HOUR"
)

// Type of Gaffer properties
type PropertyMap map[string]interface{}

// We handle a timestamp set as a map internally.
type TimestampSet map[uint64]bool

// Gaffer edge
type Edge struct {
	Source      string
	Destination string
	Group       string
	Count       uint64
	Time        TimestampSet
}

// Create a new edge object
func NewEdge(source, destination, group string) *Edge {
	e := Edge{
		Source:      source,
		Destination: destination,
		Group:       group,
		Time:        TimestampSet{},
	}
	return &e
}

// Add time to an edge object
func (e *Edge) AddTime(tm time.Time) *Edge {
	e.Time[uint64(tm.Unix())] = true
	return e
}

// Add to an edge object's count
func (e *Edge) AddCount(count uint64) *Edge {
	e.Count += count
	return e
}

// Combine two edge objects by adding count/time information from second
// to first.
func (e *Edge) Merge(e2 *Edge) {

	e.Count += e2.Count

	for k, _ := range e2.Time {
		e.Time[k] = true
	}

}

// Converts edge objects to Gaffer representation.
func (e *Edge) ToGaffer() map[string]interface{} {

	tset := make([]uint64, 0, len(e.Time))

	for v, _ := range e.Time {
		tset = append(tset, v)
	}

	return map[string]interface{}{
		"class":       EDGE,
		"group":       e.Group,
		"source":      e.Source,
		"destination": e.Destination,
		"directed":    true,
		"properties": PropertyMap{
			"time": PropertyMap{
				TIMESTAMP_SET: PropertyMap{
					"timeBucket": TIME_BUCKET,
					"timestamps": tset,
				},
			},
			"count": e.Count,
		},
	}
}

// Gaffer entity object
type Entity struct {
	Vertex string
	Group  string
	Count  uint64
	Time   TimestampSet
	Network     string
	Type string
}

// Create a new entity object
func NewEntity(vertex, group string) *Entity {
	return &Entity{
		Vertex: vertex,
		Group:  group,
		Time:   TimestampSet{},
	}
}

// Add time to an entity object
func (e *Entity) AddTime(tm time.Time) *Entity {
	e.Time[uint64(tm.Unix())] = true
	return e
}

// Add to an entity object's count
func (e *Entity) AddCount(count uint64) *Entity {
	e.Count += count
	return e
}

// 
func (e *Entity) AddNetwork(network string) *Entity {
	e.Network = network
	return e
}

// 
func (e *Entity) AddType(val string) *Entity {
	e.Type = val
	return e
}

// Combine two entity objects by adding count/time information from second
// to first.
func (e *Entity) Merge(e2 *Entity) {

	e.Count += e2.Count

	for k, _ := range e2.Time {
		e.Time[k] = true
	}

}

// Converts entity objects to Gaffer representation.
func (e *Entity) ToGaffer() map[string]interface{} {

	tset := make([]uint64, 0, len(e.Time))

	for v, _ := range e.Time {
		tset = append(tset, v)
	}

	props := &PropertyMap{
		"time": PropertyMap{
			TIMESTAMP_SET: PropertyMap{
				"timeBucket": TIME_BUCKET,
				"timestamps": tset,
			},
		},
		"count": e.Count,
	}

	if e.Network != "" {
		(*props)["network"] = e.Network
	}

	if e.Type != "" {
		(*props)["type"] = e.Type
	}

	return map[string]interface{}{
		"class":  ENTITY,
		"group":  e.Group,
		"vertex": e.Vertex,
		"properties": props,
	}
}

// Helper functions to create various entities.
func NewActor(address string) *Entity  { return NewEntity(address, "actor") }
func NewResource(v string) *Entity    { return NewEntity(v, "resource") }
func NewRisk(v string) *Entity  { return NewEntity(v, "risk") }

// Helper functions to create various edges
func NewActorResource(src, dest string) *Edge { return NewEdge(src, dest, "actorresource") }
func NewActorRisk(s, d string) *Edge       { return NewEdge(s, d, "actorrisk") }
func NewResourceRisk(s, d string) *Edge    { return NewEdge(s, d, "resourcerisk") }

// Takes an event and outputs the riskgraph elements.
func DescribeRiskElements(ev *pb.Event) ([]*Entity, []*Edge, error) {

	// Get timestamp rounded to nearest second.
	tm, _ := ptypes.Timestamp(ev.Time)
	tm = tm.Round(time.Second)

	actor := ev.Device
	network := ev.Network

	// Start with empty arrays
	entities := []*Entity{}
	edges := []*Edge{}

	for _, v := range ev.Indicators {

		a_e := NewActor(actor).AddTime(tm).AddCount(1)

		if network != "" {
			a_e = a_e.AddNetwork(network)
		}
		
		r_e := NewResource(v.Value).
			AddTime(tm).AddCount(1).AddType(v.Type)
		k_e := NewRisk(v.Category).
			AddTime(tm).AddCount(1)

		entities = append(entities, a_e)
		entities = append(entities, r_e)
		entities = append(entities, k_e)

		ar_e := NewActorResource(actor, v.Value).
			AddTime(tm).AddCount(1)
		ak_e := NewActorRisk(actor, v.Category).
			AddTime(tm).AddCount(1)
		rk_e := NewResourceRisk(v.Value, v.Category).
			AddTime(tm).AddCount(1)

		edges = append(edges, ar_e)
		edges = append(edges, ak_e)
		edges = append(edges, rk_e)

	}

	return entities, edges, nil

}
