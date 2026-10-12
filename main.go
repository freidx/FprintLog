package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	maxConsecutiveFailures = 3
	graceWindow            = 10 * time.Second
)

func notify(summary, body string) {
	cmd := exec.Command("notify-send", "-u", "critical", "-a", "Fingerprint", summary, body)
	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to launch notification:", err)
		return
	}
	go cmd.Wait()
}

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

	var (
		failures    int
		lastFailure time.Time
		deadline    <-chan time.Time
	)

	reset := func() {
		failures = 0
		deadline = nil
	}

	for {
		select {
		case v, ok := <-c:
			if !ok {
				return
			}
			if len(v.Body) == 0 {
				continue
			}

			result, ok := v.Body[0].(string)
			if !ok {
				fmt.Printf("[!] unexpected data: %v\n", v.Body[0])
				continue
			}

			switch result {
			case "verify-match":
				fmt.Println("Fingerprint matched.")
				reset()

			case "verify-no-match":
				failures++
				lastFailure = time.Now()
				fmt.Printf("Fingerprint was rejected (%d in a row).\n", failures)

				if failures >= maxConsecutiveFailures {
					notify(
						"Someone failed to authenticate.",
						fmt.Sprintf("%d failed fingerprint attempts in a row (last at %s)",
							failures, lastFailure.Format(time.TimeOnly)),
					)
					reset()
					continue
				}

				deadline = time.After(graceWindow)

			case "verify-retry-scan":
				fmt.Println("Fingerprint not rejected or matched, poor recording.")

			default:
				fmt.Printf("[!] unhandled fprint signal: %v\n", result)
			}

		case <-deadline:
			if failures == 1 {
				notify(
					"Someone failed to authenticate.",
					fmt.Sprintf("Single failed fingerprint attempt, no successful login within %s (attempted at %s)", graceWindow, lastFailure.Format(time.TimeOnly)),
				)
			} else {
				notify(
					"Someone failed to authenticate.",
					fmt.Sprintf("%d failed fingerprint attempt, no successful login within %s (last at %s)", failures, graceWindow, lastFailure.Format(time.TimeOnly)),
				)
			}

			reset()
		}
	}
}
