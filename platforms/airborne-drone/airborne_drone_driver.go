package airbornedrone

import (
	"github.com/potix/gobot"
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
func (a *AirborneDroneDriver) TakeOff() {
	gobot.Publish(a.Event("flying"), a.adaptor().drone.Takeoff())
}

// Land causes the drone to land
func (a *AirborneDroneDriver) Land() {
	a.adaptor().drone.Land()
}

// Up makes the drone gain altitude.
// speed can be a value from `0.0` to `1.0`.
func (a *AirborneDroneDriver) Up(speed float64) {
	a.adaptor().drone.Up(speed)
}

// Down makes the drone reduce altitude.
// speed can be a value from `0.0` to `1.0`.
func (a *AirborneDroneDriver) Down(speed float64) {
	a.adaptor().drone.Down(speed)
}

// Left causes the drone to bank to the left, controls the roll, which is
// a horizontal movement using the camera as a reference point.
// speed can be a value from `0.0` to `1.0`.
func (a *AirborneDroneDriver) Left(speed float64) {
	a.adaptor().drone.Left(speed)
}

// Right causes the drone to bank to the right, controls the roll, which is
// a horizontal movement using the camera as a reference point.
// speed can be a value from `0.0` to `1.0`.
func (a *AirborneDroneDriver) Right(speed float64) {
	a.adaptor().drone.Right(speed)
}

// Forward causes the drone go forward, controls the pitch.
// speed can be a value from `0.0` to `1.0`.
func (a *AirborneDroneDriver) Forward(speed float64) {
	a.adaptor().drone.Forward(speed)
}

// Backward causes the drone go backward, controls the pitch.
// speed can be a value from `0.0` to `1.0`.
func (a *AirborneDroneDriver) Backward(speed float64) {
	a.adaptor().drone.Backward(speed)
}

// Clockwise causes the drone to spin in clockwise direction
// speed can be a value from `0.0` to `1.0`.
func (a *AirborneDroneDriver) Clockwise(speed float64) {
	a.adaptor().drone.Clockwise(speed)
}

// CounterClockwise the drone to spin in counter clockwise direction
// speed can be a value from `0.0` to `1.0`.
func (a *AirborneDroneDriver) CounterClockwise(speed float64) {
	a.adaptor().drone.Counterclockwise(speed)
}

// Hover makes the drone to hover in place.
func (a *AirborneDroneDriver) Hover() {
	a.adaptor().drone.Hover()
}
