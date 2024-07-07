package watcher

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

var signalMap = map[string]os.Signal{
	"1":         syscall.SIGHUP,
	"HUP":       syscall.SIGHUP,
	"SIGHUP":    syscall.SIGHUP,
	"SIG_HUP":   syscall.SIGHUP,
	"2":         syscall.SIGINT,
	"INT":       syscall.SIGINT,
	"SIGINT":    syscall.SIGINT,
	"SIG_INT":   syscall.SIGINT,
	"3":         syscall.SIGQUIT,
	"QUIT":      syscall.SIGQUIT,
	"SIGQUIT":   syscall.SIGQUIT,
	"SIG_QUIT":  syscall.SIGQUIT,
	"9":         syscall.SIGKILL,
	"KILL":      syscall.SIGKILL,
	"SIGKILL":   syscall.SIGKILL,
	"SIG_KILL":  syscall.SIGKILL,
	"10":        syscall.SIGUSR1,
	"USR1":      syscall.SIGUSR1,
	"SIGUSR1":   syscall.SIGUSR1,
	"SIG_USR1":  syscall.SIGUSR1,
	"12":        syscall.SIGUSR2,
	"USR2":      syscall.SIGUSR2,
	"SIGUSR2":   syscall.SIGUSR2,
	"SIG_USR2":  syscall.SIGUSR2,
	"15":        syscall.SIGTERM,
	"TERM":      syscall.SIGTERM,
	"SIGTERM":   syscall.SIGTERM,
	"SIG_TERM":  syscall.SIGTERM,
	"28":        syscall.SIGWINCH,
	"WINCH":     syscall.SIGWINCH,
	"SIGWINCH":  syscall.SIGWINCH,
	"SIG_WINCH": syscall.SIGWINCH,
	"":          syscall.SIGTERM, // Default case
}

func parseSignal(signal string) (os.Signal, error) {
	sig, exists := signalMap[strings.ToUpper(signal)]
	if !exists {
		return nil, fmt.Errorf("unsupported signal: %s", signal)
	}
	return sig, nil
}
