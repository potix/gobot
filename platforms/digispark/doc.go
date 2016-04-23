/*
Package digispark provides the Gobot adaptor for the Digispark ATTiny-based USB development board.

Installing:

This package requires installing `libusb`.
Then you can install the package with:

	go get github.com/potix/gobot/platforms/digispark

Example:

	package main

	import (
		"time"

		"github.com/potix/gobot"
		"github.com/potix/gobot/platforms/digispark"
		"github.com/potix/gobot/platforms/gpio"
	)

	func main() {
		gbot := gobot.NewGobot()

		digisparkAdaptor := digispark.NewDigisparkAdaptor("Digispark")
		led := gpio.NewLedDriver(digisparkAdaptor, "led", "0")

		work := func() {
			gobot.Every(1*time.Second, func() {
				led.Toggle()
			})
		}

		robot := gobot.NewRobot("blinkBot",
			[]gobot.Connection{digisparkAdaptor},
			[]gobot.Device{led},
			work,
		)

		gbot.AddRobot(robot)

		gbot.Start()
	}

For further information refer to digispark README:
https://github.com/potix/gobot/blob/master/platforms/digispark/README.md
*/
package digispark
