package client

import (
	"fmt"
	"log"
	"strings"
        "time"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/airborne-drone/client/service"
	"github.com/potix/gatt"
)

// drive tick times (millisecond)
const DriveTick = 25

// TODO: handle other OS defaults besides Linux
var DefaultClientOptions = []gatt.Option{
	gatt.LnxMaxConnections(1),
	gatt.LnxDeviceID(-1, false),
}

var _ gobot.Adaptor = (*Adaptor)(nil)

type DriveParam struct {
	pcmd  uint8
	flag  uint8
	roll  int8
	pitch int8
	yaw   int8
	gaz   int8
}

// Represents a Connection to a BLE Peripheral
type Adaptor struct {
	name            string
	uuid            string
	device          gatt.Device
	peripheral      gatt.Peripheral
	state		gatt.State
	services	map[string]*BLEService
	connected       bool
	peripheralReady	chan error
	seq		map[uint16]uint8
	pcmdLoopEnd	chan bool
	driveParamMutex  *sync.Mutex
	driveParam      []interface{}
	continuousMode  bool

}

// NewAdaptor returns a new Adaptor given a name and uuid
func NewAdaptor(name string, uuid string) *Adaptor {
	a := &Adaptor{
		name:      name,
		uuid:      uuid,
		connected: false,
		peripheralReady: make(chan error),
		services: make(map[string]*BLEService),
		seq: make(map[uint16]uint8),
		driveLoopEnd: make(chan bool),
		driveParamMutex: New(*sync.Mutex),
		driveParam: make([]interface{}),
		continuousMode: false,
	}
	a.seq[0xfa0a] = 0
	a.seq[0xfa0b] = 0
	a.seq[0xfa0c] = 0
	a.seq[0xfa1e] = 0
	a.seq[0xfd23] = 0
	a.seq[0xfd24] = 0
	return a
}

func (b *Adaptor) Name() string                { return b.name }
func (b *Adaptor) UUID() string                { return b.uuid }
func (b *Adaptor) Peripheral() gatt.Peripheral { return b.peripheral }
func (b *Adaptor) State() gatt.State           { return b.state }

// Connect initiates a connection to the BLE peripheral. Returns true on successful connection.
func (b *Adaptor) Connect() (errs []error) {
	var err error
	errs = make([]error, 1)

	device, err := gatt.NewDevice(DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open BLE device, err: %s\n", err)
		errs[0] = err
		return errs
	}
	device.Option(DefaultClientOptions...)

	b.device = device

	// Register handlers.
	device.Handle(
		gatt.PeripheralDiscovered(b.onPeripheralDiscovered),
		gatt.PeripheralConnected(b.onPeripheralConnected),
		gatt.PeripheralDisconnected(b.onPeripheralDisconnected),
	)

	device.Init(b.onStateChanged)

	// Peripheral ready
	if err = <-b.peripheralReady; err != nil {
		log.Fatalf("Failed to open BLE device, err: %s\n", err)
		errs[0] = err
		return errs
	}
	close(b.peripheralReady)

	// connected
	fmt.Println("connected")
	b.connected = true

	// setup
	if err := b.setMTU(64); err != nil {
		fmt.Println("disconnect")
		b.Disconnect()
		b.connected = false
		log.Fatalf("Failed to open BLE device, err: %s\n", err)
		errs[0] = err
		return errs
	}
	if err := b.discoveryService(); err != nil {
		fmt.Println("disconnect")
		b.Disconnect()
		b.connected = false
		log.Fatalf("Failed to open BLE device, err: %s\n", err)
		errs[0] = err
		return errs
	}

	go b.driveLoop()

	return nil
}

func (b *Adaptor) TakeOff() error {
	c := b.services["9a66fa000800919111e4012d1540cb8e"].characteristics["9a66fa0b0800919111e4012d1540cb8e"]
	value := make([]byte, 0, 16)
	append(value, 0x04 /*type*/, b.seq[0xfa0b] /*seq*/, 0x02 /*prjid*/, 0x00 /*clsid*/, 0x01, 0x00 /*cmdid 2byte*/)
	err := b.peripheral.WriteCharacteristic(c, value[:6], false)
	b.seq[0xfa0b] += 1
	if err {
		return err
	}
	return nil
}

func (b *Adaptor) Landing() error {
	c := b.services["9a66fa000800919111e4012d1540cb8e"].characteristics["9a66fa0b0800919111e4012d1540cb8e"]
	value := make([]byte, 0, 16)
	append(value, 0x04 /*type*/, b.seq[0xfa0b] /*seq*/, 0x02 /*prjid*/, 0x00 /*clsid*/, 0x03, 0x00 /*cmdid 2byte*/)
	err := b.peripheral.WriteCharacteristic(c, value[:6], false)
	b.seq[0xfa0b] += 1
	if err {
		return err
	}
	return nil
}

