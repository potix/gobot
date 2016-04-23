package main

import (
	"time"

	"github.com/potix/gobot"
	"github.com/potix/gobot/api"
	"github.com/potix/gobot/platforms/gpio"
	"github.com/potix/gobot/platforms/spark"
)

func main() {
	gbot := gobot.NewGobot()
	api.NewAPI(gbot).Start()

	sparkCore := spark.NewSparkCoreAdaptor("spark", "device_id", "access_token")
	led := gpio.NewLedDriver(sparkCore, "led", "D7")

	work := func() {
		gobot.Every(1*time.Second, func() {
			led.Toggle()
		})
	}

	robot := gobot.NewRobot("spark",
		[]gobot.Connection{sparkCore},
		[]gobot.Device{led},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
