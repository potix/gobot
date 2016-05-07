package client

import (
	"fmt"
	"log"
	"strings"
        "time"
        "sync"
        "math"
        "errors"
        "bytes"
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
	name                    string
	uuid                    string
	device                  gatt.Device
	peripheral              gatt.Peripheral
	state                   gatt.State
	services                map[string]*BLEService
	connected               bool
	peripheralReady         chan error
	seq                     map[uint16]uint8
	seqMutex                *sync.Mutex
	driveLoopEnd            chan bool
	driveParamMutex         *sync.Mutex
	driveParam              []*DriveParam
	continuousMode          bool
	flyingState             uint32
	alertState              uint32
	battery                 uint8
	automaticTakeoff        uint8
	maxAltitudeCurrent      float32
	maxAltitudeMin          float32
	maxAltitudeMax          float32
	maxTiltCurrent          float32
	maxTiltMin              float32
	maxTiltMax              float32
	maxVerticalSpeedCurrent float32
	maxVerticalSpeedMin     float32
	maxVerticalSpeedMax     float32
	maxRotationSpeedCurrent float32
	maxRotationSpeedMin     float32
	maxRotationSpeedMax     float32
	headlightLeft		uint8
	headlightRight		uint8
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
		seqMutex: new(sync.Mutex),
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

func (b *Adaptor) Headlight(left uint8, right uint8) error {
	data := make([]byte, 2, 2)
	data[0] = left
	data[1] = right
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x16, 0x00, data, 8)
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
	var ok bool
	var bles *BLEService
	var blec *BLECharacteristic
	bles, ok = b.services[srvid]
	if !ok {
		return errors.New("not found service")
	}
	blec, ok = bles.characteristics[charid]
	if !ok {
		return errors.New("not found characteristic")
	}
	value := make([]byte, 0, size)
        b.seqMutex.Lock()
	seq := b.seq[seqid]
	b.seq[seqid] += 1
        b.seqMutex.Unlock()
	value = append(value, reqtype, seq, prjid, clsid)
	binary.LittleEndian.PutUint16(value[4:6], cmdid)
	if data != nil {
		value = append(value[:6], data...)
	}
	//--- debug ---
	// fmt.Printf("char = %s, write = %02x\n", blec.uuid, value[:size])
	if err := b.peripheral.WriteCharacteristic(blec.characteristic, value[:size], true); err != nil {
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
			return nil
		}
	}
}

func (b *Adaptor) appendDriveParam(driveParam *DriveParam) {
	b.driveParamMutex.Lock()
	defer b.driveParamMutex.Unlock()
	b.driveParam = append(b.driveParam, driveParam)
}