func (b *Adaptor) Flip(uint32 v) error {
	c := b.services["9a66fa000800919111e4012d1540cb8e"].characteristics["9a66fa0b0800919111e4012d1540cb8e"]
	value := make([]byte, 0, 16)
	append(value, 0x04 /*type*/, b.seq[0xfa0b] /*seq*/, 0x02 /*prjid*/, 0x04 /*clsid*/, 0x00, 0x00 /*cmdid 2byte*/)
	binary.LittleEndian.PutUint16(value[6:10], v)
	err := b.peripheral.WriteCharacteristic(c, value[:10], false)
	b.seq[0xfa0b] += 1
	if err {
		return err
	}
	return nil
}

/// XXXXXXXXXXXXXXXXXXXXXXX
func (client *Client) SetMaxAltitude(altitude float) error {
        if altitude < 2.6 || altitude > 10.0 {
                return errors.New("altitude is out of range")
        }
        client.adaptor.SetMaxAltitude(altitude)
}

func (client *Client) SetMaxTilt(tilt float) error {
        if tilt < 5.0 || tilt > 25.0 {
                return Errors.New("tilt is out of range")
        }
        client.adaptor.SetMaxTilt(tilt)
}

func (client *Client) SetMaxVirticalSpeed(virticalSpeed float) error {
        if virticalSpeed < 0.5 || virticalSpeed > 2.0 {
                return Errors.New("virticalSpeed is out of range")
        }
        client.adaptor.SetMaxVirticalSpeed(virticalSpeed)
}

func (client *Client) SetMaxRotationSpeed(rotationSpeed float) error {
        if rotationSpeed < 0.0 || rotationSpeed > 360 {
                return Errors.New("rotationSpeed is out of range")
        }
        client.adaptor.SetMaxRotationSpeed(rotationSpeed)
}
/// XXXXXXXXXXXXXXXXXXXX

func (b *Adaptor) AddDrive(int tickCnt, flag uint8, roll int8, pitch int8, yaw int8, gaz int8) {
	for i := 0; i < tickCnt; i++ {
		dp := &DriveParam {
			pcmd = 1,
			flag = flag,
			roll = roll,
			pitch = pitch,
			yaw = yaw,
			gaz = gaz,
		}
		b.appendDriveParam(dp)
	}
}

func (b *Adaptor) SetContinuousMode(continuousMode bool) {
	b.driveParamMutex.Lock()
	defer b.driveParamMutex.Unlock()
	b.continuousMode = continuousMode
}

func (b *Adaptor) takeDriveParam(lastDP *DriveParam) (*DriveParam, bool) {
	b.driveParamMutex.Lock()
	defer b.driveParamMutex.Unlock()
	if l = len(b.driveParam); l > 0 {
		// return new drive param
		b.driveParam = b.driveParam[1:len(b.driveParam)]
		return b.driveParam[0]
	} else {
		if b.continuousMode {
			// last param retry
			return lastDP
		} else {
			// initialize (hover)
			return &DriveParam {
				pcmd = 1,
				flag = 0,
				roll = 0,
				pitch = 0,
				yaw = 0,
				gaz = 0,
			}
		}
	}
}

func (b *Adaptor) appendDriveParam(driveParam *DriveParam) {
	b.driveParamMutex.Lock()
	defer b.driveParamMutex.Unlock()
	b.driveParam = append(b.driveParam, driveParam)
}

func (b *Adaptor) driveLoop() {
	dp := &DriveParam {
		pcmd = 1,
		flag = 0,
		roll = 0,
		pitch = 0,
		yaw = 0,
		gaz = 0,
	}
	now := time.Now()
	ticker := time.NewTicker(DriveTick * time.Millisecond)
	loop:
	for {
		select {
		case t := <-ticker.C:
			dp = b.takeDriveParam(dp)
			if (dp.pcmd) {
				millisec := uint32(t.Sub(now).Seconds() * 1000)
				c := b.services["9a66fa000800919111e4012d1540cb8e"].characteristics["9a66fa0a0800919111e4012d1540cb8e"]
				value := make([]byte, 0, 32)
				append(value, 0x02 /*type*/, b.seq[0xfa0a] /*seq*/, 0x02 /*prjid*/, 0x00 /*clsid*/, 0x02, 0x00 /*cmdid 2byte*/)
				append(value, dp.flag, dp.roll, dp.pitch, dp.yaw, dp.gaz)
				binary.LittleEndian.PutUint16(value[11:15], millisec)
				err := b.peripheral.WriteCharacteristic(c, value[:15], false)
				b.seq[0xfa0a] += 1
				if err {
					fmt.Println(err)
				}
			}
		case <-b.driveLoopEnd:
			break loop
		}
	}
	ticker.Stop()
}

// Reconnect attempts to reconnect to the BLE peripheral. If it has an active connection
// it will first close that connection and then establish a new connection.
// Returns true on Successful reconnection
func (b *Adaptor) Reconnect() (errs []error) {
	if b.connected {
		b.Disconnect()
	}
	return b.Connect()
}

