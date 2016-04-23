package main

import (
	"fmt"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/leap"
)

func main() {
	gbot := gobot.NewGobot()

	leapMotionAdaptor := leap.NewLeapMotionAdaptor("leap", "127.0.0.1:6437")
	l := leap.NewLeapMotionDriver(leapMotionAdaptor, "leap")

	work := func() {
		gobot.On(l.Event("hand"), func(data interface{}) {
			printHand(data.(leap.Hand))
		})
	}

	robot := gobot.NewRobot("leapBot",
		[]gobot.Connection{leapMotionAdaptor},
		[]gobot.Device{l},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}

func printHand(hand leap.Hand) {
	fmt.Println("Hand", hand)
}
