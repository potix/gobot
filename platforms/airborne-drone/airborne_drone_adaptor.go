package airbornedrone

import (
	"github.com/potix/gobot/platforms/airborne-drone/client"
	"github.com/potix/gobot"
)

var _ gobot.Adaptor = (*AirborneDroneAdaptor)(nil)

// drone defines expected drone behaviour
type drone interface {
	Takeoff() bool
	Land()
	Up(n float64)
	Down(n float64)
	Left(n float64)
	Right(n float64)
	Forward(n float64)
	Backward(n float64)
	Clockwise(n float64)
	Counterclockwise(n float64)
	Hover()
}

// AirborneDroneAdaptor is gobot.Adaptor representation for the AirborneDrone
type AirborneDroneAdaptor struct {
	name    string
	drone   drone
	config  client.Config
	connect func(*AirborneDroneAdaptor) (drone, error)
}

// NewAirborneDroneAdaptor returns a new AirborneDroneAdaptor and optionally accepts:
//
//  string: The airborneDrones IP Address
//
func NewAirborneDroneAdaptor(name string, v ...string) *AirborneDroneAdaptor {
	a := &AirborneDroneAdaptor{
		name: name,
		connect: func(a *AirborneDroneAdaptor) (drone, error) {
			return client.Connect(a.config)
		},
	}

	a.config = client.DefaultConfig()
	if len(v) > 0 {
		a.config.DroneAddress = v[0]
	}

	return a
}

// Name returns the AirborneDroneAdaptors Name
func (a *AirborneDroneAdaptor) Name() string { return a.name }

// Connect establishes a connection to the airborneDrone
func (a *AirborneDroneAdaptor) Connect() (errs []error) {
	d, err := a.connect(a)
	if err != nil {
		return []error{err}
	}
	a.drone = d
	return
}

// Finalize terminates the connection to the airborneDrone
func (a *AirborneDroneAdaptor) Finalize() (errs []error) { return }
