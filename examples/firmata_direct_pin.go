package main

import (
	"time"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/firmata"
	"github.com/potix/gobot/platforms/gpio"
)

func main() {
	gbot := gobot.NewGobot()

	firmataAdaptor := firmata.NewFirmataAdaptor("myFirmata", "/dev/ttyACM0")
	pin := gpio.NewDirectPinDriver(firmataAdaptor, "pin", "13")

	work := func() {
		level := byte(1)

		gobot.Every(1*time.Second, func() {
			pin.DigitalWrite(level)
			if level == 1 {
				level = 0
			} else {
				level = 1
			}
		})
	}

	robot := gobot.NewRobot("pinBot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{pin},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