// Disconnect terminates the connection to the BLE peripheral. Returns true on successful disconnect.
func (b *Adaptor) Disconnect() (errs []error) {
	b.driveLoopEnd <- true
	b.peripheral.Device().CancelConnection(b.peripheral)
	return
}

// Finalize finalizes the Adaptor
func (b *Adaptor) Finalize() (errs []error) {
	return b.Disconnect()
}

// ReadCharacteristic returns bytes from the BLE device for the
// requested service and characteristic
func (b *Adaptor) ReadCharacteristic(sUUID string, cUUID string) (data []byte, err error) {
        if !b.connected {
                log.Fatalf("Cannot read from BLE device until connected")
                return
        }

        characteristic := b.services[sUUID].characteristics[cUUID]
        val, err := b.peripheral.ReadCharacteristic(characteristic)
        if err != nil {
                fmt.Printf("Failed to read characteristic, err: %s\n", err)
                return  nil, err
        }

         return val, nil
}

func (b *Adaptor) onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	b.state = s
	switch s {
	case gatt.StatePoweredOn:
		// set service
		d.AddService(service.NewGattGapService())
		// start scan
		fmt.Println("scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func (b *Adaptor) onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	// check device uuid
	id := strings.ToUpper(b.UUID())
	if strings.ToUpper(p.ID()) != id {
		return
	}

	// check device name
	if !strings.HasPrefix(p.Name(), "Swat_") {
		return
	}

        ms := fmt.Sprintf("%x", a.ManufacturerData)
        fmt.Printf("Manufacturer = %s\n", ms)
	if ms != "4300cf1907090100" {
		return
	}

	// Stop scanning once we've got the peripheral we're looking for.
	p.Device().StopScanning()

	// and connect to it
	p.Device().Connect(p)
}

func (b *Adaptor) onPeripheralConnected(p gatt.Peripheral, err error) {
	fmt.Printf("\nConnected Peripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
	b.peripheral = p
	if err != nil {
		b.peripheralReady <- err
	} else {
		b.peripheralReady <- nil
	}
}

func (b *Adaptor) onPeripheralDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Disconnected")
}

func (b *Adaptor) setMTU(mtu uint16) error {
	fmt.Println("setMTU")
	if err := b.peripheral.SetMTU(mtu); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
		return err
	}
	return nil
}

func (b *Adaptor) discoveryService() error {
	fmt.Println("discoveryService")
	ss, err := b.peripheral.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		return err
	}

	for _, s := range ss {
		b.services[s.UUID().String()] = NewBLEService(s.UUID().String(), s)

		cs, err := b.peripheral.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			b.services[s.UUID().String()].characteristics[c.UUID().String()] = c
		}
	}

	// add service
	if s, ok := b.services["9a66fb000800919111e4012d1540cb8e"]; ok {
		if c, ok := s.characteristics["9a66fb0f0800919111e4012d1540cb8e"]; ok {
			// notify (request with response on arnetwork)
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-REQ fb0f-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fb0e0800919111e4012d1540cb8e"]; ok {
			// notify (request with no response on arnetwork)
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-REQ fb0e-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fb1b0800919111e4012d1540cb8e"]; ok {
			// notify (response on arnetwork)
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-RES fb1b-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fb1c0800919111e4012d1540cb8e"]; ok {
			// notify (low latency response on arnetwork)
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-RES fb1c-")
				fmt.Println(b)
			})
		}
	}
	if s, ok := b.services["9a66fd210800919111e4012d1540cb8e"]; ok {
		if c, ok := s.characteristics["9a66fd220800919111e4012d1540cb8e"]; ok {
			// ????
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-??? fd22-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fd230800919111e4012d1540cb8e"]; ok {
			// notify (ftp data transfer)
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-FTP DATA fd23-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fd240800919111e4012d1540cb8e"]; ok {
			// notify (ftp control)
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-FTP CNTRL fd24-")
				fmt.Println(b)
			})
		}
	}
	if s, ok := b.services["9a66fd510800919111e4012d1540cb8e"]; ok {
		if c, ok := s.characteristics["9a66fd520800919111e4012d1540cb8e"]; ok {
			// ????
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-??? fd52-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fd530800919111e4012d1540cb8e"]; ok {
			// ????
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-??? fd53-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fd540800919111e4012d1540cb8e"]; ok {
			// ????
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-??? fd54-")
				fmt.Println(b)
			})
		}
	}

	return nil
}

// Represents a BLE Peripheral's Service
type BLEService struct {
	uuid            string
	service         *gatt.Service
	characteristics map[string]*gatt.Characteristic
}

// NewAdaptor returns a new BLEService given a uuid
func NewBLEService(sUuid string, service *gatt.Service) *BLEService {
	return &BLEService{
		uuid:            sUuid,
		service:         service,
		characteristics: make(map[string]*gatt.Characteristic),
	}
}
