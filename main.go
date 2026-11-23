package main

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/schollz/progressbar/v3"
)

// --- CONFIG ---
// Put a SAFE point inside the iPhone Mirroring window that does nothing harmful.
// Example: a blank area / wallpaper area / status area.
const (
	TargetX         = 2540
	TargetY         = 791
	IntervalSeconds = 60
	MoveDuration    = 200 * time.Millisecond
	ActionPause     = 200 * time.Millisecond // Equivalent to pyautogui.PAUSE = 0.2
	Failsafe        = true                   // Move mouse to top-left corner (0, 0) to abort.
)

var errFailsafe = errors.New("failsafe triggered")

func checkFailsafe() error {
	if !Failsafe {
		return nil
	}
	x, y := robotgo.Location()
	if x <= 0 && y <= 0 {
		return errFailsafe
	}
	return nil
}

// moveTo replicates pyautogui.moveTo(..., duration=MOVE_DURATION) + pyautogui.PAUSE
func moveTo(targetX, targetY int, duration time.Duration) error {
	if err := checkFailsafe(); err != nil {
		return err
	}

	startX, startY := robotgo.Location()
	const steps = 20
	stepSleep := duration / steps

	for i := 1; i <= steps; i++ {
		if err := checkFailsafe(); err != nil {
			return err
		}
		progress := float64(i) / float64(steps)
		nx := startX + int(math.Round(float64(targetX-startX)*progress))
		ny := startY + int(math.Round(float64(targetY-startY)*progress))
		robotgo.Move(nx, ny)
		time.Sleep(stepSleep)
	}

	time.Sleep(ActionPause)
	return checkFailsafe()
}

// click replicates pyautogui.click() + pyautogui.PAUSE
func click() error {
	if err := checkFailsafe(); err != nil {
		return err
	}
	robotgo.Click("left", false)
	time.Sleep(ActionPause)
	return checkFailsafe()
}

// waitWithProgress replicates tqdm.trange(..., leave=False) while polling failsafe
func waitWithProgress(seconds int) error {
	bar := progressbar.NewOptions(
		seconds,
		progressbar.OptionSetDescription("Next keep-alive click"),
		progressbar.OptionSetItsString("s"),
		progressbar.OptionShowIts(),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(), // Equivalent to leave=False
	)

	const checksPerSecond = 10
	for i := 0; i < seconds; i++ {
		for j := 0; j < checksPerSecond; j++ {
			if err := checkFailsafe(); err != nil {
				_ = bar.Clear()
				return err
			}
			time.Sleep(time.Second / checksPerSecond)
		}
		_ = bar.Add(1)
	}
	return nil
}

func keepAlive() {
	origX, origY := robotgo.Location()
	fmt.Printf("%d, %d\n", origX, origY)
	fmt.Println("Starting keep-alive loop. Move mouse to top-left corner to stop.")

	for {
		origX, origY = robotgo.Location()

		if err := moveTo(TargetX, TargetY, MoveDuration); errors.Is(err, errFailsafe) {
			fmt.Println("Failsafe triggered. Exiting.")
			break
		}
		if err := click(); errors.Is(err, errFailsafe) {
			fmt.Println("Failsafe triggered. Exiting.")
			break
		}
		if err := moveTo(origX, origY, MoveDuration); errors.Is(err, errFailsafe) {
			fmt.Println("Failsafe triggered. Exiting.")
			break
		}

		fmt.Printf("Clicked at (%d, %d)\n", TargetX, TargetY)

		if err := waitWithProgress(IntervalSeconds); errors.Is(err, errFailsafe) {
			fmt.Println("Failsafe triggered. Exiting.")
			break
		}
	}
}

func main() {
	keepAlive()
}
