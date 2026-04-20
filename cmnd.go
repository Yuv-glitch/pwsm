package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const configDir = `C:\ProgramData\pwsm\services`

func configPath(name string) string {
	return filepath.Join(configDir, name+".yaml")
}

func exePath() (string, error) {
	return os.Executable()
}

var rootCmd = &cobra.Command{
	Use:   "pwsm",
	Short: "Persistent Windows Service Manager",
	Long:  "pwsm - manage any script or executable as a Windows service.",
}

var initcmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise the working directory and config folder",
	Run: func(cmd *cobra.Command, args []string) {
		dirs := []string{
			`C:\ProgramData\pwsm\services`,
			`C:\ProgramData\pwsm\logs`,
		}
		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("Failed to create directory %s: %v", dir, err)
			}
		}
		fmt.Println("Initialization complete. Configs should be placed in C:\\ProgramData\\pwsm\\services with .yaml extension.")
	},
}

var installCmd = &cobra.Command{
	Use:   "install <servicename>",
	Short: "Register a service from its config file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		cfgPath := configPath(name)

		// validate config exists and is readable
		cfg, err := readconfig(cfgPath)
		if err != nil {
			log.Fatalf("Could not read config for %q: %v", name, err)
		}

		exe, err := exePath()
		if err != nil {
			log.Fatalf("Could not determine exe path: %v", err)
		}

		m, err := mgr.Connect()
		if err != nil {
			log.Fatalf("Could not connect to SCM: %v", err)
		}
		defer m.Disconnect()

		scCmd := exec.Command("sc.exe", "create", name, "binPath=", exe+" --config "+name, "start=", "auto", "DisplayName=", cfg.Name)
		out, err := scCmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Could not create service: %v\n%s", err, out)
		}

		fmt.Printf("Service %q installed successfully.\n", name)
		fmt.Printf("Config: %s\n", cfgPath)
		fmt.Printf("Run: pwsm start %s\n", name)
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <servicename>",
	Short: "Remove a registered service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		m, err := mgr.Connect()
		if err != nil {
			log.Fatalf("Could not connect to SCM: %v", err)
		}
		defer m.Disconnect()

		s, err := m.OpenService(name)
		if err != nil {
			log.Fatalf("Service %q not found: %v", name, err)
		}
		defer s.Close()

		err = s.Delete()
		if err != nil {
			log.Fatalf("Could not delete service %q: %v", name, err)
		}

		fmt.Printf("Service %q uninstalled successfully.\n", name)
		cfgPath := configPath(name)
		if err := os.Remove(cfgPath); err != nil {
			log.Printf("Warning: could not remove config file %s: %v", cfgPath, err)
		}
		fmt.Printf("Config file %s removed.\n", cfgPath)
	},
}

var startCmd = &cobra.Command{
	Use:   "start <servicename>",
	Short: "Start a registered service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		m, err := mgr.Connect()
		if err != nil {
			log.Fatalf("Could not connect to SCM: %v", err)
		}
		defer m.Disconnect()

		s, err := m.OpenService(name)
		if err != nil {
			log.Fatalf("Service %q not found: %v", name, err)
		}
		defer s.Close()

		err = s.Start()
		if err != nil {
			log.Fatalf("Could not start service %q: %v", name, err)
		}

		fmt.Printf("Service %q started.\n", name)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop <servicename>",
	Short: "Stop a running service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		m, err := mgr.Connect()
		if err != nil {
			log.Fatalf("Could not connect to SCM: %v", err)
		}
		defer m.Disconnect()

		s, err := m.OpenService(name)
		if err != nil {
			log.Fatalf("Service %q not found: %v", name, err)
		}
		defer s.Close()

		status, err := s.Control(svc.Stop)
		if err != nil {
			log.Fatalf("Could not stop service %q: %v", name, err)
		}

		// wait up to 10s for the service to stop
		timeout := time.Now().Add(10 * time.Second)
		for status.State != svc.Stopped {
			if time.Now().After(timeout) {
				log.Fatalf("Timed out waiting for service %q to stop.", name)
			}
			time.Sleep(300 * time.Millisecond)
			status, err = s.Query()
			if err != nil {
				log.Fatalf("Could not query service %q: %v", name, err)
			}
		}

		fmt.Printf("Service %q stopped.\n", name)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status <servicename>",
	Short: "Show the current status of a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		m, err := mgr.Connect()
		if err != nil {
			log.Fatalf("Could not connect to SCM: %v", err)
		}
		defer m.Disconnect()

		s, err := m.OpenService(name)
		if err != nil {
			log.Fatalf("Service %q not found: %v", name, err)
		}
		defer s.Close()

		status, err := s.Query()
		if err != nil {
			log.Fatalf("Could not query service %q: %v", name, err)
		}

		stateStr := map[svc.State]string{
			svc.Stopped:         "Stopped",
			svc.StartPending:    "Start Pending",
			svc.StopPending:     "Stop Pending",
			svc.Running:         "Running",
			svc.ContinuePending: "Continue Pending",
			svc.PausePending:    "Pause Pending",
			svc.Paused:          "Paused",
		}

		state, ok := stateStr[status.State]
		if !ok {
			state = "Unknown"
		}

		fmt.Printf("Service : %s\n", name)
		fmt.Printf("State   : %s\n", state)
		fmt.Printf("PID     : %d\n", status.ProcessId)
	},
}

func initCLI() {
	rootCmd.AddCommand(initcmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
}
