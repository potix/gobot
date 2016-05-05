package service

import (
	"fmt"
	"github.com/potix/gatt"
)

var (
	attrGATTUUID   = gatt.UUID16(0x1801)
	attrDeviceName = gatt.UUID16(0x2A00)
)

func NewGattGapService() *gatt.Service {
	s := gatt.NewService(attrGATTUUID)
	s.AddCharacteristic(attrDeviceName).HandleReadFunc(
		func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
			v := []byte{ 'g', 'o', 'b', 'o', 't' }
			fmt.Println("$ Device Name Read Handler")
			_, error := rsp.Write(v)
			if error != nil {
				fmt.Println(error)
			}
		})
	return s
}

