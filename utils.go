package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// makeChildDoneChan returns a chan that notifies the child process has exited.
//
// On UNIX like OS, it is notified by SIGCHLD.
func makeChildDoneChan() <-chan struct{} {
	c := make(chan struct{}, 1)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGCHLD)
		for {
			<-sig
			select {
			case c <- struct{}{}:
			default:
			}
		}
	}()
	return c
}

func clearChBuf[T any](c <-chan T) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}

func prepareCommand(cmd []string) *exec.Cmd {
	c := exec.Command(cmd[0], cmd[1:]...)
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return c
}

func waitCmd(cmd *exec.Cmd) error {
	return cmd.Wait()
}

func killChilds(c *exec.Cmd, sig syscall.Signal) error {
	err := syscall.Kill(-c.Process.Pid, sig)
	if err == nil && sig != syscall.SIGKILL && sig != syscall.SIGCONT {
		// prosess can be stopped, so it must be start by SIGCONT.
		err = syscall.Kill(-c.Process.Pid, syscall.SIGCONT)
	}
	return err
}
