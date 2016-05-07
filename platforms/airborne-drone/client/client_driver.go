package client

import (
	"errors"
	"fmt"
	"time"
)

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

func (client *Client) SetMaxAltitude(altitude float32) error {
	if altitude < 2.6 || altitude > 10.0 {
		return errors.New("altitude is out of range")
	}
	client.adaptor.SetMaxAltitude(altitude)
	return nil
}

func (client *Client) SetMaxTilt(tilt float32) error {
	if tilt < 5.0 || tilt > 25.0 {
		return errors.New("tilt is out of range")
	}
	client.adaptor.SetMaxTilt(tilt)
	return nil
}

func (client *Client) SetMaxVirticalSpeed(virticalSpeed float32) error {
	if virticalSpeed < 0.5 || virticalSpeed > 2.0 {
		return errors.New("virticalSpeed is out of range")
	}
	client.adaptor.SetMaxVirticalSpeed(virticalSpeed)
	return nil
}

func (client *Client) SetMaxRotationSpeed(rotationSpeed float32) error {
	if rotationSpeed < 0.0 || rotationSpeed > 360 {
		return errors.New("rotationSpeed is out of range")
	}
	client.adaptor.SetMaxRotationSpeed(rotationSpeed)
	return nil
}

func (client *Client) SetContinuousMode(onOff bool) {
	client.adaptor.SetContinuousMode(onOff)
}

func (client *Client) Headlight(left uint8, right uint8) error {
	return client.adaptor.Headlight(left, right)
}

func (client *Client) TakePicture() error {
	return client.adaptor.TakePicture()
}

func (client *Client) GetBattery() uint8 {
	return client.adaptor.GetBattery()
}

func (client *Client) GetFlyingState() uint32 {
	return client.adaptor.GetFlyingState()
}

func (client *Client) GetPictureState() uint32 {
	return client.adaptor.GetPictureState()
}

// roll, pitch, yaw, gaz
func (client *Client) Roll(duration time.Duration, speedFactor int8) error {
	if speedFactor < -100 || speedFactor > 100 {
		return errors.New("speedFactor is out of range")
	}
	tc := int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, speedFactor, 0, 0, 0)
	return nil
}

func (client *Client) Pitch(duration time.Duration, speedFactor int8) error {
	if speedFactor < -100 || speedFactor > 100 {
		return errors.New("speedFactor is out of range")
	}
	tc := int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, 0, speedFactor, 0, 0)
	return nil
}

func (client *Client) Yaw(duration time.Duration, speedFactor int8) error {
	if speedFactor < -100 || speedFactor > 100 {
		return errors.New("speedFactor is out of range")
	}
	tc := int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, 0, 0, speedFactor, 0)
	return nil
}

func (client *Client) Gaz(duration time.Duration, speedFactor int8) error {
	if speedFactor < -100 || speedFactor > 100 {
		return errors.New("speedFactor is out of range")
	}
	tc := int(duration/time.Duration(DriveTick * time.Millisecond))
	client.adaptor.AddDrive(tc, 1, 0, 0, 0, speedFactor)
	return nil
}

func (client *Client) Hover() {
	client.adaptor.AddDrive(1, 0, 0, 0, 0, 0)
}

func (client *Client) Up(speedFactor int8) {
	if speedFactor < 0 {
		speedFactor = 0
	} else if speedFactor > 100 {
		speedFactor = 100
	}
	client.adaptor.AddDrive(1, 1, 0, 0, 0, speedFactor)
}

func (client *Client) Down(speedFactor int8) {
	if speedFactor < 0 {
		speedFactor = 0
	} else if speedFactor > 100 {
		speedFactor = 100
	}
	client.adaptor.AddDrive(1, 1, 0, 0, 0, -speedFactor)
}

func (client *Client) Forward(speedFactor int8) {
	if speedFactor < 0 {
		speedFactor = 0
	} else if speedFactor > 100 {
		speedFactor = 100
	}
	client.adaptor.AddDrive(1, 1, 0, speedFactor, 0, 0)
}

func (client *Client) Backward(speedFactor int8) {
	if speedFactor < 0 {
		speedFactor = 0
	} else if speedFactor > 100 {
		speedFactor = 100
	}
	client.adaptor.AddDrive(1, 1, 0, -speedFactor, 0, 0)
}

func (client *Client) Right(speedFactor int8) {
	if speedFactor < 0 {
		speedFactor = 0
	} else if speedFactor > 100 {
		speedFactor = 100
	}
	client.adaptor.AddDrive(1, 1, speedFactor, 0, 0, 0)
}

func (client *Client) Left(speedFactor int8) {
	if speedFactor < 0 {
		speedFactor = 0
	} else if speedFactor > 100 {
		speedFactor = 100
	}
	client.adaptor.AddDrive(1, 1, -speedFactor, 0, 0, 0)
}

