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
        "crypto/md5"
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

type driveParam struct {
	pcmd  bool
	flag  uint8
	roll  int8
	pitch int8
	yaw   int8
	gaz   int8
}

type ftpCommand struct {
	cmd string
	path string
}

type ftpResult struct {
	result []byte
	err error
}

// Represents a Connection to a BLE Peripheral
type Adaptor struct {
	name                        string
	uuid                        string
	device                      gatt.Device
	peripheral                  gatt.Peripheral
	state                       gatt.State
	services                    map[string]*BLEService
	connected                   bool
	peripheralReady             chan error
	seq                         map[uint16]uint8
	seqMutex                    *sync.Mutex
	driveLoopEnd                chan bool
	driveParamMutex             *sync.Mutex
	driveParam                  []*driveParam
	continuousMode              bool
	flyingState                 uint32
	alertState                  uint32
	battery                     uint8
	automaticTakeoff            uint8
	maxAltitudeCurrent          float32
	maxAltitudeMin              float32
	maxAltitudeMax              float32
	maxTiltCurrent              float32
	maxTiltMin                  float32
	maxTiltMax                  float32
	maxVerticalSpeedCurrent     float32
	maxVerticalSpeedMin         float32
	maxVerticalSpeedMax         float32
	maxRotationSpeedCurrent     float32
	maxRotationSpeedMin         float32
	maxRotationSpeedMax         float32
	maxHorizontalSpeedCurrent     float32
	maxHorizontalSpeedMin         float32
	maxHorizontalSpeedMax         float32
	supportedAccessory          uint32
	productName                 string
	productSoftwareVersion      string
	productHardwareVersion      string
	productSerialHigh           string
	productSerialLow            string
	productModel                uint32
	libARCommandsVersion        string
	currentCountry              string
	autoCountry                 uint8
	currentDate                 string
	currentTime                 string
	headlightLeft               uint8
	headlightRight              uint8
	cutOutMode                  bool
	chargingPhase               uint32
	chargingRate                uint32
	chargeIntensity             uint8
	chargeFullChargingTime      uint8
	pictureStateV2	            uint32
	pictureStateV2Error         uint32
	pictureStateV1              uint8
	pictureStateV1MassStorageID uint8
	pictureEvent		    uint32
	pictureEventError	    uint32
	disconnectionCause          uint32
	emergencyLoopChan           chan bool
	emergencyLoopEnd            chan bool
	ftpBuffer                   []byte
	ftpCmdType                  string
	ftpState                    uint8
	ftpResult                   []byte
	ftpLocalDigest              string
	ftpReqChan                  chan *ftpCommand
	ftpResChan                  chan *ftpResult
	ftpLoopEnd                  chan bool
	sensorIMU                   uint8
	sensorBarometer             uint8
	sensorUltrasound            uint8
	sensorGPS                   uint8
	sensorMagnetometer          uint8
	sensorVerticalCamera        uint8
	massStorageID               uint8
	massStorageName             string
        massStorageSize             uint32
        massStorageUsedSize         uint32
        massStoragePlugged          uint8
        massStorageFull             uint8
        massStorageInternal         uint8
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
		driveParam: make([]*driveParam, 0, 0),
		emergencyLoopChan: make(chan bool),
		emergencyLoopEnd: make(chan bool),
		ftpBuffer: make([]byte, 0, 0),
		ftpReqChan: make(chan *ftpCommand),
		ftpResChan: make(chan *ftpResult),
		ftpLoopEnd: make(chan bool),
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


	// set datetime
	b.setDateTime()

	// all settings
	b.allSettings()

	// all states
	b.allStates()

	// start drive
	b.startDrive()

	return nil
}