func (b *Adaptor) driveLoop() {
	var dp *DriveParam = nil
	start := time.Now()
	ticker := time.NewTicker(DriveTick * time.Millisecond)
	loop:
	for {
		select {
		case t := <-ticker.C:
			dp = b.takeDriveParam(dp)
			if dp == nil {
				continue;
			}
			if dp.pcmd {
				fmt.Printf(">>> %v\n", dp)
				millisec := uint32(t.Sub(start).Seconds() * 1000)
				data := make([]byte, 0, 9)
				data = append(data, byte(dp.flag))
				data = append(data, byte(dp.roll))
				data = append(data, byte(dp.pitch))
				data = append(data, byte(dp.yaw))
				data = append(data, byte(dp.gaz))
				binary.LittleEndian.PutUint32(data[5:9], millisec)
				err := b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0a0800919111e4012d1540cb8e", 0x02, 0xfa0a, 0x02, 0x00, 0x02, data[0:9], 15)
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

        blec := b.services[sUUID].characteristics[cUUID]
        val, err := b.peripheral.ReadCharacteristic(blec.characteristic)
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

	// manufacturer
        ms := fmt.Sprintf("%x", a.ManufacturerData)
        fmt.Printf("manufacturer = %s\n", ms)

	// name
	fmt.Printf("name = %s\n", p.Name())

	// check device
	if !strings.HasPrefix(p.Name(), "Swat_") && !strings.HasPrefix(p.Name(), "Maclan_") && ms != "4300cf1907090100" {
		// not match device
		return
	}

	fmt.Printf("device match\n")

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

func (b *Adaptor) notificationBase(c *gatt.Characteristic, data []byte, err error, nores bool, ressrvid string, rescharid string, resseqid uint16){
	if err != nil {
		fmt.Printf("notification errror (%v)\n", err)
		return
	}
	if len(data) < 6 {
		fmt.Printf("invalid notification\n")
		fmt.Printf("%02x\n", data)
		return
	}
	var reqcmdid uint16
	reqtype := data[0]
	reqseq := data[1]
	reqprjid := data[2]
	reqclsid := data[3]
	binary.Read(bytes.NewReader(data[4:6]), binary.LittleEndian, &reqcmdid)
	switch reqtype {
	case 1:
		// response
		fmt.Printf("unexpected request type (ack)\n")
		fmt.Printf("%02x\n", data)
	case 2:
		switch reqprjid {
		case 0: // common
			switch reqclsid {
			case 5:
				switch reqcmdid {
				case 1:
					binary.Read(bytes.NewReader(data[6:7]), binary.LittleEndian, &b.battery)
					fmt.Printf("battry %d\n", b.battery)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			case 23:
				switch reqcmdid {
				case 0:
					b.headlightLeft = data[6]
					b.headlightRight = data[7]
					fmt.Printf("headlightIntensityChanged left = %d, right = %d\n", b.headlightLeft, b.headlightRight)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			default:
				fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x)\n", reqprjid, reqclsid)
				fmt.Printf("%02x\n", data)
			}
		case 2: // minidrone
			switch reqclsid {
			default:
				fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x)\n", reqprjid, reqclsid)
				fmt.Printf("%02x\n", data)
			}
		case 128: // common debug
			// common debug project id
			fmt.Printf("unexpected project id (common debug)\n")
			fmt.Printf("%02x\n", data)
		case 130: // minidrone debug
			// unknown project id
			fmt.Printf("unexpected project id (minidrone debug)\n")
			fmt.Printf("%02x\n", data)
		default:
			// unknown project id
			fmt.Printf("unexpected project id (unkown)\n")
			fmt.Printf("%02x\n", data)
		}
	case 3:
		// low latency request is exists ???
		fmt.Printf("unexpected request type (low latency)\n")
		fmt.Printf("%02x\n", data)
	case 4:
		switch reqprjid {
		case 0: // common
			switch reqclsid {
			default:
				fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x)\n", reqprjid, reqclsid)
				fmt.Printf("%02x\n", data)
			}
		case 2: // minidrone
			switch reqclsid {
			case 3:
				switch reqcmdid {
				case 0:
					fmt.Printf("FlatTrimChanged\n")
				case 1:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.flyingState)
					fmt.Printf("FlyingStateChanged %d\n", b.flyingState)
				case 2:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.alertState)
					fmt.Printf("AlertStateChanged %d\n", b.alertState)
				case 3:
					binary.Read(bytes.NewReader(data[6:7]), binary.LittleEndian, &b.automaticTakeoff)
					fmt.Printf("AutomaticTakeoffMode %d\n", b.automaticTakeoff)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			case 5:
				switch reqcmdid {
				case 0:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.maxVerticalSpeedCurrent)
					binary.Read(bytes.NewReader(data[10:14]), binary.LittleEndian, &b.maxVerticalSpeedMin)
					binary.Read(bytes.NewReader(data[14:18]), binary.LittleEndian, &b.maxVerticalSpeedMax)
					fmt.Printf("MaxVerticalSpeedChanged current = %f, min = %f, max = %f\n", b.maxVerticalSpeedCurrent, b.maxVerticalSpeedMin, b.maxVerticalSpeedMax)
				case 1:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.maxRotationSpeedCurrent)
					binary.Read(bytes.NewReader(data[10:14]), binary.LittleEndian, &b.maxRotationSpeedMin)
					binary.Read(bytes.NewReader(data[14:18]), binary.LittleEndian, &b.maxRotationSpeedMax)
					fmt.Printf("MaxRotationSpeedChanged current = %f, min = %f, max = %f\n", b.maxRotationSpeedCurrent, b.maxRotationSpeedMin, b.maxRotationSpeedMax)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			case 9:
				switch reqcmdid {
				case 0:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.maxAltitudeCurrent)
					binary.Read(bytes.NewReader(data[10:14]), binary.LittleEndian, &b.maxAltitudeMin)
					binary.Read(bytes.NewReader(data[14:18]), binary.LittleEndian, &b.maxAltitudeMax)
					fmt.Printf("MaxAltitudeChanged current = %f, min = %f, max = %f\n", b.maxAltitudeCurrent, b.maxAltitudeMin, b.maxAltitudeMax)
				case 1:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.maxTiltCurrent)
					binary.Read(bytes.NewReader(data[10:14]), binary.LittleEndian, &b.maxTiltMin)
					binary.Read(bytes.NewReader(data[14:18]), binary.LittleEndian, &b.maxTiltMax)
					fmt.Printf("MaxTilitChanged current = %f, min = %f, max = %f\n", b.maxTiltCurrent, b.maxTiltMin, b.maxTiltMax)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			default:
				fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x)\n", reqprjid, reqclsid)
				fmt.Printf("%02x\n", data)
			}
		case 128: // common debug
			// common debug project id
			fmt.Printf("unexpected project id (common debug)\n")
			fmt.Printf("%02x\n", data)
		case 130: // minidrone debug
			// unknown project id
			fmt.Printf("unexpected project id (minidrone debug)\n")
			fmt.Printf("%02x\n", data)
		default:
			// unknown project id
			fmt.Printf("unexpected project id (unkown)\n")
			fmt.Printf("%02x\n", data)
		}
	default:
		// unknown request type
		fmt.Printf("unexpected request type (unkown)\n")
		fmt.Printf("%02x\n", data)
	}
	if nores {
		if reqtype == 0x04 {
			fmt.Printf("unexpected request type (nores is true but need response)\n")
			fmt.Printf("%02x\n", data)
		}
		return
	}
	// response
	var ok bool
	var bles *BLEService
	var blec *BLECharacteristic
	bles, ok = b.services[ressrvid]
	if !ok {
		fmt.Println("not found service")
	}
	blec, ok = bles.characteristics[rescharid]
	if !ok {
		fmt.Println("not found characteristic")
	}
	value := make([]byte, 0, 3)
	b.seqMutex.Lock()
	resseq := b.seq[resseqid]
	b.seq[resseqid] += 1
	b.seqMutex.Unlock()
	value = append(value, 0x01, resseq, reqseq)
	//--- debug ---
	// fmt.Printf("char = %s, write = %02x\n", blec.uuid, value[:3])
	if err := b.peripheral.WriteCharacteristic(blec.characteristic, value[:3], true); err != nil {
		fmt.Printf("notification response failure\n", err)
		return
	}
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
		fmt.Printf("\t%s\n", s.UUID().String())
		fmt.Println("\tdiscoveryService")
		cs, err := b.peripheral.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}
		for _, c := range cs {
			b.services[s.UUID().String()].characteristics[c.UUID().String()] = NewBLECharacteristic(c.UUID().String(), c)
			fmt.Printf("\t\t%s\n", c.UUID().String())
			fmt.Println("\t\tdiscoveryDescripto")
			ds, err := b.peripheral.DiscoverDescriptors(nil, c)
			if err != nil {
				fmt.Printf("Failed to discover discriptors, err: %s\n", err)
				continue
			}
			for _, d := range ds {
				b.services[s.UUID().String()].characteristics[c.UUID().String()].descriptors[d.UUID().String()] = NewBLEDescriptor(d.UUID().String(), d)
				fmt.Printf("\t\t\t%s\n", d.UUID().String())
			}
		}
	}
	// add service
	if bles, ok := b.services["9a66fb000800919111e4012d1540cb8e"]; ok {
		if blec, ok := bles.characteristics["9a66fb0f0800919111e4012d1540cb8e"]; ok {
			// notify (request with no response on arnetwork)
			if err := b.peripheral.SetNotifyValue(blec.characteristic, func(c *gatt.Characteristic, data []byte, err error){
				fmt.Println("-Notification REQ fb0f-")
				b.notificationBase(c, data, err, true, "", "", 0)
			}); err != nil {
				fmt.Println(err)
			}
		}
		if blec, ok := bles.characteristics["9a66fb0e0800919111e4012d1540cb8e"]; ok {
			// notify (request with need response on arnetwork)
			if err := b.peripheral.SetNotifyValue(blec.characteristic, func(c *gatt.Characteristic, data []byte, err error){
				fmt.Println("-Notification REQ fb0e-")
				b.notificationBase(c, data, err, false, "9a66fa000800919111e4012d1540cb8e", "9a66fa1e0800919111e4012d1540cb8e", 0xfa1e)
			}); err != nil {
				fmt.Println(err)
			}
		}
		if blec, ok := bles.characteristics["9a66fb1b0800919111e4012d1540cb8e"]; ok {
			// notify (response on arnetwork)
			if err := b.peripheral.SetNotifyValue(blec.characteristic, func(c *gatt.Characteristic, data []byte, err error){
				fmt.Println("Notification -RES fb1b-")
				fmt.Printf("%02x\n", data)
				// TODO check seq
			}); err != nil {
				fmt.Println(err)
			}
		}
		if blec, ok := bles.characteristics["9a66fb1c0800919111e4012d1540cb8e"]; ok {
			// notify (low latency response on arnetwork)
			if err := b.peripheral.SetNotifyValue(blec.characteristic, func(c *gatt.Characteristic, data []byte, err error){
				fmt.Println("-Notification RES fb1c-")
				fmt.Printf("%02x\n", data)
				// TODO check seq
			}); err != nil {
				fmt.Println(err)
			}
		}
	}
	if bles, ok := b.services["9a66fd210800919111e4012d1540cb8e"]; ok {
		if blec, ok := bles.characteristics["9a66fd220800919111e4012d1540cb8e"]; ok {
			// ????
			if err := b.peripheral.SetNotifyValue(blec.characteristic, func(c *gatt.Characteristic, data []byte, err error){
				fmt.Println("-??? fd22-")
				fmt.Printf("%02x\n", data)
			}); err != nil {
				fmt.Println(err)
			}
		}
		if blec, ok := bles.characteristics["9a66fd230800919111e4012d1540cb8e"]; ok {
			// notify (ftp data transfer)
			if err := b.peripheral.SetNotifyValue(blec.characteristic, func(c *gatt.Characteristic, data []byte, err error){
				fmt.Println("-FTP DATA fd23-")
				fmt.Printf("%02x\n", data)
			}); err != nil {
				fmt.Println(err)
			}
		}
		if blec, ok := bles.characteristics["9a66fd240800919111e4012d1540cb8e"]; ok {
			// notify (ftp control)
			if err := b.peripheral.SetNotifyValue(blec.characteristic, func(c *gatt.Characteristic, data []byte, err error){
				fmt.Println("-FTP CNTRL fd24-")
				fmt.Printf("%02x\n", data)
			}); err != nil {
				fmt.Println(err)
			}
		}
	}
	if bles, ok := b.services["9a66fd510800919111e4012d1540cb8e"]; ok {
		if blec, ok := bles.characteristics["9a66fd520800919111e4012d1540cb8e"]; ok {
			// ????
			if err := b.peripheral.SetNotifyValue(blec.characteristic, func(c *gatt.Characteristic, data []byte, err error){
				fmt.Println("-??? fd52-")
				fmt.Printf("%02x\n", data)
			}); err != nil {
				fmt.Println(err)
			}
		}
		if blec, ok := bles.characteristics["9a66fd530800919111e4012d1540cb8e"]; ok {
			// ????
			if err := b.peripheral.SetNotifyValue(blec.characteristic, func(c *gatt.Characteristic, data []byte, err error){
				fmt.Println("-??? fd53-")
				fmt.Printf("%02x\n", data)
			}); err != nil {
				fmt.Println(err)
			}
		}
		if blec, ok := bles.characteristics["9a66fd540800919111e4012d1540cb8e"]; ok {
			// ????
			if err := b.peripheral.SetNotifyValue(blec.characteristic, func(c *gatt.Characteristic, data []byte, err error){
				fmt.Println("-??? fd54-")
				fmt.Printf("%02x\n", data)
			}); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// Represents a BLE Peripheral's Service
type BLEService struct {
	uuid            string
	service         *gatt.Service
	characteristics map[string]*BLECharacteristic
}

// Represents a BLE Peripheral's Charactoristic
type BLECharacteristic struct {
	uuid            string
	characteristic  *gatt.Characteristic
	descriptors     map[string]*BLEDescriptor
}

// Represents a BLE Peripheral's Charactoristic
type BLEDescriptor struct {
	uuid            string
	descriptor      *gatt.Descriptor
}

// NewBLEService returns a new BLEService given a uuid
func NewBLEService(suuid string, service *gatt.Service) *BLEService {
	return &BLEService{
		uuid:            suuid,
		service:         service,
		characteristics: make(map[string]*BLECharacteristic),
	}
}

// NewBLECharacteristic returns a new NewBLECharacteristic given a uuid
func NewBLECharacteristic(cuuid string, characteristic *gatt.Characteristic) *BLECharacteristic {
	return &BLECharacteristic{
		uuid:            cuuid,
		characteristic:  characteristic,
		descriptors:     make(map[string]*BLEDescriptor),
	}
}

// NewAdaptor returns a new BLEService given a uuid
func NewBLEDescriptor(duuid string, descriptor *gatt.Descriptor) *BLEDescriptor {
	return &BLEDescriptor{
		uuid:        duuid,
		descriptor:  descriptor,
	}
}
