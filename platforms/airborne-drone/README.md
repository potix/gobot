# Ardrone

The Airbrone Drone from Parrot is an inexpensive quadcopter that is controlled using bluetooth. It includes a built-in bottom facing camera.

For more info about the  Airborne drone platform click [here](http://www.parrot.com/jp/products/minidrones/)

## How to Install
```
go get -d -u github.com/potix/gobot/... && go install github.com/potix/gobot/platforms/airborne_drone
```
## How to Use
```go
package main

import (
        "time"

        "github.com/potix/gobot"
        "github.com/potix/gobot/platforms/airborne_drone"
)

func main() {
        gbot := gobot.NewGobot()

        airborneDroneAdaptor := airborneDrone.NewAirborneDroneAdaptor("swat", "01:23:45:67:89:AB", "hci0")
        drone := airborneDrone.NewAirborneDroneDriver(airborneDroneAdaptor, "swat")

        work := func() {
                gobot.On(drone.Event("flying"), func(data interface{}) {
                        gobot.After(3*time.Second, func() {
                                drone.Land()
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



## How to Connect

The Airborne Drone is a bluetooth device.
Therefore, you must obtain the MAC address of the drone in advance.

### get the MAC address of the drone
'''
sudo hcitool lescan
LE Scan ...
01:23:45:67:89:AB Swat_012345
'''
