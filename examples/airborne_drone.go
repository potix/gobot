package main

import (
	"fmt"
	"time"
	"reflect"

	"github.com/potix/gobot"
	"github.com/potix/gobot/api"
	"github.com/potix/gobot/platforms/airborne-drone"
)

func main() {
	gbot := gobot.NewGobot()

	a := api.NewAPI(gbot)
        a.Debug()
        a.Start()

	airborneDroneAdaptor := airbornedrone.NewAirborneDroneAdaptor("airbone_drone", "E0:14:0A:BF:3D:80", "/var/tmp/airborn_drone")
	drone := airbornedrone.NewAirborneDroneDriver(airborneDroneAdaptor, "airbone_drone_swat")

	work := func() {
		gobot.On(drone.Event("flying"), func(data interface{}) {
			fmt.Println("flying")
		})
		drone.SetMaxAltitude(5)
		drone.SetMaxTilt(15)
		drone.SetMaxVirticalSpeed(1.225)
		drone.SetMaxRotationSpeed(205)
		drone.SetCutOutMode(true)
		drone.SetAutoDownloadMode(true)
		drone.SetContinuousMode(true)
	}

	robot := gobot.NewRobot("airbone_drone_swat_01",
		[]gobot.Connection{airborneDroneAdaptor},
		[]gobot.Device{drone},
		work,
	)

	gbot.AddRobot(robot)

	robot.AddCommand("take_off", func(params map[string]interface{}) interface{} {
		drone.TakeOff()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("landing", func(params map[string]interface{}) interface{} {
		drone.Landing()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("front_flip", func(params map[string]interface{}) interface{} {
		drone.FrontFlip()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("back_flip", func(params map[string]interface{}) interface{} {
		drone.BackFlip()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("right_flip", func(params map[string]interface{}) interface{} {
		drone.RightFlip()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("left_flip", func(params map[string]interface{}) interface{} {
		drone.LeftFlip()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("max_altitude", func(params map[string]interface{}) interface{} {
		var ok bool
		vi, ok := params["value"]
		if !ok {
			return fmt.Sprintf("no value %v", robot.Name)
		}
		v, ok := vi.(float64)
		if !ok {
			return fmt.Sprintf("value is not float %v, value %v", robot.Name, v)
		}
		if v < 2 || v > 10 {
			return fmt.Sprintf("value is out of range %v, value %v", robot.Name, v)
		}
		drone.SetMaxAltitude(float32(v))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("max_tilt", func(params map[string]interface{}) interface{} {
		var ok bool
		vi, ok := params["value"]
		if !ok {
			return fmt.Sprintf("no value %v", robot.Name)
		}
		v, ok := vi.(float64)
		if !ok {
			return fmt.Sprintf("value is not float %v, value %v", robot.Name, v)
		}
		if v < 5 || v > 25 {
			return fmt.Sprintf("value is out of range %v, value %v", robot.Name, v)
		}
		drone.SetMaxTilt(float32(v))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("max_virtical_speed", func(params map[string]interface{}) interface{} {
		var ok bool
		vi, ok := params["value"]
		if !ok {
			return fmt.Sprintf("no value %v", robot.Name)
		}
		v, ok := vi.(float64)
		if !ok {
			return fmt.Sprintf("value is not float %v, value %v", robot.Name, v)
		}
		if v < 0.5 || v > 2 {
			return fmt.Sprintf("value is out of range %v, value %v", robot.Name, v)
		}
		drone.SetMaxVirticalSpeed(float32(v))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("max_rotation_speed", func(params map[string]interface{}) interface{} {
		var ok bool
		vi, ok := params["value"]
		if !ok {
			return fmt.Sprintf("no value %v", robot.Name)
		}
		v, ok := vi.(float64)
		if !ok {
			return fmt.Sprintf("value is not bool %v, value %v", robot.Name, v)
		}
		if v < 50 || v > 360 {
			return fmt.Sprintf("value is out of range %v, value %v", robot.Name, v)
		}
		drone.SetMaxRotationSpeed(float32(v))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("continuous_mode", func(params map[string]interface{}) interface{} {
		var ok bool
		vi, ok := params["value"]
		if !ok {
			return fmt.Sprintf("no value %v", robot.Name)
		}
		v, ok := vi.(bool)
		if !ok {
			return fmt.Sprintf("value is not bool %v, value %v", robot.Name, v)
		}
		drone.SetContinuousMode(v)
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("auto_download_mode", func(params map[string]interface{}) interface{} {
		var ok bool
		vi, ok := params["value"]
		if !ok {
			return fmt.Sprintf("no value %v", robot.Name)
		}
		v, ok := vi.(bool)
		if !ok {
			return fmt.Sprintf("value is not boolt %v, value %v", robot.Name, v)
		}
		drone.SetAutoDownloadMode(v)
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("cut_out_mode", func(params map[string]interface{}) interface{} {
		var ok bool
		vi, ok := params["value"]
		if !ok {
			return fmt.Sprintf("no value %v", robot.Name)
		}
		v, ok := vi.(bool)
		if !ok {
			return fmt.Sprintf("value is not bool %v, value %v", robot.Name, v)
		}
		drone.SetCutOutMode(v)
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("force_download", func(params map[string]interface{}) interface{} {
		drone.ForceDownload()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("flat_trim", func(params map[string]interface{}) interface{} {
		drone.FlatTrim()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("emergency", func(params map[string]interface{}) interface{} {
		drone.Emergency()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("headlight", func(params map[string]interface{}) interface{} {
		var ok bool
		lefti, ok := params["left"]
		if !ok {
			return fmt.Sprintf("no left %v", robot.Name)
		}
		left, ok := lefti.(float64)
		if !ok {
			return fmt.Sprintf("left is not uint8 %v, value %v", robot.Name, left)
		}
		if left < 0 || left > 255 {
			return fmt.Sprintf("left is out of range %v, value %v", robot.Name, left)
		}
		righti, ok := params["right"]
		if !ok {
			return fmt.Sprintf("no right %v", robot.Name)
		}
		right, ok := righti.(float64)
		if !ok {
			return fmt.Sprintf("right is not uint8 %v, value %v", robot.Name, right)
		}
		if right < 0 || right > 255 {
			return fmt.Sprintf("left is out of range %v, value %v", robot.Name, left)
		}
		drone.Headlight(uint8(left), uint8(right))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("headlight_flash", func(params map[string]interface{}) interface{} {
		var ok bool
		vi, ok := params["value"]
		if !ok {
			return fmt.Sprintf("no value %v", robot.Name)
		}
		v, ok := vi.(bool)
		if !ok {
			return fmt.Sprintf("value is not bool %v, value %v", robot.Name, v)
		}
		if v {
			drone.HeadlightFlashStart()
		} else {
			drone.HeadlightFlashStop()
		}
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("headlight_blink", func(params map[string]interface{}) interface{} {
		var ok bool
		vi, ok := params["value"]
		if !ok {
			return fmt.Sprintf("no value %v", robot.Name)
		}
		v, ok := vi.(bool)
		if !ok {
			return fmt.Sprintf("value is not bool %v, value %v", robot.Name, v)
		}
		if v {
			drone.HeadlightBlinkStart()
		} else {
			drone.HeadlightBlinkStop()
		}
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("headlight_oscillation", func(params map[string]interface{}) interface{} {
		var ok bool
		vi, ok := params["value"]
		if !ok {
			return fmt.Sprintf("no value %v", robot.Name)
		}
		v, ok := vi.(bool)
		if !ok {
			return fmt.Sprintf("value is not bool %v, value %v", robot.Name, v)
		}
		if v {
			drone.HeadlightOscillationStart()
		} else {
			drone.HeadlightOscillationStop()
		}
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("take_picture", func(params map[string]interface{}) interface{} {
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("battery", func(params map[string]interface{}) interface{} {
		v := drone.GetBattery()
		return fmt.Sprintf("ok %v, battery %v", robot.Name, v)
	})
	robot.AddCommand("flying_state", func(params map[string]interface{}) interface{} {
		v := drone.GetBattery()
		return fmt.Sprintf("ok %v, flying state %v", robot.Name, v)
	})
	robot.AddCommand("picture_state", func(params map[string]interface{}) interface{} {
		v := drone.GetBattery()
		return fmt.Sprintf("ok %v, picture state %v", robot.Name, v)
	})
	robot.AddCommand("roll", func(params map[string]interface{}) interface{} {
		var ok bool
		mseci, ok := params["msec"]
		if !ok {
			return fmt.Sprintf("no msec %v", robot.Name)
		}
		msec, ok := mseci.(float64)
		if !ok {
			return fmt.Sprintf("msec is not uint %v, value %v", robot.Name, msec)
		}
		if msec < 0 {
			return fmt.Sprintf("msec is out of range %v, value %v", robot.Name, msec)
		}
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speed is not int8 %v, value %v", robot.Name, speed)
		}
		if speed < -100 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.Roll(time.Duration(msec) * time.Millisecond, int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("pitch", func(params map[string]interface{}) interface{} {
		var ok bool
		mseci, ok := params["msec"]
		if !ok {
			return fmt.Sprintf("no msec %v", robot.Name)
		}
		msec, ok := mseci.(float64)
		if !ok {
			return fmt.Sprintf("msec is not uint %v, value %v", robot.Name, msec)
		}
		if msec < 0 {
			return fmt.Sprintf("msec is out of range %v, value %v", robot.Name, msec)
		}
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speedis not int8 %v, value %v", robot.Name, speed)
		}
		if speed < -100 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.Pitch(time.Duration(msec) * time.Millisecond, int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("yaw", func(params map[string]interface{}) interface{} {
		var ok bool
		mseci, ok := params["msec"]
		if !ok {
			return fmt.Sprintf("no msec %v", robot.Name)
		}
		msec, ok := mseci.(float64)
		if !ok {
			return fmt.Sprintf("msec is not uint %v, value %v", robot.Name, msec)
		}
		if msec < 0 {
			return fmt.Sprintf("msec is out of range %v, value %v", robot.Name, msec)

		}
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speed is not int8 %v, value %v", robot.Name, speed)
		}
		if speed < -100 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.Yaw(time.Duration(msec) * time.Millisecond, int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("gaz", func(params map[string]interface{}) interface{} {
		var ok bool
		mseci, ok := params["msec"]
		if !ok {
			return fmt.Sprintf("no msec %v", robot.Name)
		}
		msec, ok := mseci.(float64)
		if !ok {
			return fmt.Sprintf("msec is not uint %v, value %v", robot.Name, msec)
		}
		if msec < 0 {
			return fmt.Sprintf("msec is out of range %v, value %v", robot.Name, msec)

		}
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speed is not int8 %v, value %v", robot.Name, speed)
		}
		if speed < -100 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.Gaz(time.Duration(msec) * time.Millisecond, int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("hover", func(params map[string]interface{}) interface{} {
		drone.Hover()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("up", func(params map[string]interface{}) interface{} {
		var ok bool
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speed is not int8 %v, value %v", robot.Name, speed)
		}
		if speed < 0 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.Up(int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("down", func(params map[string]interface{}) interface{} {
		var ok bool
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speed is not int8 %v, value %v", robot.Name, speed)
		}
		if speed < 0 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.Down(int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("left", func(params map[string]interface{}) interface{} {
		var ok bool
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speed is not int8 %v, value %v", robot.Name, speed)
		}
		if speed < 0 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.Left(int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("right", func(params map[string]interface{}) interface{} {
		var ok bool
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speedis not int8 %v, value %v", robot.Name, speed)
		}
		if speed < 0 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.Right(int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("forward", func(params map[string]interface{}) interface{} {
		var ok bool
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speed is not int8 %v, value %v", robot.Name, speed)
		}
		if speed < 0 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.Forward(int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("backward", func(params map[string]interface{}) interface{} {
		var ok bool
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speed is not int8 %v, value %v", robot.Name, speed)
		}
		if speed < 0 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.Backward(int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("turn_left", func(params map[string]interface{}) interface{} {
		var ok bool
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}

		fmt.Println(reflect.TypeOf(speedi).Kind())

		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speed is not int8 %v, value %v", robot.Name, speed)
		}
		if speed < 0 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.TurnLeft(int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("turn_right", func(params map[string]interface{}) interface{} {
		var ok bool
		speedi, ok := params["speed"]
		if !ok {
			return fmt.Sprintf("no speed %v", robot.Name)
		}
		speed, ok := speedi.(float64)
		if !ok {
			return fmt.Sprintf("speed is not int8 %v, value %v", robot.Name, speed)
		}
		if speed < 0 || speed > 100 {
			return fmt.Sprintf("speed is out of range %v, value %v", robot.Name, speed)
		}
		drone.TurnRight(int8(speed))
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("stop", func(params map[string]interface{}) interface{} {
		drone.Stop()
		return fmt.Sprintf("ok %v", robot.Name)
	})
	robot.AddCommand("multi", func(params map[string]interface{}) interface{} {
		var ok bool
		commandsi, ok := params["commands"]
		if !ok {
			return fmt.Sprintf("no commands %v", robot.Name)
		}
		commands , ok := commandsi.([]map[string]interface{})
		if !ok {
			return fmt.Sprintf("invalid commands %v, commands %v", robot.Name, commands)
		}
		m := drone.NewMultiplexer()
		for _, cmd := range commands {
			ti, ok :=  cmd["type"]
			if !ok {
				continue
			}
			t, ok := ti.(string)
			if !ok {
				continue
			}
			switch t {
			case "reset":
				m.Reset()
			case "go":
				mseci, ok := cmd["msec"]
				if !ok {
					break
				}
				msec, ok := mseci.(float64)
				if !ok {
					break
				}
				if msec < 0 {
					break
				}
				m.Go(time.Duration(msec) * time.Millisecond)
			case "up":
				speedi, ok := params["speed"]
				if !ok {
					return fmt.Sprintf("no speed %v", robot.Name)
				}
				speed, ok := speedi.(float64)
				if !ok {
					return fmt.Sprintf("speed is not uint8 %v, value %v", robot.Name, speed)
				}
				m.Up(uint8(speed))
			case "down":
				speedi, ok := params["speed"]
				if !ok {
					return fmt.Sprintf("no speed %v", robot.Name)
				}
				speed, ok := speedi.(float64)
				if !ok {
					return fmt.Sprintf("speed is not uint8 %v, value %v", robot.Name, speed)
				}
				m.Down(uint8(speed))
			case "forward":
				speedi, ok := params["speed"]
				if !ok {
					return fmt.Sprintf("no speed %v", robot.Name)
				}
				speed, ok := speedi.(float64)
				if !ok {
					return fmt.Sprintf("speed is not uint8 %v, value %v", robot.Name, speed)
				}
				m.Forward(uint8(speed))
			case "backward":
				speedi, ok := params["speed"]
				if !ok {
					return fmt.Sprintf("no speed %v", robot.Name)
				}
				speed, ok := speedi.(float64)
				if !ok {
					return fmt.Sprintf("speed is not uint8 %v, value %v", robot.Name, speed)
				}
				m.Backward(uint8(speed))
			case "left":
				speedi, ok := params["speed"]
				if !ok {
					return fmt.Sprintf("no speed %v", robot.Name)
				}
				speed, ok := speedi.(float64)
				if !ok {
					return fmt.Sprintf("speed is not uint8 %v, value %v", robot.Name, speed)
				}
				m.Left(uint8(speed))
			case "right":
				speedi, ok := params["speed"]
				if !ok {
					return fmt.Sprintf("no speed %v", robot.Name)
				}
				speed, ok := speedi.(float64)
				if !ok {
					return fmt.Sprintf("speed is not uint8 %v, value %v", robot.Name, speed)
				}
				m.Right(uint8(speed))
			case "turnLeft":
				speedi, ok := params["speed"]
				if !ok {
					return fmt.Sprintf("no speed %v", robot.Name)
				}
				speed, ok := speedi.(float64)
				if !ok {
					return fmt.Sprintf("speed is not uint8 %v, value %v", robot.Name, speed)
				}
				m.TurnLeft(uint8(speed))
			case "turnRight":
				speedi, ok := params["speed"]
				if !ok {
					return fmt.Sprintf("no speed %v", robot.Name)
				}
				speed, ok := speedi.(float64)
				if !ok {
					return fmt.Sprintf("speed is not uint8 %v, value %v", robot.Name, speed)
				}
				m.TurnRight(uint8(speed))
			}
		}
		return fmt.Sprintf("ok %v", robot.Name)
	})

	gbot.Start()
}
