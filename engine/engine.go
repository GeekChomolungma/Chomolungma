package engine

import (
	"github.com/GeekChomolungma/Chomolungma/engine/binance"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi"
)

var EngineBus *TradeEngine

// each cylinder of the main engine
type Cylinder interface {
	Ignite()
	Flush()
	Flameout()
}

// main engine for trade
type TradeEngine struct {
	Cylinders map[string]Cylinder
}

func (te *TradeEngine) Load() {
	te.Cylinders["binance"] = &binance.BinanCylinder{}
	te.Cylinders["huobi"] = &huobi.HuoBiCylinder{}
}

func (te *TradeEngine) Run() {
	for _, cy := range te.Cylinders {
		cy.Ignite()
		cy.Flush()
	}
}

func (te *TradeEngine) Stop() {
	for _, cy := range te.Cylinders {
		cy.Flameout()
	}
}
