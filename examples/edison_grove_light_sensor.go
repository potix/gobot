package main

import (
	"fmt"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/gpio"
	"github.com/potix/gobot/platforms/intel-iot/edison"
)

func main() {
	gbot := gobot.NewGobot()

	board := edison.NewEdisonAdaptor("board")
	sensor := gpio.NewGroveLightSensorDriver(board, "sensor", "0")

	work := func() {
		gobot.On(sensor.Event("data"), func(data interface{}) {
			fmt.Println("sensor", data)
		})
	}

	robot := gobot.NewRobot("sensorBot",
		[]gobot.Connection{board},
		[]gobot.Device{sensor},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
