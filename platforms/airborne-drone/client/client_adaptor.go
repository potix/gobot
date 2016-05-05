package client

import (
	"fmt"
	"log"
	"strings"
        "time"
        "sync"
        "math"
        "errors"
	"encoding/binary"

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
	pcmd  bool
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
	driveLoopEnd	chan bool
	driveParamMutex  *sync.Mutex
	driveParam      []*DriveParam
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
		driveParamMutex: new(sync.Mutex),
		driveParam: make([]*DriveParam, 0, 0),
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
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x00, 0x01, nil, 6)
}

func (b *Adaptor) Landing() error {
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x00, 0x03, nil, 6)
}

func (b *Adaptor) Flip(value uint32) error {
	data := make([]byte, 0, 4)
	binary.LittleEndian.PutUint32(data[0:4], value)
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x04, 0x00, data, 10)
}

func (b *Adaptor) SetMaxAltitude(altitude float32) error {
	data := make([]byte, 0, 4)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(altitude))
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x08, 0x00, data, 10)
}

func (b *Adaptor) SetMaxTilt(tilt float32) error {
	data := make([]byte, 0, 4)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(tilt))
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x08, 0x01, data, 10)
}

func (b *Adaptor) SetMaxVirticalSpeed(virticalSpeed float32) error {
	data := make([]byte, 0, 4)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(virticalSpeed))
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x01, 0x00, data, 10)
}

func (b *Adaptor) SetMaxRotationSpeed(rotationSpeed float32) error {
	data := make([]byte, 0, 4)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(rotationSpeed))
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x01, 0x01, data, 10)
}

func (b *Adaptor) AddDrive(tickCnt int, flag uint8, roll int8, pitch int8, yaw int8, gaz int8) {
	for i := 0; i < tickCnt; i++ {
		dp := &DriveParam {
			pcmd: true,
			flag: flag,
			roll: roll,
			pitch: pitch,
			yaw: yaw,
			gaz: gaz,
		}
		b.appendDriveParam(dp)
	}
}

func (b *Adaptor) SetContinuousMode(continuousMode bool) {
	b.driveParamMutex.Lock()
	defer b.driveParamMutex.Unlock()
	b.continuousMode = continuousMode
}

func (b *Adaptor) writeCharBase(srvid string, charid string, reqtype uint8, seqid uint16, prjid uint8, clsid uint8, cmdid uint16, data []byte, size int) error {
	var ok bool;
	var s *BLEService;
	var c *gatt.Characteristic;
	if s, ok = b.services[srvid]; !ok {
		return errors.New("not found service")
	}
	if c, ok = s.characteristics[charid]; !ok {
		return errors.New("not found characteristic")
	}
	fmt.Printf("char = %s\n", c.UUID().String())
	value := make([]byte, 0, size)
	value = append(value, reqtype, b.seq[seqid], prjid, clsid)
	binary.LittleEndian.PutUint16(value[4:6], cmdid)
	if data != nil {
		value = append(value[:6], data...)
	}
	fmt.Printf("write = %02x\n", value[:size])
	err := b.peripheral.WriteCharacteristic(c, value[:size], true)
	b.seq[seqid] += 1
	if err != nil {
		return err
	}
	return nil
}

func (b *Adaptor) takeDriveParam(lastDP *DriveParam) *DriveParam {
	b.driveParamMutex.Lock()
	defer b.driveParamMutex.Unlock()
	if l := len(b.driveParam); l > 0 {
		// return new drive param
		dp := b.driveParam[0]
		b.driveParam = b.driveParam[1:len(b.driveParam)]
		return dp
	} else {
		if b.continuousMode {
			// last param retry
			return lastDP
		} else {
			// initialize (hover)
			return &DriveParam {
				pcmd: true,
				flag: 0,
				roll: 0,
				pitch: 0,
				yaw: 0,
				gaz: 0,
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
		pcmd: true,
		flag: 0,
		roll: 0,
		pitch: 0,
		yaw: 0,
		gaz: 0,
	}
	start := time.Now()
	ticker := time.NewTicker(DriveTick * time.Millisecond)
	loop:
	for {
		select {
		case t := <-ticker.C:
			dp = b.takeDriveParam(dp)
			if dp.pcmd {
				millisec := uint32(t.Sub(start).Seconds() * 1000)
				data := make([]byte, 0, 9)
				data = append(data, byte(dp.flag))
				data = append(data, byte(dp.roll))
				data = append(data, byte(dp.pitch))
				data = append(data, byte(dp.yaw))
				data = append(data, byte(dp.gaz))
				binary.LittleEndian.PutUint32(data[5:9], millisec)
				err := b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0a0800919111e4012d1540cb8e", 0x02, 0xfa0a, 0x02, 0x00, 0x00, data[0:9], 15)
				if err != nil {
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
			fmt.Println("set notify REQ fb0f")
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-REQ fb0f-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fb0e0800919111e4012d1540cb8e"]; ok {
			// notify (request with no response on arnetwork)
			fmt.Println("set notify REQ fb0e")
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-REQ fb0e-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fb1b0800919111e4012d1540cb8e"]; ok {
			// notify (response on arnetwork)
			fmt.Println("set notify RES fb1b")
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-RES fb1b-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fb1c0800919111e4012d1540cb8e"]; ok {
			// notify (low latency response on arnetwork)
			fmt.Println("set notify RES fb1c")
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-RES fb1c-")
				fmt.Println(b)
			})
		}
	}
	if s, ok := b.services["9a66fd210800919111e4012d1540cb8e"]; ok {
		if c, ok := s.characteristics["9a66fd220800919111e4012d1540cb8e"]; ok {
			// ????
			fmt.Println("set notify ??? fd22")
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-??? fd22-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fd230800919111e4012d1540cb8e"]; ok {
			// notify (ftp data transfer)
			fmt.Println("set notify DTP DATA fd23")
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-FTP DATA fd23-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fd240800919111e4012d1540cb8e"]; ok {
			// notify (ftp control)
			fmt.Println("set notify DTP DATA fd24")
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-FTP CNTRL fd24-")
				fmt.Println(b)
			})
		}
	}
	if s, ok := b.services["9a66fd510800919111e4012d1540cb8e"]; ok {
		if c, ok := s.characteristics["9a66fd520800919111e4012d1540cb8e"]; ok {
			// ????
			fmt.Println("set notify ??? fd52")
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-??? fd52-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fd530800919111e4012d1540cb8e"]; ok {
			// ????
			fmt.Println("set notify ??? fd53")
			b.peripheral.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, err error){
				fmt.Println("-??? fd53-")
				fmt.Println(b)
			})
		}
		if c, ok := s.characteristics["9a66fd540800919111e4012d1540cb8e"]; ok {
			// ????
			fmt.Println("set notify ??? fd54")
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
