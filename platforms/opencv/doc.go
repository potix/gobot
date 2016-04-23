/*
Packge opencv contains the Gobot drivers for opencv.

Installing:

This package requires `opencv` to be installed on your system

Then you can install the package with:

	go get github.com/potix/gobot && go install github.com/potix/gobot/platforms/opencv

Example:

	package main

	import (
		cv "github.com/hybridgroup/go-opencv/opencv"
		"github.com/potix/gobot"
		"github.com/potix/gobot/platforms/opencv"
	)

	func main() {
		gbot := gobot.NewGobot()

		window := opencv.NewWindowDriver("window")
		camera := opencv.NewCameraDriver("camera", 0)

		work := func() {
			gobot.On(camera.Event("frame"), func(data interface{}) {
				window.ShowImage(data.(*cv.IplImage))
			})
		}

		robot := gobot.NewRobot("cameraBot",
			[]gobot.Device{window, camera},
			work,
		)

		gbot.AddRobot(robot)

		gbot.Start()
	}

For further information refer to opencv README:
https://github.com/potix/gobot/blob/master/platforms/opencv/README.md
*/
package opencv
