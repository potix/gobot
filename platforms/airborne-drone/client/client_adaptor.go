package client

import (
	"fmt"
	"log"
	"strings"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/airborne-drone/client/service"
	"github.com/potix/gatt"
)

// TODO: handle other OS defaults besides Linux
var DefaultClientOptions = []gatt.Option{
	gatt.LnxMaxConnections(1),
	gatt.LnxDeviceID(-1, false),
}

var _ gobot.Adaptor = (*Adaptor)(nil)

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
}

// NewAdaptor returns a new Adaptor given a name and uuid
func NewAdaptor(name string, uuid string) *Adaptor {
	return &Adaptor{
		name:      name,
		uuid:      uuid,
		connected: false,
		peripheralReady: make(chan error),
		services: make(map[string]*BLEService),
	}
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
	if err = <- b.peripheralReady; err != nil {
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

	return nil
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
		d.AddService(service.NewGattService()) 
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
	return nil
}

// Represents a BLE Peripheral's Service
type BLEService struct {
	uuid            	string
	service        		*gatt.Service
	characteristics 	map[string]*gatt.Characteristic
}

// NewAdaptor returns a new BLEService given a uuid
func NewBLEService(sUuid string, service *gatt.Service) *BLEService {
	return &BLEService{
		uuid:      sUuid,
		service: 	 service,
		characteristics: make(map[string]*gatt.Characteristic),
	}
}
