package engine

import "github.com/GeekChomolungma/Chomolungma/engine/huobi"

var EngineBus *TradeEngine

// each cylinder of the main engine
type Cylinder interface {
	Ignite()
	Flameout()
}

// main engine for trade
type TradeEngine struct {
	Cylinders map[string]Cylinder
}

func (te *TradeEngine) Load() {
	te.Cylinders["huobi"] = &huobi.HuoBiCylinder{}
}

func (te *TradeEngine) Run() {
	for _, cy := range te.Cylinders {
		cy.Ignite()
	}
}

func (te *TradeEngine) Stop() {
	for _, cy := range te.Cylinders {
		cy.Flameout()
	}
}
