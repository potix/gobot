package main

import (
	"fmt"

	"github.com/potix/gobot"
	"github.com/potix/gobot/api"
	"github.com/potix/gobot/platforms/pebble"
)

func main() {
	gbot := gobot.NewGobot()
	a := api.NewAPI(gbot)
	a.Port = "8080"
	a.Start()

	pebbleAdaptor := pebble.NewPebbleAdaptor("pebble")
	pebbleDriver := pebble.NewPebbleDriver(pebbleAdaptor, "pebble")

	work := func() {
		gobot.On(pebbleDriver.Event("accel"), func(data interface{}) {
			fmt.Println(data.(string))
		})
	}

	robot := gobot.NewRobot("pebble",
		[]gobot.Connection{pebbleAdaptor},
		[]gobot.Device{pebbleDriver},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