func (client *Client) TurnRight(speedFactor int8) {
	if speedFactor < 0 {
		speedFactor = 0
	} else if speedFactor > 100 {
		speedFactor = 100
	}
	client.adaptor.AddDrive(1, 1, 0, 0, speedFactor, 0)
}

func (client *Client) TurnLeft(speedFactor int8) {
	if speedFactor < 0 {
		speedFactor = 0
	} else if speedFactor > 100 {
		speedFactor = 100
	}
	client.adaptor.AddDrive(1, 1, 0, 0, -speedFactor, 0)
}

func (client *Client) Stop() {
	client.adaptor.AddDrive(1, 0, 0, 0, 0, 0)
}

// command style
type Commander struct {
	driveParam *driveParam
	client *Client
}

func (client *Client) NewCommander() *Commander {
	cmd := new(Commander)
	cmd.driveParam = new(driveParam)
	cmd.client = client
	return cmd
}

func (cmd *Commander) ContinuousMode(onOff bool) *Commander {
        cmd.client.adaptor.SetContinuousMode(onOff)
	return cmd
}

func (cmd *Commander) Headlight(left uint8, right uint8) *Commander {
	cmd.client.adaptor.Headlight(left, right)
	return cmd
}

func (cmd *Commander) TakePicture() *Commander {
	if cmd.client.GetPictureState() == 0 {
		cmd.client.adaptor.TakePicture()
	}
	return cmd
}

func (cmd *Commander) FrontFlip() *Commander {
	cmd.client.adaptor.Flip(0)
	return cmd
}

func (cmd *Commander) BackFlip() *Commander {
	cmd.client.adaptor.Flip(1)
	return cmd
}

func (cmd *Commander) RightFlip() *Commander {
	cmd.client.adaptor.Flip(2)
	return cmd
}

func (cmd *Commander) LeftFlip() *Commander {
	cmd.client.adaptor.Flip(3)
	return cmd
}

func (cmd *Commander) Up(speedFactor uint8) *Commander {
	if speedFactor > 100 {
		speedFactor = 100
	}
	cmd.driveParam.gaz += int8(speedFactor)
	return cmd
}

func (cmd *Commander) Down(speedFactor uint8) *Commander {
	if speedFactor > 100 {
		speedFactor = 100
	}
	cmd.driveParam.gaz -= int8(speedFactor)
	return cmd
}

func (cmd *Commander) Forward(speedFactor uint8) *Commander {
	if speedFactor > 100 {
		speedFactor = 100
	}
	cmd.driveParam.pitch += int8(speedFactor)
	return cmd
}

func (cmd *Commander) Backward(speedFactor uint8) *Commander {
	if speedFactor > 100 {
		speedFactor = 100
	}
	cmd.driveParam.pitch -= int8(speedFactor)
	return cmd
}

func (cmd *Commander) Right(speedFactor uint8) *Commander {
	if speedFactor > 100 {
		speedFactor = 100
	}
	cmd.driveParam.roll += int8(speedFactor)
	return cmd
}

func (cmd *Commander) Left(speedFactor uint8) *Commander {
	if speedFactor > 100 {
		speedFactor = 100
	}
	cmd.driveParam.roll -= int8(speedFactor)
	return cmd
}

func (cmd *Commander) TurnRight(speedFactor uint8) *Commander {
	if speedFactor > 100 {
		speedFactor = 100
	}
	cmd.driveParam.yaw += int8(speedFactor)
	return cmd
}

func (cmd *Commander) TurnLeft(speedFactor uint8) *Commander {
	if speedFactor > 100 {
		speedFactor = 100
	}
	cmd.driveParam.yaw -= int8(speedFactor)
	return cmd
}

func (cmd *Commander) Stop() *Commander {
	cmd.driveParam.flag = 0
	cmd.driveParam.roll = 0
	cmd.driveParam.pitch = 0
	cmd.driveParam.yaw = 0
	cmd.driveParam.gaz = 0
	cmd.client.adaptor.AddDrive(1, cmd.driveParam.flag, cmd.driveParam.roll, cmd.driveParam.pitch, cmd.driveParam.yaw, cmd.driveParam.gaz)
	return cmd
}

func (cmd *Commander) Do(duration time.Duration) *Commander {
	tc := int(duration/time.Duration(DriveTick * time.Millisecond))
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
	cmd.client.adaptor.AddDrive(tc, cmd.driveParam.flag, cmd.driveParam.roll, cmd.driveParam.pitch, cmd.driveParam.yaw, cmd.driveParam.gaz)
	time.Sleep(time.Duration(tc) * DriveTick * time.Millisecond)
	return cmd
}

func (cmd *Commander) Reset() *Commander {
	cmd.driveParam.flag = 0
	cmd.driveParam.roll = 0
	cmd.driveParam.pitch = 0
	cmd.driveParam.yaw = 0
	cmd.driveParam.gaz = 0
	return cmd
}


//func (client *Client) Clockwise(speed float64) {
//}

//func (client *Client) Counterclockwise(speed float64) {
//}

