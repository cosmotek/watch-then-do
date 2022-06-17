package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var params []string

var RootCmd = &cobra.Command{
	Use:  "wtd [pid] [action]",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("could not parse pid: %w", err)
		}

		action, ok := actions[args[1]]
		if !ok {
			return fmt.Errorf("failed to find action %s", args[1])
		}

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		exists := false
		ticker := time.NewTicker(time.Second * 5)

		for {
			running, err := ProcessRunning(int(pid))
			if err != nil {
				return err
			}

			if !exists && !running {
				return fmt.Errorf("process %d doesn't exist or has already stopped", pid)
			} else if exists && !running {
				fmt.Println("running action...")
				return action.Exec(params)
			} else if !exists && running {
				exists = true
			}

			select {
			case <-ticker.C:
			case <-sigs:
				return nil
			}
		}
	},
}

var actions = map[string]Action{
	"send_text": {
		ShellCommand: "osascript /Users/cosmotek/code/gomod/watch-then-do/sendMessage.applescript [0] '[Watch-Then-Do]: the process has completed'",
	},
	"echo": {
		ShellCommand: "echo [0]",
	},
}

func (a Action) Exec(params []string) error {
	cmdStr := a.ShellCommand
	for i, param := range params {
		cmdStr = strings.Replace(cmdStr, fmt.Sprintf("[%d]", i), param, -1)
	}

	args := strings.Split(cmdStr, " ")
	cmd := exec.Command(args[0], args[0:]...)

	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

type Action struct {
	ShellCommand string
}

func init() {
	RootCmd.Flags().StringArrayVarP(&params, "params", "p", []string{}, "set params for post-completion action")
}

func main() {
	err := RootCmd.Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func ProcessRunning(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, fmt.Errorf("failed to find process: %w", err)
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		if err.Error() == "os: process already finished" {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
