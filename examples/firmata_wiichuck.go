package main

import (
	"fmt"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/firmata"
	"github.com/potix/gobot/platforms/i2c"
)

func main() {
	gbot := gobot.NewGobot()

	firmataAdaptor := firmata.NewFirmataAdaptor("firmata", "/dev/ttyACM0")
	wiichuck := i2c.NewWiichuckDriver(firmataAdaptor, "wiichuck")

	work := func() {
		gobot.On(wiichuck.Event("joystick"), func(data interface{}) {
			fmt.Println("joystick", data)
		})

		gobot.On(wiichuck.Event("c"), func(data interface{}) {
			fmt.Println("c")
		})

		gobot.On(wiichuck.Event("z"), func(data interface{}) {
			fmt.Println("z")
		})
		gobot.On(wiichuck.Event("error"), func(data interface{}) {
			fmt.Println("Wiichuck error:", data)
		})
	}

	robot := gobot.NewRobot("chuck",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{wiichuck},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
