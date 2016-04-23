package main

import (
	"github.com/potix/gobot"
	"github.com/potix/gobot/api"
	"github.com/potix/gobot/platforms/digispark"
	"github.com/potix/gobot/platforms/gpio"
)

func main() {
	gbot := gobot.NewGobot()

	api.NewAPI(gbot).Start()

	digisparkAdaptor := digispark.NewDigisparkAdaptor("Digispark")
	led := gpio.NewLedDriver(digisparkAdaptor, "led", "0")

	robot := gobot.NewRobot("digispark",
		[]gobot.Connection{digisparkAdaptor},
		[]gobot.Device{led},
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
