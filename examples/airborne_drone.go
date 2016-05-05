package main

import (
	"time"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/airborne-drone"
)

func main() {
	gbot := gobot.NewGobot()

	airborneDroneAdaptor := airbornedrone.NewAirborneDroneAdaptor("swat", "E0:14:0A:BF:3D:80")
	drone := airbornedrone.NewAirborneDroneDriver(airborneDroneAdaptor, "swat")

	work := func() {
		gobot.On(drone.Event("flying"), func(data interface{}) {
			gobot.After(3*time.Second, func() {
				cmd := drone.NewCommander()
				cmd.Up(20).Right(20).Forward(20).Do(time.Duration(500 * time.Millisecond))
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Left(40).Forward(20).Do(time.Duration(500 * time.Millisecond))
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Down(20).Right(20).Backward(40).Do(time.Duration(500 * time.Millisecond))
				time.Sleep(time.Duration(1 * time.Second))
				drone.Landing()
			})
		})
		drone.TakeOff()
	}

	robot := gobot.NewRobot("drone",
		[]gobot.Connection{airborneDroneAdaptor},
		[]gobot.Device{drone},
		work,
	)
	gbot.AddRobot(robot)

	gbot.Start()
}