func (b *Adaptor) setDateTime() {
	now := time.Now()
	data := make([]byte, 0, 0)

	// set current date
	data = append(data, []byte(fmt.Sprintf("%02d-%02d-%02d", now.Year(), now.Month(), now.Day()))...)
	data = append(data, 0x00)
        if err := b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x04, 0x01, data); err != nil {
		fmt.Println("could not set current date")
	}

	// set current time
	var pn  string
	data = data[:0]
	_, offset := now.Zone()
	if offset > 0 {
		pn = "+"
	} else if offset < 0 {
		pn = "-"
	}
	if offset == 0 {
		data = append(data, []byte(fmt.Sprintf("T%02d%02d%02dZ", now.Hour(), now.Minute(), now.Second()))...)
	} else {
		data = append(data, []byte(fmt.Sprintf("T%02d%02d%02d%s%02d%02d", now.Hour(), now.Minute(), now.Second(), pn, offset / 3600, (offset % 3600) / 60))...)
	}
	data = append(data, 0x00)
        if err := b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x04, 0x02, data); err != nil {
		fmt.Println("could not set current time")
	}
}

func (b *Adaptor) allSettings() {
	// retry .... ummm
        if err := b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x02, 0x00, nil); err != nil {
		fmt.Println("could not get all settings")
	}
}

func (b *Adaptor) allStates() {
	// retry .... ummm
        if err :=  b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x04, 0x00, nil); err != nil {
		fmt.Println("could not get all states")
	}
}

func (b *Adaptor) TakeOff() error {
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x00, 0x01, nil)
}

func (b *Adaptor) Landing() error {
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x00, 0x03, nil)
}

func (b *Adaptor) Flip(value uint32) error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], value)
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x04, 0x00, data)
}

func (b *Adaptor) SetMaxAltitude(altitude float32) error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(altitude))
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x08, 0x00, data)
}

func (b *Adaptor) SetMaxTilt(tilt float32) error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(tilt))
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x08, 0x01, data)
}

func (b *Adaptor) SetMaxVirticalSpeed(virticalSpeed float32) error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(virticalSpeed))
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x01, 0x00, data)
}

func (b *Adaptor) SetMaxRotationSpeed(rotationSpeed float32) error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(rotationSpeed))
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x01, 0x01, data)
}

func (b *Adaptor) SetCutOutMode(onOff bool) error {
	data := make([]byte, 1, 1)
	if onOff {
		data[0] = 1
	} else {
		data[0] = 0
	}
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0xa, 0x00, data)
}

func (b *Adaptor) FlatTrim() error {
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x0, 0x00, nil)
}

func (b *Adaptor) Emergency() error {
	// TODO retry... ummmm
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0c0800919111e4012d1540cb8e", 0x04, 0xfa0c, 0x02, 0x0, 0x04, nil)
}

func (b *Adaptor) Headlight(left uint8, right uint8) error {
	data := make([]byte, 2, 2)
	data[0] = left
	data[1] = right
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x16, 0x00, data)
}

func (b *Adaptor) HeadlightFlashStart() error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], 0)
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x18, 0x00, data)
}

func (b *Adaptor) HeadlightBlinkStart() error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], 1)
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x18, 0x00, data)
}

func (b *Adaptor) HeadlightOscillationStart() error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], 2)
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x18, 0x00, data)
}

func (b *Adaptor) HeadlightFlashStop() error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], 0)
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x18, 0x01, data)
}

func (b *Adaptor) HeadlightBlinkStop() error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], 1)
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x18, 0x01, data)
}

func (b *Adaptor) HeadlightOscillationStop() error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data[0:4], 2)
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x00, 0x18, 0x01, data)
}

func (b *Adaptor) TakePicture() error {
	if (b.pictureStateV2 != 0) {
		return errors.New("Now it is't possible to take a picture ")
	}
        return b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0b0800919111e4012d1540cb8e", 0x04, 0xfa0b, 0x02, 0x06, 0x01, nil)
}

// TODO
//   all settings(retry) 00-02-00        ProductNameChanged 00-03-02  ProductSerialHighChanged 00-03-04  ProductSerialLowChanged 00-03-05 SupportedAccessoriesListChanged 00-1b(27)-00
//                                ProductVersionChanged 00-03-03 AutoCountryChanged 00-03-07 CountryChanged 00-03-06 AccessoryConfigChanged  00-1b(27)-01 MaxHorizontalSpeedChanged 02-05-03
//                                AllSettingsChanged 00-03-00
//   AllStates(retry)    00-04-00  DeviceLibARCommandsVersion 00-12(18)-02 ProductModel 00-05-09 HeadlightsState 00-23-00 AnimationsStateList 00-19(25)-00 ChargingInfo 00-1d(29)-03
//                                 MassStorageStateListChanged 00-05-02

