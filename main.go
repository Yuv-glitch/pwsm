package main

import (
	"flag"
	"log"
	"os"

	"golang.org/x/sys/windows/svc"
)

func main() {
	// detect if running as a Windows service
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Could not determine execution context: %v", err)
	}

	if isService {
		// running under SCM — parse --config flag and start service
		configName := flag.String("config", "", "service config name (without .yaml)")
		flag.Parse()

		if *configName == "" {
			log.Fatalf("--config flag is required when running as a service")
		}

		cfgPath := `C:\ProgramData\pwsm\services\` + *configName + ".yaml"
		cfg, err := readconfig(cfgPath)
		if err != nil {
			log.Fatalf("Failed to load config %q: %v", cfgPath, err)
		}

		if err := svc.Run(*configName, &service{config: cfg}); err != nil {
			log.Fatalf("Service %q failed: %v", *configName, err)
		}
		return
	}

	// running interactively — hand off to CLI
	if len(os.Args) < 2 {
		_ = rootCmd.Help()
		return
	}

	initCLI()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
