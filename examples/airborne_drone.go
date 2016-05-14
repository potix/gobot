package main

import (
	"fmt"
	"time"
        "html"
	"net/http"

	"github.com/potix/gobot"
	"github.com/potix/gobot/api"
	"github.com/potix/gobot/platforms/airborne-drone"
)

func main() {
	gbot := gobot.NewGobot()

	a := api.NewAPI(gbot)
        a.AddHandler(func(w http.ResponseWriter, r *http.Request) {
                fmt.Fprintf(w, "Hello, %q \n", html.EscapeString(r.URL.Path))
        })
        a.Debug()
        a.Start()

	airborneDroneAdaptor := airbornedrone.NewAirborneDroneAdaptor("swat", "E0:14:0A:BF:3D:80", "/var/tmp/airborn_drone")
	drone := airbornedrone.NewAirborneDroneDriver(airborneDroneAdaptor, "drone")

	work := func() {
/*
		gobot.On(drone.Event("flying"), func(data interface{}) {
			gobot.After(3*time.Second, func() {
				m := drone.NewMultiplexer()
				m.Stop()
				drone.FrontFlip()
				time.Sleep(time.Duration(8 * time.Second))
				drone.HeadlightBlinkStart()
				m.Backward(10).Exec(time.Duration(2 * time.Second)).Stop()
				time.Sleep(time.Duration(1 * time.Second))
				drone.HeadlightBlinkStop()
				drone.Headlight(128, 0)
				m.Left(10).Up(10).Exec(time.Duration(2 * time.Second)).Stop()
				time.Sleep(time.Duration(1 * time.Second))
				drone.Headlight(0, 128)
				m.Right(10).Down(10).Exec(time.Duration(2 * time.Second)).Stop()
				time.Sleep(time.Duration(1 * time.Second))
				drone.Headlight(0, 0)
				m.TurnLeft(50).Exec(time.Duration(2 * time.Second)).Stop()
				time.Sleep(time.Duration(1 * time.Second))
				m.TurnRight(50).Exec(time.Duration(2 * time.Second)).Stop()
				time.Sleep(time.Duration(1 * time.Second))
				drone.Landing()
			})
		})
		drone.SetMaxAltitude(2.5)
		drone.SetMaxTilt(15)
		drone.SetMaxVirticalSpeed(0.5)
		drone.SetMaxRotationSpeed(180)
		drone.SetCutOutMode(false)
		drone.SetAutoDownloadMode(true)
		drone.SetContinuousMode(false)
		drone.TakeOff()
*/
		for  {
			time.Sleep(time.Duration(1 * time.Second))
		}
	}

	robot := gobot.NewRobot("airbone_drone",
		[]gobot.Connection{airborneDroneAdaptor},
		[]gobot.Device{drone},
		work,
	)
	gbot.AddRobot(robot)

	robot.AddCommand("hi_there", func(params map[string]interface{}) interface{} {
		return fmt.Sprintf("This command is attached to the robot %v", robot.Name)
	})

	gbot.Start()
}
