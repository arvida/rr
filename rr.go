package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	Reset   = "\x1b[0m"
	FgBlack = "\x1b[30m"
	FgWhite = "\x1b[37m"
	BgRed   = "\x1b[41m"
	BgGreen = "\x1b[42m"
)

func main() {

	if len(os.Args) == 1 {
		fmt.Printf("usage %s <command with args to run>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	command := strings.Join(os.Args[1:], " ")
	runs := 0
	fails := 0

	for {
		exitStatus := run(command)

		runs += 1
		if exitStatus != 0 {
			fails += 1
		}
		color := colorTheme(exitStatus)

		fmt.Printf(statusBar(color, runs, fails))

		var input string
		fmt.Scanln(&input)
	}
}

func run(shellCmd string) int {
	exitStatus := 0
	command := exec.Command("/bin/sh", "-c", shellCmd)

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			status := exiterr.Sys().(syscall.WaitStatus)
			exitStatus = status.ExitStatus()
		} else {
			log.Fatalf("err: %q", err)
		}
	}

	return exitStatus
}

func statusBar(color string, runs int, fails int) string {
	cols, err := terminalWidth()
	if err != nil {
		log.Fatalf("err: %q", err)
	}

	status := fmt.Sprintf(" Runs: %d ☂ Fails: %d", runs, fails)
	currentTime := timestamp()
	textWidth := len(status) + len(currentTime)
	spacing := strings.Repeat(" ", cols-textWidth)

	return color + status + spacing + currentTime + " " + Reset
}

func timestamp() string {
	now := time.Now()

	return now.Format(time.Stamp)
}

func colorTheme(exitStatus int) string {
	if exitStatus != 0 {
		return BgRed
	}

	return BgGreen + FgBlack
}

//  Get terminal window size
// Stolen from: 
//  http://stackoverflow.com/questions/1733155/how-to-set-the-terminals-size

const (
	TIOCGWINSZ     = 0x5413
	TIOCGWINSZ_OSX = 1074295912
)

type window struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func terminalWidth() (int, error) {
	w := new(window)
	tio := syscall.TIOCGWINSZ
	if runtime.GOOS == "darwin" {
		tio = TIOCGWINSZ_OSX
	}
	res, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(tio),
		uintptr(unsafe.Pointer(w)),
	)
	if int(res) == -1 {
		return 0, err
	}
	return int(w.Col), nil
}
