package main

import (
	"os"

	"github.com/urfave/cli"
	"github.com/zach-klippenstein/goadb"
	"log"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"regexp"
	"time"
)

var (
	client *adb.Adb
	device *adb.Device
)

func main() {
	app := cli.NewApp()
	app.Name = "monkey-madness"
	app.Usage = "stress testing Android apps"
	app.Version = "1.0.0"
	app.Authors = []cli.Author{
		{
			Name:  "Radim Vaculik",
			Email: "radim.vaculik@thefuntasty.com",
		},
	}
	cli.OsExiter = func(c int) {
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "output, o",
			Usage: "Output directory",
		},
		cli.StringFlag{
			Name: "serial, s",
			Usage: "Device serial number",
		},
		cli.BoolFlag{
			Name: "quiet, q",
		},
	}

	app.Action = func(c *cli.Context) error {
		client, _ = adb.NewWithConfig(adb.ServerConfig{
			Port: adb.AdbPort,
		})

		err := client.StartServer()
		if err != nil {
			log.Fatal(err)
		}

		if (c.String("s") != "") {
			device = client.Device(adb.DeviceWithSerial(c.String("s")))
		} else {
			serials, err := client.ListDeviceSerials()
			if err != nil {
				log.Fatal(err)
			}

			if len(serials) > 1 {
				return errors.New("More than one device detected!")
			}

			if len(serials) == 0 {
				return errors.New("No connected device!")
			}

			for _, serial := range serials {
				device = client.Device(adb.DeviceWithSerial(serial))
			}

		}

		output, _ := device.RunCommand("getprop ro.build.version.sdk")
		deviceSdk, _ := strconv.Atoi(strings.Replace(output, "\n", "", 1))

		if deviceSdk < 21 {
			return errors.New("Device with API 21+ supported only!")
		}

		b := true;
		for {
			output, err = device.RunCommand("dumpsys window windows | grep mCurrentFocus")
			if strings.Contains(output, "mCurrentFocus=null") || strings.Contains(output, "StatusBar") {
				time.Sleep(200 * time.Millisecond)
				if b {
					fmt.Print("Turn the screen on and unlock the device to continue...\n")
					b = false
				}
			} else {
				break
			}
		}

		rp := regexp.MustCompile(`mCurrentFocus=Window{[a-f0-9]+ u0 ([a-zA-Z0-9\.]+)/[a-zA-Z0-9\.]+}`)
		appName := rp.FindAllStringSubmatch(output, -1)[0][1]

		fmt.Printf("Testing %s\n", appName)

		client.KillServer()

		return nil
	}

	app.Run(os.Args)
}
