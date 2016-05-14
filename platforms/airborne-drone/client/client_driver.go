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
	DownloadPath string
}

func DefaultConfig() Config {
	return Config{
		DroneAddress:   "00:00:00:00:00:00",
		DownloadPath:   "/var/tmp/airborn_drone",
	}
}

func Connect(config Config) (*Client, error) {
	client := &Client{Config: config}
	client.adaptor = NewAdaptor("ble", client.Config.DroneAddress, client.Config.DownloadPath)
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

func (client *Client) SetAutoDownloadMode(onOff bool) {
	client.adaptor.SetAutoDownloadMode(onOff)
}

func (client *Client) SetCutOutMode(onOff bool) error {
	return client.adaptor.SetCutOutMode(onOff)
}

func (client *Client) ForceDownload() {
	client.adaptor.ForceDownload()
}

func (client *Client) FlatTrim() error {
	return client.adaptor.FlatTrim()
}

func (client *Client) Emergency() error {
	return client.adaptor.Emergency()
}

func (client *Client) Headlight(left uint8, right uint8) error {
	return client.adaptor.Headlight(left, right)
}

func (client *Client) HeadlightFlashStart() error {
	return client.adaptor.HeadlightFlashStart()
}

func (client *Client) HeadlightBlinkStart() error {
	return client.adaptor.HeadlightBlinkStart()
}

func (client *Client) HeadlightOscillationStart() error {
	return client.adaptor.HeadlightOscillationStart()
}

func (client *Client) HeadlightFlashStop() error {
	return client.adaptor.HeadlightFlashStop()
}

func (client *Client) HeadlightBlinkStop() error {
	return client.adaptor.HeadlightBlinkStop()
}

func (client *Client) HeadlightOscillationStop() error {
	return client.adaptor.HeadlightOscillationStop()
}

func (client *Client) TakePicture() error {
	return client.adaptor.TakePicture()
}

func (client *Client) FTPList(path string) ([]byte, error) {
	return client.adaptor.FTPList(path)
}

func (client *Client) FTPGet(path string) ([]byte, error) {
	return client.adaptor.FTPGet(path)
}

func (client *Client) FTPDelete(path string) ([]byte, error) {
	return client.adaptor.FTPDelete(path)
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

// Multiplexer style
type Multiplexer struct {
	driveParam *driveParam
	client *Client
}

func (client *Client) NewMultiplexer() *Multiplexer {
	m := new(Multiplexer)
	m.driveParam = new(driveParam)
	m.client = client
	return m
}

func (m *Multiplexer) Up(speedFactor uint8) *Multiplexer {
	if speedFactor > 100 {
		speedFactor = 100
	}
	m.driveParam.gaz += int8(speedFactor)
	return m
}

func (m *Multiplexer) Down(speedFactor uint8) *Multiplexer {
	if speedFactor > 100 {
		speedFactor = 100
	}
	m.driveParam.gaz -= int8(speedFactor)
	return m
}

func (m *Multiplexer) Forward(speedFactor uint8) *Multiplexer {
	if speedFactor > 100 {
		speedFactor = 100
	}
	m.driveParam.pitch += int8(speedFactor)
	return m
}

func (m *Multiplexer) Backward(speedFactor uint8) *Multiplexer {
	if speedFactor > 100 {
		speedFactor = 100
	}
	m.driveParam.pitch -= int8(speedFactor)
	return m
}

func (m *Multiplexer) Right(speedFactor uint8) *Multiplexer {
	if speedFactor > 100 {
		speedFactor = 100
	}
	m.driveParam.roll += int8(speedFactor)
	return m
}

func (m *Multiplexer) Left(speedFactor uint8) *Multiplexer {
	if speedFactor > 100 {
		speedFactor = 100
	}
	m.driveParam.roll -= int8(speedFactor)
	return m
}

func (m *Multiplexer) TurnRight(speedFactor uint8) *Multiplexer {
	if speedFactor > 100 {
		speedFactor = 100
	}
	m.driveParam.yaw += int8(speedFactor)
	return m
}

func (m *Multiplexer) TurnLeft(speedFactor uint8) *Multiplexer {
	if speedFactor > 100 {
		speedFactor = 100
	}
	m.driveParam.yaw -= int8(speedFactor)
	return m
}

func (m *Multiplexer) Reset() *Multiplexer {
	m.driveParam.flag = 0
	m.driveParam.roll = 0
	m.driveParam.pitch = 0
	m.driveParam.yaw = 0
	m.driveParam.gaz = 0
	return m
}

func (m *Multiplexer) Go(duration time.Duration) *Multiplexer {
	tc := int(duration/time.Duration(DriveTick * time.Millisecond))
	if m.driveParam.roll < -100 {
		m.driveParam.roll = -100
	} else if m.driveParam.roll > 100{
		m.driveParam.roll = 100
	}
	if m.driveParam.pitch < -100 {
		m.driveParam.pitch = -100
	} else if m.driveParam.pitch > 100{
		m.driveParam.pitch = 100
	}
	if m.driveParam.yaw < -100 {
		m.driveParam.yaw = -100
	} else if m.driveParam.yaw > 100{
		m.driveParam.yaw = 100
	}
	if m.driveParam.gaz < -100 {
		m.driveParam.gaz = -100
	} else if m.driveParam.gaz > 100{
		m.driveParam.gaz = 100
	}
	m.driveParam.flag = 1
	m.client.adaptor.AddDrive(tc, m.driveParam.flag, m.driveParam.roll, m.driveParam.pitch, m.driveParam.yaw, m.driveParam.gaz)
	time.Sleep(time.Duration(tc) * DriveTick * time.Millisecond)
	return m
}
