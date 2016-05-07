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
				drone.SetMaxAltitude(3)
				drone.SetMaxTilt(15)
				drone.SetMaxVirticalSpeed(1)
				drone.SetMaxRotationSpeed(180)
				cmd := drone.NewCommander()
				cmd.ContinuousMode(true)
				time.Sleep(time.Duration(1 * time.Second))
				cmd.FrontFlip()
				time.Sleep(time.Duration(5 * time.Second))
				cmd.RightFlip()
				time.Sleep(time.Duration(5 * time.Second))
				cmd.LeftFlip()
				time.Sleep(time.Duration(5 * time.Second))
				cmd.BackFlip()
				time.Sleep(time.Duration(5 * time.Second))
				cmd.TakePicture()
				time.Sleep(time.Duration(5 * time.Second))
				cmd.Headlight(10, 10)
				time.Sleep(time.Duration(2 * time.Second))
				cmd.Headlight(50, 50)
				time.Sleep(time.Duration(2 * time.Second))
				cmd.Headlight(100, 100)
				time.Sleep(time.Duration(2 * time.Second))
				cmd.Headlight(150, 150)
				time.Sleep(time.Duration(2 * time.Second))
				cmd.Headlight(200, 200)
				time.Sleep(time.Duration(2 * time.Second))
				cmd.Headlight(0, 0)
				time.Sleep(time.Duration(2 * time.Second))
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
