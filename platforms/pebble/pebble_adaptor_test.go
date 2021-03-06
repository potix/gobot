package pebble

import (
	"testing"

	"github.com/potix/gobot/gobottest"
)

func initTestPebbleAdaptor() *PebbleAdaptor {
	return NewPebbleAdaptor("pebble")
}

func TestPebbleAdaptor(t *testing.T) {
	a := initTestPebbleAdaptor()
	gobottest.Assert(t, a.Name(), "pebble")
}
func TestPebbleAdaptorConnect(t *testing.T) {
	a := initTestPebbleAdaptor()
	gobottest.Assert(t, len(a.Connect()), 0)
}

func TestPebbleAdaptorFinalize(t *testing.T) {
	a := initTestPebbleAdaptor()
	gobottest.Assert(t, len(a.Finalize()), 0)
}
