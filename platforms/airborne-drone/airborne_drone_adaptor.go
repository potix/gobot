package airbornedrone

import (
	"time"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/airborne-drone/client"
)

var _ gobot.Adaptor = (*AirborneDroneAdaptor)(nil)

// drone defines expected drone behaviour
type drone interface {
	Takeoff() bool
	Landing() error
	FrontFlip() error
	BackFlip() error
	RightFlip() error
	LeftFlip() error
	SetMaxAltitude(altitude float32) error
	SetMaxTilt(tilt float32) error
	SetMaxVirticalSpeed(virticalSpeed float32) error
	SetMaxRotationSpeed(rotationSpeed float32) error
	SetContinuousMode(onOff bool)
	Roll(duration time.Duration, speedFactor int8) error
	Pitch(duration time.Duration, speedFactor int8) error
	Yaw(duration time.Duration, speedFactor int8) error
	Gaz(duration time.Duration, speedFactor int8) error
	Hover()
	NewCommander() *client.Commander
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
