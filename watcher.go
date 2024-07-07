package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

const (
	defaultTarget  = "./"
	defaultPattern = "**"
	waitForTerm    = 5 * time.Second
)

var (
	targets   []string
	patterns  []string
	ignores   []string
	delay     time.Duration
	restart   bool
	sigopt    string
	filteropt []string
	verbose   bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "watcher [COMMAND]",
		Short: "A file watching and command restarting tool",
		Long:  `Run the COMMAND and restart when a file matching the pattern has been modified.`,
		Args:  cobra.MinimumNArgs(1),
		Run:   execute,
	}

	rootCmd.PersistentFlags().StringArrayVarP(&targets, "target", "t", []string{defaultTarget}, "observation target `path` (default \"./\")")
	rootCmd.PersistentFlags().StringArrayVarP(&patterns, "pattern", "p", []string{defaultPattern}, "trigger pathname `glob` pattern (default \"**\")")
	rootCmd.PersistentFlags().StringArrayVarP(&ignores, "ignore", "i", nil, "ignore pathname `glob` pattern")
	rootCmd.PersistentFlags().DurationVarP(&delay, "delay", "d", time.Second, "`duration` to delay the restart of the command")
	rootCmd.PersistentFlags().BoolVarP(&restart, "restart", "r", false, "restart the command on exit")
	rootCmd.PersistentFlags().StringVarP(&sigopt, "signal", "s", "", "`signal` used to stop the command (default \"SIGTERM\")")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringArrayVarP(&filteropt, "filter", "f", nil, "filter file system `event` (CREATE|WRITE|REMOVE|RENAME|CHMOD)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execute(cmd *cobra.Command, args []string) {
	parsedSignal, err := parseSignal(sigopt)
	if err != nil {
		log.Fatalf("[WATCHER] %v", err)
	}

	filters, err := parseFilters(filteropt)
	if err != nil {
		log.Fatalf("[WATCHER] %v", err)
	}

	watcherLog("verbose: %v", verbose)
	watcherLog("targets:  %q", targets)
	watcherLog("patterns: %q", patterns)
	watcherLog("ignores:  %q", ignores)
	watcherLog("filter:   %v", filters)
	watcherLog("delay:    %v", delay)
	watcherLog("signal:   %s", parsedSignal)
	watcherLog("restart:  %v", restart)

	modC, errC, err := watch(targets, patterns, ignores, filters)
	if err != nil {
		log.Fatalf("[WATCHER] error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	reload := runner(ctx, &wg, args, delay, parsedSignal.(syscall.Signal), restart)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case name, ok := <-modC:
				if !ok {
					cancel()
					wg.Wait()
					log.Fatalf("[WATCHER] wacher closed")
					return
				}
				reload <- name
			case err := <-errC:
				cancel()
				wg.Wait()
				log.Fatalf("[WATCHER] wacher error: %v", err)
				return
			}
		}
	}()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	parsedSignal = <-s
	log.Printf("[WATCHER] signal: %v", parsedSignal)
	cancel()
	wg.Wait()
}

func watch(targets, patterns, ignores []string, filters fsnotify.Op) (<-chan string, <-chan error, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}

	if err := addTargets(w, targets, patterns, ignores); err != nil {
		return nil, nil, err
	}

	modC := make(chan string)
	errC := make(chan error)
	watchOp := ^filters

	go func() {
		defer close(modC)
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					errC <- fmt.Errorf("watcher.Events closed")
					return
				}

				name := filepath.ToSlash(event.Name)
				watcherLog("event: %v %q", event.Op, name)

				if ignore, err := matchPatterns(name, ignores); err != nil {
					errC <- fmt.Errorf("match ignores: %w", err)
					return
				} else if ignore {
					continue
				}

				if event.Has(watchOp) {
					if match, err := matchPatterns(name, patterns); err != nil {
						errC <- fmt.Errorf("match patterns: %w", err)
						return
					} else if match {
						modC <- name
					}
				}

				// add watcher if new directory.
				if event.Has(fsnotify.Create) {
					fi, err := os.Stat(name)
					if err != nil {
						// ignore stat errors (notfound, permission, etc.)
						log.Printf("[WATCHER] watcher: %v", err)
					} else if fi.IsDir() {
						err := addDirRecursive(w, name, patterns, ignores, modC)
						if err != nil {
							errC <- err
							return
						}
					}
				}

			case err, ok := <-w.Errors:
				errC <- fmt.Errorf("watcher.Errors (%v): %w", ok, err)
				return
			}
		}
	}()

	return modC, errC, nil
}

