package main

import (
	"fmt"
	"time"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/digispark"
	"github.com/potix/gobot/platforms/gpio"
)

func main() {
	gbot := gobot.NewGobot()

	digisparkAdaptor := digispark.NewDigisparkAdaptor("digispark")
	servo := gpio.NewServoDriver(digisparkAdaptor, "servo", "0")

	work := func() {
		gobot.Every(1*time.Second, func() {
			i := uint8(gobot.Rand(180))
			fmt.Println("Turning", i)
			servo.Move(i)
		})
	}

	robot := gobot.NewRobot("servoBot",
		[]gobot.Connection{digisparkAdaptor},
		[]gobot.Device{servo},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
