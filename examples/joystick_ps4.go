package main

import (
	"fmt"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/joystick"
)

func main() {
	gbot := gobot.NewGobot()

	joystickAdaptor := joystick.NewJoystickAdaptor("ps4")
	joystick := joystick.NewJoystickDriver(joystickAdaptor,
		"ps4",
		"./platforms/joystick/configs/dualshock4.json",
	)

	work := func() {
		gobot.On(joystick.Event("square_press"), func(data interface{}) {
			fmt.Println("square_press")
		})
		gobot.On(joystick.Event("square_release"), func(data interface{}) {
			fmt.Println("square_release")
		})
		gobot.On(joystick.Event("triangle_press"), func(data interface{}) {
			fmt.Println("triangle_press")
		})
		gobot.On(joystick.Event("triangle_release"), func(data interface{}) {
			fmt.Println("triangle_release")
		})
		gobot.On(joystick.Event("circle_press"), func(data interface{}) {
			fmt.Println("circle_press")
		})
		gobot.On(joystick.Event("circle_release"), func(data interface{}) {
			fmt.Println("circle_release")
		})
		gobot.On(joystick.Event("x_press"), func(data interface{}) {
			fmt.Println("x_press")
		})
		gobot.On(joystick.Event("x_release"), func(data interface{}) {
			fmt.Println("x_release")
		})


		gobot.On(joystick.Event("left_x"), func(data interface{}) {
			fmt.Println("left_x", data)
		})
		gobot.On(joystick.Event("left_y"), func(data interface{}) {
			fmt.Println("left_y", data)
		})
		gobot.On(joystick.Event("right_x"), func(data interface{}) {
			fmt.Println("right_x", data)
		})
		gobot.On(joystick.Event("right_y"), func(data interface{}) {
			fmt.Println("right_y", data)
		})
	}

	robot := gobot.NewRobot("joystickBot",
		[]gobot.Connection{joystickAdaptor},
		[]gobot.Device{joystick},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