func runner(ctx context.Context, wg *sync.WaitGroup, cmd []string, delay time.Duration, sig syscall.Signal, autorestart bool) chan<- string {
	reload := make(chan string)
	trigger := make(chan string)

	go func() {
		for name := range reload {
			// ignore restart when the trigger is not waiting
			select {
			case trigger <- name:
			default:
			}
		}
	}()

	var pcmd string // command string for display.
	for _, s := range cmd {
		if strings.ContainsAny(s, " \t\"'") {
			s = fmt.Sprintf("%q", s)
		}
		pcmd += " " + s
	}
	pcmd = pcmd[1:]

	stdinC := make(chan bytesErr, 1)
	go func() {
		b1 := make([]byte, 255)
		b2 := make([]byte, 255)
		for {
			n, err := os.Stdin.Read(b1)
			stdinC <- bytesErr{b1[:n], err}
			b1, b2 = b2, b1
		}
	}()

	chldDone := makeChildDoneChan()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			cmdctx, cancel := context.WithCancel(ctx)
			restart := make(chan struct{})
			done := make(chan struct{})

			go func() {
				log.Printf("[WATCHER] start: %s", pcmd)
				clearChBuf(chldDone)
				stdin := &stdinReader{stdinC, chldDone}
				err := runCmd(cmdctx, cmd, sig, stdin)
				if err != nil {
					log.Printf("[WATCHER] command error: %v", err)
				} else {
					log.Printf("[WATCHER] command exit status 0")
				}
				if autorestart {
					close(restart)
				}

				close(done)
			}()

			select {
			case <-ctx.Done():
				cancel()
				<-done
				return
			case name := <-trigger:
				log.Printf("[WATCHER] triggered: %q", name)
			case <-restart:
				watcherLog("auto restart")
			}

			watcherLog("wait %v", delay)
			select {
			case <-ctx.Done():
				cancel()
				<-done
				return
			case <-time.After(delay):
			}
			cancel()
			<-done // wait process closed
		}
	}()

	return reload
}

func runCmd(ctx context.Context, cmd []string, sig syscall.Signal, stdin *stdinReader) error {
	c := prepareCommand(cmd)
	c.Stdin = stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Start(); err != nil {
		return err
	}

	var cerr error
	done := make(chan struct{})
	go func() {
		cerr = waitCmd(c)
		close(done)
	}()

	select {
	case <-done:
		if cerr != nil {
			cerr = fmt.Errorf("process exit: %w", cerr)
		}
		return cerr
	case <-ctx.Done():
		if err := killChilds(c, sig); err != nil {
			return fmt.Errorf("kill childs: %w", err)
		}
	}

	select {
	case <-done:
	case <-time.After(waitForTerm):
		if err := killChilds(c, syscall.SIGKILL); err != nil {
			return fmt.Errorf("kill childs (SIGKILL): %w", err)
		}
		<-done
	}

	if cerr != nil {
		return fmt.Errorf("process canceled: %w", cerr)
	}
	return nil
}

func addTargets(w *fsnotify.Watcher, targets, patterns, ignores []string) error {
	for _, t := range targets {
		t = path.Clean(t)
		fi, err := os.Stat(t)
		if err != nil {
			return fmt.Errorf("stat: %w", err)
		}
		if fi.IsDir() {
			if err := addDirRecursive(w, t, patterns, ignores, nil); err != nil {
				return err
			}
		}
		watcherLog("watching target: %q", t)
		if err := w.Add(t); err != nil {
			return err
		}
	}
	return nil
}

func addDirRecursive(w *fsnotify.Watcher, t string, patterns, ignores []string, ch chan<- string) error {
	watcherLog("watching target: %q", t)
	err := w.Add(t)
	if err != nil {
		return fmt.Errorf("wacher add: %w", err)
	}
	des, err := os.ReadDir(t)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}
	for _, de := range des {
		name := path.Join(t, de.Name())
		if ignore, err := matchPatterns(name, ignores); err != nil {
			return fmt.Errorf("match ignores: %w", err)
		} else if ignore {
			continue
		}
		if ch != nil {
			if match, err := matchPatterns(name, patterns); err != nil {
				return fmt.Errorf("match patterns: %w", err)
			} else if match {
				ch <- name
			}
		}
		if de.IsDir() {
			err = addDirRecursive(w, name, patterns, ignores, ch)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func watcherLog(str string, args ...interface{}) {
	if verbose {
		log.Printf("[WATCHER] "+str, args...)
	}
}
