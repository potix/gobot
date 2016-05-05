package client

import (
	"errors"
	"fmt"
	"time"
)

// XXXXXXXXXXXXXXXXXX check
type State struct {
	Pitch     float64  // -1 = max back, 1 = max forward
	Roll      float64  // -1 = max left, 1 = max right
	Yaw       float64  // -1 = max counter clockwise, 1 = max clockwise
	Vertical  float64  // -1 = max down, 1 = max up
	Fly       bool     // Set to false for landing.
	Emergency bool     // Used to disable / trigger emergency mode
	Config    []KeyVal // Config values to send
}

// XXXXXXXXXXXXXXX check
type KeyVal struct {
	Key   string
	Value string
}

type Client struct {
	Config  Config
	adaptor *Adaptor
}

type Config struct {
	DroneAddress string
}

func DefaultConfig() Config {
	return Config{
		DroneAddress:   "00:00:00:00:00:00",
	}
}

func Connect(config Config) (*Client, error) {
	client := &Client{Config: config}
	client.adaptor = NewAdaptor("ble", client.Config.DroneAddress)
	return client, client.Connect()
}

func (client *Client) Connect() error  {
	// BLE connect
	if errs := client.adaptor.Connect(); errs != nil {
		for err := range errs {
			fmt.Println(err)
		}
		return errors.New("cloud not connect")
	}


//	client.navdataConn = navdataConn
//	client.navdataConn.SetReadTimeout(client.Config.NavdataTimeout)
//
//
//	client.controlConn = controlConn
//	client.commands = &commands.Sequence{}
//
//	client.Navdata = make(chan *navdata.Navdata, 0)
//
//	go client.sendLoop()
//	go client.navdataLoop()
//
//	// disable emergency mode (if on) and request demo navdata from drone.
//	for {
//		data := <-client.Navdata
//
//		state := State{}
//		// Sets emergency state if we are in an emergency (which disables it)
//		state.Emergency = data.Header.State&navdata.STATE_EMERGENCY_LANDING != 0
//
//		// Request demo navdata if we are not receiving it yet
//		if data.Demo == nil {
//			state.Config = []KeyVal{{Key: "general:navdata_demo", Value: "TRUE"}}
//		} else {
//			state.Config = []KeyVal{}
//		}
//
//		client.Apply(state)
//
//		// Once emergency is disabled and full navdata is being sent, we are done
//		if !state.Emergency && data.Demo != nil {
//			break
//		}
//	}

	return nil
}


func (client *Client) Takeoff() bool {
	client.adaptor.TakeOff()
	return true
}

func (client *Client) Landing() {
	client.adaptor.Landing()
}

func (client *Client) FrontFlip() {
	client.adaptor.Flip(0)
}

func (client *Client) BackFlip() {
	client.adaptor.Flip(1)
}

func (client *Client) RightFlip() {
	client.adaptor.Flip(2)
}

func (client *Client) LeftFlip() {
	client.adaptor.Flip(3)
}

// XXXX check follow functions
//func (client *Client) Apply(state State) {
//}

// XXXX check follow functions
//func (client *Client) ApplyFor(duration time.Duration, state State) {
//}

// XXXX check follow functions
//func (client *Client) Vertical(duration time.Duration, speed float64) {
//}

func (client *Client) SetAltitude(altitude float) error {
	if altitude < 2.6 || altitude > 10.0 {
		return errors.New("altitude is out of range")
	}
	client.adaptor.SetAltitude(altitude)
}

func (client *Client) SetTilt(tilt float) error {
	if tilt < 5.0 || tilt > 25.0 {
		return Errors.New("tilt is out of range")
	}
	client.adaptor.SetTilt(tilt)
}

func (client *Client) SetVirticalSpeed(virticalSpeed float) error {
	if virticalSpeed < 0.5 || virticalSpeed > 2.0 {
		return Errors.New("virticalSpeed is out of range")
	}
	client.adaptor.SetVirticalSpeed(virticalSpeed)
}

func (client *Client) SetRotationSpeed(rotationSpeed float) error {
	if rotationSpeed < 0.0 || rotationSpeed > 360 {
		return Errors.New("rotationSpeed is out of range")
	}
	client.adaptor.SetRotationSpeed(rotationSpeed)
}

func (client *Client) SetContinuousMode(onOff uint8) {
	client.adaptor.SetContinuousMode(onOff)
}

// roll, pitch, yaw, gaz
func (client *Client) Roll(duration time.Duration, speedFactor int8) {
	tc = int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, speedFactor, 0, 0, 0)
}

func (client *Client) Pitch(duration time.Duration, speedFactor int8) {
	tc = int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, 0, speedFactor, 0, 0)
}

func (client *Client) Yaw(duration time.Duration, speedFactor int8) {
	tc = int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, 0, 0, speedFactor, 0)
}

func (client *Client) Gaz(duration time.Duration, speedFactor int8) {
	tc = int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, 0, 0, 0 speedFactor)
}






// command mode
func (client *Client) NewCommand(duration time.Duration, speed float64) {
}
func (client *Client) Up(duration time.Duration, speed float64, ) {
}

func (client *Client) Down(duration time.Duration, speed float64) {
}

func (client *Client) Forward(duration time.Duration,speed float64) {
}

func (client *Client) Backward(speed float64) {
}

func (client *Client) Right(speed float64) {
}

func (client *Client) Left(speed float64) {
}

func (client *Client) TurnRight(speed float64) {
}

func (client *Client) TurnLeft(speed float64) {
}





//func (client *Client) Clockwise(speed float64) {
//}

//func (client *Client) Counterclockwise(speed float64) {
//}

func (client *Client) Hover() {
	client.adaptor.AddDrive(1 /*tick count*/, 0 /*flag*/, 0 /*roll*/, 0 /*pitch*/, 0 /*yaw*/, 0 /*gaz*/)
}

