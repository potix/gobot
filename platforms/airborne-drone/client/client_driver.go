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


// XXXXXXXXXXXXXXXX check follow functions
func (client *Client) Takeoff() bool {
	return true
}

func (client *Client) Land() {
}

func (client *Client) Apply(state State) {
}

func (client *Client) ApplyFor(duration time.Duration, state State) {
}

func (client *Client) Vertical(duration time.Duration, speed float64) {
}

func (client *Client) Roll(duration time.Duration, speed float64) {
}

func (client *Client) Pitch(duration time.Duration, speed float64) {
}

func (client *Client) Yaw(duration time.Duration, speed float64) {
}

func (client *Client) Up(speed float64) {
}

func (client *Client) Down(speed float64) {
}

func (client *Client) Right(speed float64) {
}

func (client *Client) Left(speed float64) {
}

func (client *Client) Clockwise(speed float64) {
}

func (client *Client) Counterclockwise(speed float64) {
}

func (client *Client) Forward(speed float64) {
}

func (client *Client) Backward(speed float64) {
}

func (client *Client) Hover() {
}

