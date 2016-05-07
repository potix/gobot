package airbornedrone

import (
	"time"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/airborne-drone/client"
)

var _ gobot.Driver = (*AirborneDroneDriver)(nil)

// AirborneDroneDriver is gobot.Driver representation for the AirborneDrone
type AirborneDroneDriver struct {
	name       string
	connection gobot.Connection
	gobot.Eventer
}

// NewAirborneDroneDriver creates an AirborneDroneDriver with specified name.
//
// It add the following events:
//     'flying' - Sent when the device has taken off.
func NewAirborneDroneDriver(connection *AirborneDroneAdaptor, name string) *AirborneDroneDriver {
	d := &AirborneDroneDriver{
		name:       name,
		connection: connection,
		Eventer:    gobot.NewEventer(),
	}
	d.AddEvent("flying")
	return d
}

// Name returns the AirborneDroneDrivers Name
func (a *AirborneDroneDriver) Name() string { return a.name }

// Connection returns the AirborneDroneDrivers Connection
func (a *AirborneDroneDriver) Connection() gobot.Connection { return a.connection }

// adaptor returns airborneDrone adaptor
func (a *AirborneDroneDriver) adaptor() *AirborneDroneAdaptor {
	return a.Connection().(*AirborneDroneAdaptor)
}

// Start starts the AirborneDroneDriver
func (a *AirborneDroneDriver) Start() (errs []error) {
	return
}

// Halt halts the AirborneDroneDriver
func (a *AirborneDroneDriver) Halt() (errs []error) {
	return
}

// TakeOff makes the drone start flying, and publishes `flying` event
func (a *AirborneDroneDriver) TakeOff(){
	gobot.Publish(a.Event("flying"), a.adaptor().drone.Takeoff())
}

// Land causes the drone to land
func (a *AirborneDroneDriver) Landing() error {
	return a.adaptor().drone.Landing()
}

// FrontFlip causes the drone to front flip
func (a *AirborneDroneDriver) FrontFlip() error {
	return a.adaptor().drone.FrontFlip()
}

// BackFlip causes the drone to back flip
func (a *AirborneDroneDriver) BackFlip() error {
	return a.adaptor().drone.BackFlip()
}

// RightFlip causes the drone to right flip
func (a *AirborneDroneDriver) RightFlip() error {
	return a.adaptor().drone.RightFlip()
}

// LeftFlip causes the drone to left flip
func (a *AirborneDroneDriver) LeftFlip() error {
	return a.adaptor().drone.RightFlip()
}

// SetMaxAltitude causes the drone to set max altitude (2.6 - 10.0)
func (a *AirborneDroneDriver) SetMaxAltitude(altitude float32) error {
	return a.adaptor().drone.SetMaxAltitude(altitude)
}

// SetMaxTilt causes the drone to set max tilt (5 - 25.0
func (a *AirborneDroneDriver) SetMaxTilt(tilt float32) error {
	return a.adaptor().drone.SetMaxTilt(tilt)
}

// SetMaxVirticalSpeed causes the drone to set max virtical speed (0.5 - 2.0)
func (a *AirborneDroneDriver) SetMaxVirticalSpeed(virticalSpeed float32) error {
	return a.adaptor().drone.SetMaxVirticalSpeed(virticalSpeed)
}

// SetMaxRotationSpeed causes the drone to set max rotation speed (0 - 360)
func (a *AirborneDroneDriver) SetMaxRotationSpeed(rotationSpeed float32) error {
	return a.adaptor().drone.SetMaxRotationSpeed(rotationSpeed)
}

// SetContinuousMode causes the drone to set continuous mode
func (a *AirborneDroneDriver) SetContinuousMode(onOff bool) {
	a.adaptor().drone.SetContinuousMode(onOff)
}

// Headlight causes the drone to headlight (0-2255)
func (a *AirborneDroneDriver) Headlight(left uint8, right uint8) error {
	return a.adaptor().drone.Headlight(left, right)
}

// TakePicture causes the drone to take picture
func (a *AirborneDroneDriver) TakePicture() error {
	return a.adaptor().drone.TakePicture()
}

// GetBattery return battery (0 - 100)
func (a *AirborneDroneDriver) GetBattery() uint8 {
	return a.adaptor().drone.GetBattery()
}

// GetFlyingState return flying state
// (0 = landed, 1 = takingoff, 2 = hovering, 3 = flying, 4 = landing, 5 = emergency, 6 = rolling, 7 = init)
func (a *AirborneDroneDriver) GetFlyingState() uint32 {
	return a.adaptor().drone.GetFlyingState()
}

// GetPictureState return picture state (0 = ready, 1 = busy, 2 = not available)
func (a *AirborneDroneDriver) GetPictureState() uint32 {
	return a.adaptor().drone.GetPictureState()
}

// speed can be a value from `0` to `100`.
func (a *AirborneDroneDriver) Up(speedFactor int8) {
        a.adaptor().drone.Up(speedFactor)
}

// speed can be a value from `0` to `100`.
func (a *AirborneDroneDriver) Down(speedFactor int8) {
        a.adaptor().drone.Down(speedFactor)
}

// speed can be a value from `0` to `100`.
func (a *AirborneDroneDriver) Left(speedFactor int8) {
        a.adaptor().drone.Left(speedFactor)
}

// speed can be a value from `0` to `100`.
func (a *AirborneDroneDriver) Right(speedFactor int8) {
        a.adaptor().drone.Right(speedFactor)
}

// speed can be a value from `0` to `100`.
func (a *AirborneDroneDriver) Forward(speedFactor int8) {
        a.adaptor().drone.Forward(speedFactor)
}

// speed can be a value from `0` to `100`.
func (a *AirborneDroneDriver) Backward(speedFactor int8) {
        a.adaptor().drone.Backward(speedFactor)
}

// speed can be a value from `0` to `100`.
func (a *AirborneDroneDriver) TurnLeft(speedFactor int8) {
        a.adaptor().drone.TurnLeft(speedFactor)
}

// speed can be a value from `0` to `100`.
func (a *AirborneDroneDriver) TurnRight(speedFactor int8) {
        a.adaptor().drone.TurnRight(speedFactor)
}

// Hover makes the drone to hover in place.
func (a *AirborneDroneDriver) Stop() {
	a.adaptor().drone.Hover()
}

// Roll causes the drone to roll  (left -100 - right 100)
func (a *AirborneDroneDriver) Roll(duration time.Duration, speedFactor int8) error {
	return a.adaptor().drone.Roll(duration, speedFactor)
}

// Pitch causes the drone to pitch  (backward -100 - forward 100)
func (a *AirborneDroneDriver) Pitch(duration time.Duration, speedFactor int8) error {
	return a.adaptor().drone.Pitch(duration, speedFactor)
}

// Yaw causes the drone to yaw  (turn left -100 - turn right 100)
func (a *AirborneDroneDriver) Yaw(duration time.Duration, speedFactor int8) error {
	return a.adaptor().drone.Yaw(duration, speedFactor)
}

// Gaz causes the drone to gaz  (down -100 - up 100)
func (a *AirborneDroneDriver) Gaz(duration time.Duration, speedFactor int8) error {
	return a.adaptor().drone.Gaz(duration, speedFactor)
}

// Hover makes the drone to hover in place.
func (a *AirborneDroneDriver) Hover() {
	a.adaptor().drone.Hover()
}

// NewCommander causes the drone to create commander
func (a *AirborneDroneDriver) NewCommander() *client.Commander {
	return a.adaptor().drone.NewCommander()
}

