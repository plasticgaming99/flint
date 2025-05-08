package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"syscall"
	"time"

	sys "golang.org/x/sys/unix"
)

var (
	noninit bool
)

var (
	logger = log.New(os.Stdout, "FLINT: ", log.Ldate|log.Ltime|log.Lmsgprefix)
	sig    = make(chan os.Signal, 1)
)

const (
	onesec = 1 * time.Second
	ttsec  = 30 * time.Second
)

func main() {
	if len(os.Args) != 1 {
		if slices.Contains(os.Args, "--noninit") {
			noninit = true
		}
	}

	signal.Notify(sig, syscall.SIGUSR1, syscall.SIGINT)

	if !noninit {
		logger.Println("very early init")
		sys.Mount("proc", "proc", "proc", 0, "")
		sys.Mount("sys", "sys", "sysfs", 0, "")
		sys.Mount("dev", "dev", "devtmpfs", 0, "")
		sys.Mkdir("/dev/pts", 0600)
		sys.Mount("dev/pts", "dev/pts", "devtmpfs", 0, "")
	}

	logger.Println("yeah flint")

	if noninit {
		goto afterinit
	}

	go func() {
		logger.Println("managin power")
		for {
			switch <-sig {
			case syscall.SIGUSR1:
				sys.Reboot(sys.LINUX_REBOOT_CMD_POWER_OFF)
			case syscall.SIGINT:
				sys.Reboot(sys.LINUX_REBOOT_CMD_RESTART)
			}
		}
	}()

afterinit:

	{
		shell := exec.Command("/bin/zsh")
		shell.Stdin = os.Stdin
		shell.Stdout = os.Stdout
		shell.Stderr = os.Stderr
		shell.Env = os.Environ()
		shell.Start()
	}

	// kill zombies
	go handleZombie()

	// dumpster of sigchld
	go handleChld()

	for {
		time.Sleep(ttsec)
	}
}

func handleZombie() {
	for {
		syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		time.Sleep(onesec)
	}
}

func handleChld() {
	sigch := make(chan os.Signal, 3)
	signal.Notify(sigch, syscall.SIGCHLD)
	for {
		switch <-sigch {
		default:
			// trash
		}
	}
}
