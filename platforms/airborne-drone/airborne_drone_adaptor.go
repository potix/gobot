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
	SetAutoDownloadMode(onOff bool)
	SetCutOutMode(onOff bool) error
	ForceDownload()
	FlatTrim() error
	Emergency() error
	Headlight(left uint8, right uint8) error
	HeadlightFlashStart() error
	HeadlightBlinkStart() error
	HeadlightOscillationStart() error
	HeadlightFlashStop() error
	HeadlightBlinkStop() error
	HeadlightOscillationStop() error
	TakePicture() error
	FTPList(path string) ([]byte, error)
	FTPGet(path string) ([]byte, error)
	FTPDelete(path string) ([]byte, error)
	GetBattery() uint8
	GetFlyingState() uint32
	GetPictureState() uint32
	Roll(duration time.Duration, speedFactor int8) error
	Pitch(duration time.Duration, speedFactor int8) error
	Yaw(duration time.Duration, speedFactor int8) error
	Gaz(duration time.Duration, speedFactor int8) error
	Hover()
	Up(speedFactor int8)
	Down(speedFactor int8)
	Left(speedFactor int8)
	Right(speedFactor int8)
	Forward(speedFactor int8)
	Backward(speedFactor int8)
	TurnLeft(speedFactor int8)
	TurnRight(speedFactor int8)
	Stop()
	NewMultiplexer() *client.Multiplexer
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
	if len(v) > 1 {
		a.config.DownloadPath = v[1]
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
