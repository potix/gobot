package client

import (
	"errors"
	"fmt"
	"time"
)

// XXXXXXXXXXXXXXXXXX check
//type State struct {
//	Fly       bool     // Set to false for landing.
//	Emergency bool     // Used to disable / trigger emergency mode
//	Config    []KeyVal // Config values to send
}

// XXXXXXXXXXXXXXX check
//type KeyVal struct {
//	Key   string
//	Value string
//}

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
	err := client.adaptor.TakeOff()
	if err != nil {
		return false
	}
	return true
}

func (client *Client) Landing() error {
	return client.adaptor.Landing()
}

func (client *Client) FrontFlip() error {
	return client.adaptor.Flip(0)
}

func (client *Client) BackFlip() error {
	return client.adaptor.Flip(1)
}

func (client *Client) RightFlip() error {
	return client.adaptor.Flip(2)
}

func (client *Client) LeftFlip() error {
	return client.adaptor.Flip(3)
}

// XXXX check follow functions
//func (client *Client) Apply(state State) {
//}

// XXXX check follow functions
//func (client *Client) ApplyFor(duration time.Duration, state State) {
//}

func (client *Client) SetMaxAltitude(altitude float) error {
	if altitude < 2.6 || altitude > 10.0 {
		return errors.New("altitude is out of range")
	}
	client.adaptor.SetMaxAltitude(altitude)
	return nil
}

func (client *Client) SetMaxTilt(tilt float) error {
	if tilt < 5.0 || tilt > 25.0 {
		return Errors.New("tilt is out of range")
	}
	client.adaptor.SetMaxTilt(tilt)
	return nil
}

func (client *Client) SetMaxVirticalSpeed(virticalSpeed float) error {
	if virticalSpeed < 0.5 || virticalSpeed > 2.0 {
		return Errors.New("virticalSpeed is out of range")
	}
	client.adaptor.SetMaxVirticalSpeed(virticalSpeed)
	return nil
}

func (client *Client) SetMaxRotationSpeed(rotationSpeed float) error {
	if rotationSpeed < 0.0 || rotationSpeed > 360 {
		return errors.New("rotationSpeed is out of range")
	}
	client.adaptor.SetMaxRotationSpeed(rotationSpeed)
	return nil
}

func (client *Client) SetContinuousMode(onOff bool) {
	client.adaptor.SetContinuousMode(onOff)
}

// roll, pitch, yaw, gaz
func (client *Client) Roll(duration time.Duration, speedFactor int8) error {
	if speedFactor < -100 || speedFactor > 100 {
		return errors.New("speedFactor is out of range")
	}
	tc = int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, speedFactor, 0, 0, 0)
	return nil
}

func (client *Client) Pitch(duration time.Duration, speedFactor int8) error {
	if speedFactor < -100 || speedFactor > 100 {
		return errors.New("speedFactor is out of range")
	}
	tc = int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, 0, speedFactor, 0, 0)
	return nil
}

func (client *Client) Yaw(duration time.Duration, speedFactor int8) error {
	if speedFactor < -100 || speedFactor > 100 {
		return errors.New("speedFactor is out of range")
	}
	tc = int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, 0, 0, speedFactor, 0)
	return nil
}

func (client *Client) Gaz(duration time.Duration, speedFactor int8) error {
	if speedFactor < -100 || speedFactor > 100 {
		return errors.New("speedFactor is out of range")
	}
	tc = int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, 0, 0, 0 speedFactor)
	return nil
}

// command style
//   example:
//     cmd = client.NewCommander()
//     time.Sleep(time.Duration(1 * time.Second)
//     cmd.Up(50).Right(50).Do(time.Duration(500 * time.Millisecond))
//     time.Sleep(time.Duration(1 * time.Second)
//     cmd.Left(100).Do(time.Duration(500 * time.Millisecond))
//     time.Sleep(time.Duration(1 * time.Second)
//     cmd.Down(50).Right(50).Do(time.Duration(500 * time.Millisecond))
//     time.Sleep(time.Duration(1 * time.Second)
type Commander struct {
	driveParam *DriveParam
	client *Client
}

func (client *Client) NewCommander() *Commander {
	cmd := New(Commander)
	cmd.driveParam = New(DriveParam)
	cmd.client = client
	return cmd
}

func (cmd *Commander) Up(speedFactor uint8) *Commander {
	cmd.driveParam.gaz += speedFactor
	return cmd
}

func (cmd *Commander) Down(speedFactor uint8) *Commander {
	cmd.driveParam.gaz -= speedFactor
	return cmd
}

func (cmd *Commander) Forward(speedFactor uint8) *Commander {
	cmd.driveParam.pitch += speedFactor
	return cmd
}

func (cmd *Commander) Backward(speedFactor uint8) *Commander {
	cmd.driveParam.pitch -= speedFactor
	return cmd
}

func (cmd *Commander) Right(speedFactor uint8) *Commander {
	cmd.driveParam.roll += speedFactor
	return cmd
}

func (cmd *Commander) Left(speedFactor uint8) *Commander {
	cmd.driveParam.roll -= speedFactor
	return cmd
}

func (cmd *Commander) TurnRight(speedFactor uint8) *Commander {
	cmd.driveParam.yaw += speedFactor
	return cmd
}

func (cmd *Commander) TurnLeft(speedFactor uint8) *Commander {
	cmd.driveParam.yaw -= speedFactor
	return cmd
}

func (cmd *Commander) Stop() *Commander {
	cmd.driveParam.flag = 0
	cmd.driveParam.roll = 0
	cmd.driveParam.pitch = 0
	cmd.driveParam.yaw = 0
	cmd.driveParam.gaz = 0
	client.adaptor.AddDrive(1, cmd.driveParam.flag, cmd.driveParam.roll, cmd.driveParam.pitch, cmd.driveParam.yaw, cmd.driveParam.gaz)
	return cmd
}

func (cmd *Commander) Do(duration time.Duration) *Commander {
	tc = int(duration/time.Duration(DriveTick * time.Millisecond))
	if cmd.driveParam.roll < -100 {
		cmd.driveParam.roll = -100
	} else if cmd.driveParam.roll > 100{
		cmd.driveParam.roll = 100
	}
	if cmd.driveParam.pitch < -100 {
		cmd.driveParam.pitch = -100
	} else if cmd.driveParam.pitch > 100{
		cmd.driveParam.pitch = 100
	}
	if cmd.driveParam.yaw < -100 {
		cmd.driveParam.yaw = -100
	} else if cmd.driveParam.yaw > 100{
		cmd.driveParam.yaw = 100
	}
	if cmd.driveParam.gaz < -100 {
		cmd.driveParam.gaz = -100
	} else if cmd.driveParam.gaz > 100{
		cmd.driveParam.gaz = 100
	}
	cmd.driveParam.flag = 1
	client.adaptor.AddDrive(tc, cmd.driveParam.flag, cmd.driveParam.roll, cmd.driveParam.pitch, cmd.driveParam.yaw, cmd.driveParam.gaz)
	return cmd
}

//func (client *Client) Clockwise(speed float64) {
//}

//func (client *Client) Counterclockwise(speed float64) {
//}

func (client *Client) Hover() {
	client.adaptor.AddDrive(1 /*tick count*/, 0 /*flag*/, 0 /*roll*/, 0 /*pitch*/, 0 /*yaw*/, 0 /*gaz*/)
}

