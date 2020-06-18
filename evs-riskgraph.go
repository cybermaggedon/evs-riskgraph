package main

import (
	evs "github.com/cybermaggedon/evs-golang-api"
	pb "github.com/cybermaggedon/evs-golang-api/protos"
	"log"
)

const ()

type RiskGraph struct {
	*Config

	// Embed EventAnalytic framework
	*evs.EventSubscriber
	evs.Interruptible

	// Gaffer management
	gaffer *Gaffer
}

// Initialisation
func NewRiskGraph(c *Config) *RiskGraph {

	t := &RiskGraph{
		Config: c,
	}

	var err error
	t.EventSubscriber, err = evs.NewEventSubscriber(t.Name, t.Input, t)
	if err != nil {
		log.Fatal(err)
	}

	t.RegisterStop(t)

	// Initialise Gaffer from configuration
	t.gaffer, err = t.Build()
	if err != nil {
		log.Fatal(err)
	}

	return t
}

// Event handler for new events.
func (a *RiskGraph) Event(ev *pb.Event, props map[string]string) error {

	// Convert event to riskgraph model
	entities, edges, _ := DescribeRiskElements(ev)

	// Send elements to  Gaffer loader
	a.gaffer.AddElements(entities, edges)

	return nil

}

func main() {

	gc := NewConfig()
	g := NewRiskGraph(gc)
	log.Print("Initialisation complete")
	g.Run()
	log.Print("Shutdown.")

}
