package main

import (
	"fmt"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/gpio"
	"github.com/potix/gobot/platforms/intel-iot/edison"
)

func main() {
	gbot := gobot.NewGobot()

	e := edison.NewEdisonAdaptor("edison")
	button := gpio.NewGroveButtonDriver(e, "button", "2")

	work := func() {
		gobot.On(button.Event(gpio.Push), func(data interface{}) {
			fmt.Println("On!")
		})

		gobot.On(button.Event(gpio.Release), func(data interface{}) {
			fmt.Println("Off!")
		})

	}

	robot := gobot.NewRobot("bot",
		[]gobot.Connection{e},
		[]gobot.Device{button},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