func (b *Adaptor) FTPList(path string) ([]byte, error) {
	ftpCmd := &ftpCommand {
		cmd: "LIS",
		path: path,
	}
	b.ftpReqChan <- ftpCmd
	fr := <-b.ftpResChan
	return fr.result, fr.err
}

func (b *Adaptor) FTPGet(path string) ([]byte, error) {
	ftpCmd := &ftpCommand {
		cmd: "GET",
		path: path,
	}
	b.ftpReqChan <- ftpCmd
	fr := <-b.ftpResChan
	return fr.result, fr.err
}

func (b *Adaptor) FTPDelete(path string) ([]byte, error) {
	ftpCmd := &ftpCommand {
		cmd: "DEL",
		path: path,
	}
	b.ftpReqChan <- ftpCmd
	fr := <-b.ftpResChan
	return fr.result, fr.err
}

func (b *Adaptor) GetBattery() uint8 {
        return b.battery
}

func (b *Adaptor) GetFlyingState() uint32 {
        return b.flyingState
}

func (b *Adaptor) GetPictureState() uint32 {
        return b.pictureStateV2
}

func (b *Adaptor) AddDrive(tickCnt int, flag uint8, roll int8, pitch int8, yaw int8, gaz int8) {
	for i := 0; i < tickCnt; i++ {
		dp := &driveParam {
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

func (b *Adaptor) writeCharBase(srvid string, charid string, reqtype uint8, seqid uint16, prjid uint8, clsid uint8, cmdid uint16, data []byte) error {
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
	var hdrSize int = 6
	value := make([]byte, hdrSize, hdrSize + len(data))
        b.seqMutex.Lock()
	seq := b.seq[seqid]
	b.seq[seqid] += 1
        b.seqMutex.Unlock()
	value[0] = byte(reqtype)
	value[1] = byte(seq)
	value[2] = byte(prjid)
	value[3] = byte(clsid)
	binary.LittleEndian.PutUint16(value[4:6], cmdid)
	if data != nil {
		value = append(value[:6], data...)
	}
	//--- debug ---
	// fmt.Printf("char = %s, write = %02x\n", blec.uuid, value[:size])
	if err := b.peripheral.WriteCharacteristic(blec.characteristic, value, true); err != nil {
		return err
	}
	return nil
}

func (b *Adaptor) ftpWriteCharBase(srvid string, charid string, data []byte, size int) error {
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
	//--- debug ---
	fmt.Printf("char = %s, write = %02x\n", blec.uuid, data[:size])
	if err := b.peripheral.WriteCharacteristic(blec.characteristic, data[:size], true); err != nil {
		return err
	}
	return nil
}

func (b *Adaptor) takeDriveParam(lastDP *driveParam) *driveParam {
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

func (b *Adaptor) appendDriveParam(driveParam *driveParam) {
	b.driveParamMutex.Lock()
	defer b.driveParamMutex.Unlock()
	b.driveParam = append(b.driveParam, driveParam)
}

func (b *Adaptor) driveLoop() {
	var skipcnt int = 0
	dp := &driveParam{
		pcmd: true,
	}
	start := time.Now()
	ticker := time.NewTicker(DriveTick * time.Millisecond)
	loop:
	for {
		select {
		case t := <-ticker.C:
			dp = b.takeDriveParam(dp)
			if dp == nil {
				skipcnt += 1
				if (skipcnt == 6) {
					dp = &driveParam{
						pcmd: true,
					}
					skipcnt = 0
				} else {
					continue;
				}
			} else {
				skipcnt = 0
			}
			if dp.pcmd {
				// --- debug ---
				//fmt.Printf(">>> %v\n", dp)
				millisec := uint32(t.Sub(start).Seconds() * 1000)
				data := make([]byte, 9, 9)
				data[0] = byte(dp.flag)
				data[1] = byte(dp.roll)
				data[2] = byte(dp.pitch)
				data[3] = byte(dp.yaw)
				data[4] = byte(dp.gaz)
				binary.LittleEndian.PutUint32(data[5:9], millisec)
				err := b.writeCharBase("9a66fa000800919111e4012d1540cb8e", "9a66fa0a0800919111e4012d1540cb8e", 0x02, 0xfa0a, 0x02, 0x00, 0x02, data)
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

func (b *Adaptor) emergencyLoop() {
	loop:
	for {
		select {
		case <-b.emergencyLoopChan:
			b.Landing()
		case <-b.emergencyLoopEnd:
			break loop
		}
	}
}

func (b *Adaptor) ftpLoop() {
	loop:
	for {
		select {
		case ftpCmd := <-b.ftpReqChan:
			b.ftpCmdType = ftpCmd.cmd
			all := []byte(ftpCmd.cmd)
			if ftpCmd.cmd == "MD5 OK" {
				all = append(all, 0x00)
				if err := b.ftpWriteCharBase("9a66fd210800919111e4012d1540cb8e", "9a66fd230800919111e4012d1540cb8e", all[0:7], 7); err != nil {
					fr := &ftpResult {
						err: errors.New("ftp command error"),
					}
					b.ftpResChan <- fr
				}
			} else {
				all = append(all, ftpCmd.path...)
				all = append(all, 0x00)
				partial := make([]byte, 20, 20)
				var sndcnt = 0
				var seq uint8
				var wsize int
				for  {
					l := len(all)
					if l == 0 {
						break
					}
					if l > 19 {
						wsize = 19
					} else {
						wsize = l
					}
					copy(partial[1:1 + wsize], all[0:wsize])
					all = all[wsize:]
					rl := len(all)
					if rl == 0 && sndcnt == 0 {
						seq = 3
					} else if rl > 0 && wsize == 19 && sndcnt == 0 {
						seq = 2
					} else if rl > 0 && wsize == 19 && sndcnt > 0 {
						seq = 0
					} else if wsize < 19 {
						seq = 1
					}
					partial[0] = seq
					if err := b.ftpWriteCharBase("9a66fd210800919111e4012d1540cb8e", "9a66fd240800919111e4012d1540cb8e", partial[0:1 + wsize], 1 + wsize); err != nil {
						fr := &ftpResult {
							err: errors.New("ftp command error"),
						}
						b.ftpResChan <- fr
						break
					}
					sndcnt += 1
				}
			}
		case <-b.ftpLoopEnd:
			break loop
		}
	}
}

// start drive
func (b *Adaptor) startDrive() {
	go b.driveLoop()
	go b.emergencyLoop()
	go b.ftpLoop()
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
	b.emergencyLoopEnd <- true
	b.ftpLoopEnd <- true
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
				case 9:
					binary.Read(bytes.NewReader(data[6:7]), binary.LittleEndian, &b.productModel)
					fmt.Printf("ProductModel model = %d\n", b.productModel)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			case 18:
				switch reqcmdid {
				case 2:
					b.libARCommandsVersion = string(data[6:len(data) - 1])
					fmt.Printf("DeviceLibARCommandsVersion version = %s\n", b.libARCommandsVersion)
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
			case 25:
				switch reqcmdid {
				case 0:
					var animationList uint32
					var animationRate uint32
					var animationError uint32
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &animationList)
					binary.Read(bytes.NewReader(data[10:14]), binary.LittleEndian, &animationRate)
					binary.Read(bytes.NewReader(data[14:18]), binary.LittleEndian, &animationError)
					fmt.Printf("AnimationsState list = %d, rate = %d, error = %d\n", animationList, animationRate, animationError)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			case 27:
				switch reqcmdid {
				case 0:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.supportedAccessory)
					fmt.Printf("SupportedAccessoriesListChanged accessory = %d\n", b.supportedAccessory)
				case 1:
					var newAccessory uint32
					var accessoryConfigError uint32
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &newAccessory)
					binary.Read(bytes.NewReader(data[10:14]), binary.LittleEndian, &accessoryConfigError)
					fmt.Printf("AccessoryConfigChanged newAccessory = %d error = %d\n", newAccessory, accessoryConfigError)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			case 29:
				switch reqcmdid {
				case 3:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.chargingPhase)
					binary.Read(bytes.NewReader(data[10:14]), binary.LittleEndian, &b.chargingRate)
					b.chargeIntensity = data[14]
					b.chargeFullChargingTime = data[15]
					fmt.Printf("ChargingInfo phase= %d, rate = %d, intensity = %d, fullChargeTime = %d\n",
					    b.chargingPhase, b.chargingRate, b.chargeIntensity, b.chargeFullChargingTime)
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
			case 1:
				switch reqcmdid {
				case 0:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.disconnectionCause)
					fmt.Printf("Disconnection case = %d\n", b.disconnectionCause)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			case 3:
				switch reqcmdid {
				case 0:
					// AllSettingsChanged
					fmt.Printf("AllSettingsChanged\n")
				case 2:
					b.productName = string(data[6:len(data) - 1])
					fmt.Printf("ProductNameChanged productName = %s\n", b.productName)
				case 3:
					var idx int = 0;
					for d := range data[6:] {
						idx += 1
						if d == 0 {
							break
						}
					}
					b.productSoftwareVersion = string(data[6:6+idx-1])
					b.productHardwareVersion = string(data[6+idx-1:len(data) - 1])
					fmt.Printf("ProductVersionChanged productSoftwareVersion = %s, productHardwareVersion = %s\n", b.productSoftwareVersion, b.productHardwareVersion )
				case 4:
					b.productSerialHigh = string(data[6:len(data) - 1])
					fmt.Printf("ProductSerialHighChanged productSerialHigh = %s\n", b.productSerialHigh)
				case 5:
					b.productSerialLow = string(data[6:len(data) - 1])
					fmt.Printf("ProductSerialLowChanged productSerialLow = %s\n", b.productSerialLow)
				case 6:
					b.currentCountry = string(data[6:len(data) - 1])
					fmt.Printf("CountryChanged currentCountry = %s\n", b.currentCountry)
				case 7:
					b.autoCountry = data[6]
					fmt.Printf("AutoCountryChanged autoCountry = %d\n", b.autoCountry)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			case 5:
				switch reqcmdid {
				case 0:
					// AllStatesChanged
					fmt.Printf("AllStatesChanged\n")
				case 2:
					b.massStorageID = data[6]
					b.massStorageName = string(data[7:len(data) - 1])
					fmt.Printf("MassStorageStateListChanged id = %d, name = %s\n", b.massStorageID, b.massStorageName)
				case 3:
					b.massStorageID = data[6]
					binary.Read(bytes.NewReader(data[7:11]), binary.LittleEndian, &b.massStorageSize)
					binary.Read(bytes.NewReader(data[11:15]), binary.LittleEndian, &b.massStorageUsedSize)
					b.massStoragePlugged = data[15]
					b.massStorageFull = data[16]
					b.massStorageInternal = data[17]
					fmt.Printf("MassStorageInfoStateListChanged id = %d, size = %d, usedSize = %d, plugged = %d, full = %d, internal = %d \n",
					    b.massStorageID, b.massStorageSize, b.massStorageUsedSize, b.massStoragePlugged, b.massStorageFull, b.massStorageInternal)
				case 4:
					b.currentDate = string(data[6:len(data) - 1])
					fmt.Printf("CurrentDateChanged date = %s\n", b.currentDate)
				case 5:
					b.currentTime = string(data[6:len(data) - 1])
					fmt.Printf("CurrentTimeChanged time = %s\n", b.currentTime)
				case 8:
					var sendorType uint32
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &sendorType)
					switch sendorType {
					case 0:
						b.sensorIMU = data[10]
						fmt.Printf("SensorsStatesListChanged sensorIMU = %d\n", b.sensorIMU)
					case 1:
						b.sensorBarometer = data[10]
						fmt.Printf("SensorsStatesListChanged sensorBarometer = %d\n", b.sensorBarometer)
					case 2:
						b.sensorUltrasound = data[10]
						fmt.Printf("SensorsStatesListChanged sensorUltrasound = %d\n", b.sensorUltrasound)
					case 3:
						b.sensorGPS = data[10]
						fmt.Printf("SensorsStatesListChanged sensorGPS = %d\n", b.sensorGPS)
					case 4:
						b.sensorMagnetometer = data[10]
						fmt.Printf("SensorsStatesListChanged sensorMagnetometer = %d\n", b.sensorMagnetometer)
					case 5:
						b.sensorVerticalCamera = data[10]
						fmt.Printf("SensorsStatesListChanged sensorVerticalCamera = %d\n", b.sensorVerticalCamera)
					}
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
			case 2:
				switch reqcmdid {
				case 0:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.pictureEvent)
					binary.Read(bytes.NewReader(data[10:14]), binary.LittleEndian, &b.pictureEventError)
					fmt.Printf("PictureEventChanged state = %d, error = %d\n", b.pictureEvent, b.pictureEventError)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			case 3:
				switch reqcmdid {
				case 0:
					// FlatTrimChanged
					fmt.Printf("FlatTrimChanged\n")
				case 1:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.flyingState)
					fmt.Printf("FlyingStateChanged %d\n", b.flyingState)
					if (b.flyingState == 5 /* emergency */) {
						b.emergencyLoopChan <- true
					}
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
				case 3:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.maxHorizontalSpeedCurrent)
					binary.Read(bytes.NewReader(data[10:14]), binary.LittleEndian, &b.maxHorizontalSpeedMin)
					binary.Read(bytes.NewReader(data[14:18]), binary.LittleEndian, &b.maxHorizontalSpeedMax)
					fmt.Printf("MaxHorizontalSpeedChanged current = %f, min = %f, max = %f\n", b.maxHorizontalSpeedCurrent, b.maxHorizontalSpeedMin, b.maxHorizontalSpeedMax)
				default:
					fmt.Printf("unexpected class id (unknown reqclsid %02x-%02x-%02x)\n", reqprjid, reqclsid, reqcmdid)
					fmt.Printf("%02x\n", data)
				}
			case 7:
				switch reqcmdid {
				case 0: // ... deprecated ...
					b.pictureStateV1 = data[6]
					b.pictureStateV1MassStorageID = data[7]
					fmt.Printf("PictureStateChangedV1 state = %d, massStorageID = %d\n", b.pictureStateV1, b.pictureStateV1MassStorageID)
				case 1:
					binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &b.pictureStateV2)
					binary.Read(bytes.NewReader(data[10:14]), binary.LittleEndian, &b.pictureStateV2Error)
					fmt.Printf("PictureStateChangedV2 state = %d, error = %d\n", b.pictureStateV2, b.pictureStateV2Error)
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
			case 11:
				switch reqcmdid {
				case 2:
					if data[6] == 1 {
						b.cutOutMode = true
					} else {
						b.cutOutMode = false
					}
					fmt.Printf("CutOutModeChanged = %d\n", b.cutOutMode)
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
		fmt.Println("\tdiscoveryCharacteristic")
		cs, err := b.peripheral.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}
		for _, c := range cs {
			b.services[s.UUID().String()].characteristics[c.UUID().String()] = NewBLECharacteristic(c.UUID().String(), c)
			fmt.Printf("\t\t%s\n", c.UUID().String())
			fmt.Println("\t\tdiscoveryDescriptor")
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
				//fmt.Println("-FTP DATA fd23-")
				fmt.Printf("%02x\n", data)
				if b.ftpCmdType == "LIS" || b.ftpCmdType == "DEL"  {
					if b.ftpState == 0 && data[0] == 3 && len(data) > 1 {
						msg := string(data[1:])
						if strings.HasPrefix(msg, "error") {
							b.ftpBuffer = b.ftpBuffer[:0]
							b.ftpCmdType = ""
							fr := &ftpResult {
								err: errors.New(string(data[1:len(data) - 1])),
							}
							b.ftpResChan <- fr
						} else if strings.HasPrefix(msg, "Delete successful") {
							b.ftpBuffer = b.ftpBuffer[:0]
							b.ftpCmdType = ""
							fr := &ftpResult {
								result: data[1:len(data) - 1],
							}
							b.ftpResChan <- fr
						} else if strings.HasPrefix(msg, "End of Transfer") {
							b.ftpResult = b.ftpBuffer
							b.ftpBuffer = make([]byte, 0, len(b.ftpResult))
							b.ftpState = 1
							b.ftpLocalDigest = fmt.Sprintf("%02x", md5.Sum(b.ftpResult))
						} else {
							b.ftpBuffer = b.ftpBuffer[:0]
							b.ftpCmdType = ""
							fr := &ftpResult {
								err: errors.New(fmt.Sprintf("unexpected message (%s)\n", msg)),
							}
							b.ftpResChan <- fr
						}
					} else if b.ftpState == 0 {
						b.ftpBuffer = append(b.ftpBuffer, data[1:]...)
					} else if b.ftpState == 1 {
						b.ftpBuffer = append(b.ftpBuffer, data[1:]...)
						if data[0] == 1 {
							ftpRemoteDigest := string(b.ftpBuffer)
							b.ftpBuffer = b.ftpBuffer[:0]
							b.ftpState = 0
							b.ftpCmdType = ""
							if (b.ftpLocalDigest != ftpRemoteDigest) {
								fr := &ftpResult {
									err: errors.New(fmt.Sprintf("error digest mismatch (local %s, remote %s)", b.ftpLocalDigest, ftpRemoteDigest)),
								}
								b.ftpResChan <- fr
							} else {
								fr := &ftpResult {
									result: b.ftpResult,
								}
								b.ftpResChan <- fr
							}
						}
					}
				} else if b.ftpCmdType == "GET" || b.ftpCmdType == "MD5 OK" {
					if data[0] == 2 && b.ftpState == 0 {
						msg := string(data[1:])
						if strings.HasPrefix(msg, "MD5") {
							b.ftpResult = b.ftpBuffer
							b.ftpBuffer = make([]byte, 0, len(b.ftpResult))
							b.ftpBuffer = append(b.ftpBuffer, data[1:]...)
							b.ftpState = 1
							b.ftpLocalDigest = fmt.Sprintf("%02x", md5.Sum([]byte(b.ftpResult)))
						} else {
							b.ftpBuffer = append(b.ftpBuffer, data[1:]...)
						}
					} else if b.ftpState == 0 {
						b.ftpBuffer = append(b.ftpBuffer, data[1:]...)
					} else if b.ftpState == 1 {
						b.ftpBuffer = append(b.ftpBuffer, data[1:]...)
						ftpRemoteDigest := string(b.ftpBuffer)[3:]
						b.ftpBuffer = b.ftpBuffer[:0]
						if (b.ftpLocalDigest != ftpRemoteDigest) {
							b.ftpState = 0
							b.ftpCmdType = ""
							fr := &ftpResult {
								err: errors.New(fmt.Sprintf("error digest mismatch1 (local %s, remote %s)", b.ftpLocalDigest, ftpRemoteDigest)),
							}
							b.ftpResChan <- fr
						} else {
							b.ftpState = 2
							ftpCmd := &ftpCommand {
								cmd: "MD5 OK",
							}
							b.ftpReqChan <- ftpCmd
						}
					} else if b.ftpState == 2 {
						if data[0] != 3 {
							b.ftpBuffer = b.ftpBuffer[:0]
							b.ftpState = 0
							b.ftpCmdType = ""
							fr := &ftpResult {
								err: errors.New(fmt.Sprintf("unexpected message type (%d)\n", data[0])),
							}
							b.ftpResChan <- fr
						} else {
							msg := string(data[1:])
							if strings.HasPrefix(msg, "End of Transfer") {
								b.ftpBuffer = b.ftpBuffer[:0]
								b.ftpState = 3
							} else {
								b.ftpBuffer = b.ftpBuffer[:0]
								b.ftpState = 0
								b.ftpCmdType = ""
								fr := &ftpResult {
									err: errors.New(fmt.Sprintf("unexpected message (%s)\n", msg)),
									result: b.ftpResult,
								}
								b.ftpResChan <- fr
							}
						}
					} else if b.ftpState == 3 {
						b.ftpBuffer = append(b.ftpBuffer, data[1:]...)
						if data[0] == 1 {
							ftpRemoteDigest := string(b.ftpBuffer)
							b.ftpBuffer = b.ftpBuffer[:0]
							b.ftpState = 0
							b.ftpCmdType = ""
							if (b.ftpLocalDigest != ftpRemoteDigest) {
								fr := &ftpResult {
									err: errors.New(fmt.Sprintf("error digest mismatch2 (local %s, remote %s)", b.ftpLocalDigest, ftpRemoteDigest)),
									result: b.ftpResult,
								}
								b.ftpResChan <- fr
							} else {
								fr := &ftpResult {
									result: b.ftpResult,
								}
								b.ftpResChan <- fr
							}
						}
					}
				}
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
