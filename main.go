package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/xackery/launcheq/client"
	"github.com/xackery/launcheq/config"
	"github.com/xackery/launcheq/gui"
	"github.com/xackery/wlk/walk"
)

//go:embed splash.png
var starteqSplash []byte

var (
	Version    string
	PatcherUrl string
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exeName, err := os.Executable()
	if err != nil {
		gui.MessageBox("Error", "Failed to get executable name", true)
		os.Exit(1)
	}
	baseName := filepath.Base(exeName)
	if strings.Contains(baseName, ".") {
		baseName = baseName[0:strings.Index(baseName, ".")]
	}
	cfg, err := config.New(context.Background(), baseName)
	if err != nil {
		gui.MessageBox("Error", "Failed to load config: "+err.Error(), true)
		os.Exit(1)
	}

	err = gui.NewMainWindow(ctx, cancel, cfg, starteqSplash)
	if err != nil {
		gui.MessageBox("Error", "Failed to create main window: "+err.Error(), true)
		os.Exit(1)
	}
	PatcherUrl = strings.TrimSuffix(PatcherUrl, "/")
	if Version == "" {
		Version = "dev"
	}

	c, err := client.New(ctx, cancel, cfg, Version, PatcherUrl)
	if err != nil {
		gui.MessageBox("Error", "Failed to create client: "+err.Error(), true)
		os.Exit(1)
	}
	defer c.DumpLog()

	gui.SubscribeClose(func(canceled *bool, reason walk.CloseReason) {
		if ctx.Err() != nil {
			fmt.Println("Accepting exit")
			return
		}
		*canceled = true
		fmt.Println("Got close message")
		gui.SetTitle("Closing...")
		cancel()
	})

	go func() {
		<-ctx.Done()
		fmt.Println("Doing clean up process...")
		gui.Close()
		walk.App().Exit(0)
		fmt.Println("Done, exiting")
		c.DumpLog()
		os.Exit(0)
	}()

	c.AutoPlay()
	errCode := gui.Run()
	if errCode != 0 {
		fmt.Println("Failed to run:", errCode)
		os.Exit(1)
	}

}
