/*
Package beaglebone provides the Gobot adaptor for the Beaglebone Black.

Installing:

	go get github.com/hybridgroup/platforms/gobot/beaglebone

Example:

	package main

	import (
		"time"

		"github.com/potix/gobot"
		"github.com/potix/gobot/platforms/beaglebone"
		"github.com/potix/gobot/platforms/gpio"
	)

	func main() {
		gbot := gobot.NewGobot()

		beagleboneAdaptor := beaglebone.NewBeagleboneAdaptor("beaglebone")
		led := gpio.NewLedDriver(beagleboneAdaptor, "led", "P9_12")

		work := func() {
			gobot.Every(1*time.Second, func() {
				led.Toggle()
			})
		}

		robot := gobot.NewRobot("blinkBot",
			[]gobot.Connection{beagleboneAdaptor},
			[]gobot.Device{led},
			work,
		)

		gbot.AddRobot(robot)

		gbot.Start()
	}

For more information refer to the beaglebone README:
https://github.com/potix/gobot/blob/master/platforms/beaglebone/README.md
*/
package beaglebone
