package main

import (
	"fmt"
	"time"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/beaglebone"
	"github.com/potix/gobot/platforms/i2c"
)

func main() {
	gbot := gobot.NewGobot()
	beagleboneAdaptor := beaglebone.NewBeagleboneAdaptor("beaglebone")
	blinkm := i2c.NewBlinkMDriver(beagleboneAdaptor, "blinkm")

	work := func() {
		gobot.Every(3*time.Second, func() {
			r := byte(gobot.Rand(255))
			g := byte(gobot.Rand(255))
			b := byte(gobot.Rand(255))
			blinkm.Rgb(r, g, b)
			color, _ := blinkm.Color()
			fmt.Println("color", color)
		})
	}

	robot := gobot.NewRobot("blinkmBot",
		[]gobot.Connection{beagleboneAdaptor},
		[]gobot.Device{blinkm},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
