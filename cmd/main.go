package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"rclone-sync-toggle/internal/config"
	"rclone-sync-toggle/internal/daemon"
	"rclone-sync-toggle/internal/i18n"
	"rclone-sync-toggle/internal/ui"
)

func main() {
	togglePath := flag.String("toggle", "", "Toggle a file or folder between local and virtual")
	configMode := flag.Bool("config", false, "Open the Folder -> Remote configuration window")
	language := flag.String("lang", "", "Interface language: es or en (default: system language)")
	flag.Parse()

	if err := i18n.Init(*language); err != nil {
		log.Fatal(err)
	}

	if *configMode {
		ui.OpenConfigWindow()
		return
	}

	profiles, err := config.LoadProfiles()
	if err != nil {
		log.Fatalf(i18n.T("config.read_error"), err)
	}
	if len(profiles) == 0 {
		log.Fatalf(i18n.T("config.empty"), config.ProfilesPath())
	}

	var activeProfiles []*config.ConnectorProfile
	for _, profile := range profiles {
		if profile.Enabled {
			activeProfiles = append(activeProfiles, profile)
		}
	}

	if *togglePath != "" {
		ui.ToggleItem(*togglePath, activeProfiles)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	for _, p := range activeProfiles {
		wg.Add(1)
		go func(prof *config.ConnectorProfile) {
			defer wg.Done()
			daemon.RunWorker(ctx, prof)
		}(p)
	}

	log.Printf(i18n.T("daemon.started"), len(activeProfiles))
	<-ctx.Done()
	log.Println(i18n.T("daemon.stopping"))
	wg.Wait()
}
