package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		panic(fmt.Errorf("failed to connect to system bus: %v", err))
	}
	defer conn.Close()

	err = conn.AddMatchSignal(
		dbus.WithMatchInterface("net.reactivated.Fprint.Device"),
		dbus.WithMatchMember("VerifyStatus"),
	)
	if err != nil {
		panic(fmt.Errorf("failed to add match rule: %v", err))
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)

	fmt.Println("Started listening for fprint events.")

	for v := range c {
		if len(v.Body) > 0 {
			result, ok := v.Body[0].(string)

			if ok {
				switch result {
				case "verify-match":
					fmt.Println("Fingerprint matched.")
					continue

				case "verify-no-match":
					fmt.Println("Fingerprint was rejected.")

					now := time.Now().Format(time.TimeOnly)

					cmd := exec.Command("notify-send", "-u", "critical", "-a", "Fingerprint", "Someone failed to authenticate.", "Failed fingerprint authentication attempt at "+now)
					if err := cmd.Start(); err != nil {
						fmt.Println("Failed to launch notification:", err)
					}
					continue

				case "verify-retry-scan":
					fmt.Println("Fingerprint not rejected or matched, poor recording.")
					continue
				}
			}

			fmt.Printf("[!] unhandled fprint signal or unexpected data: %v\n", v.Body[0])
		}
	}
}
