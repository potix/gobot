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
				cmd.ContinuousMode(true)
				cmd.Up(20).Do(time.Duration(25 * time.Millisecond))
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Stop()
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Right(40).Do(time.Duration(25 * time.Millisecond))
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Stop()
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Left(40).Do(time.Duration(25 * time.Millisecond))
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Stop()
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Forward(40).Do(time.Duration(25 * time.Millisecond))
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Stop()
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Backward(40).Do(time.Duration(25 * time.Millisecond))
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Stop()
				time.Sleep(time.Duration(1 * time.Second))
				cmd.TurnRight(100).Do(time.Duration(25 * time.Millisecond))
				time.Sleep(time.Duration(2 * time.Second))
				cmd.Stop()
				time.Sleep(time.Duration(1 * time.Second))
				cmd.TurnLeft(100).Do(time.Duration(25 * time.Millisecond))
				time.Sleep(time.Duration(2 * time.Second))
				cmd.Stop()
				time.Sleep(time.Duration(1 * time.Second))
				cmd.Down(20).Do(time.Duration(25 * time.Millisecond))
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
