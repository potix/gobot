package main

import (
	"fmt"

	"github.com/potix/gobot"
	"github.com/potix/gobot/platforms/spark"
)

func main() {
	gbot := gobot.NewGobot()

	sparkCore := spark.NewSparkCoreAdaptor("spark", "DEVICE_ID", "ACCESS_TOKEN")

	work := func() {
		if result, err := sparkCore.Function("brew", "202,230"); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("result from \"brew\":", result)
		}
	}

	robot := gobot.NewRobot("spark",
		[]gobot.Connection{sparkCore},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
