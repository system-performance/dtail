package main

import (
	"flag"
	"os"
	"time"

	"github.com/mimecast/dtail/internal/color"
	"github.com/mimecast/dtail/internal/config"
	"github.com/mimecast/dtail/internal/logger"
	"github.com/mimecast/dtail/internal/pprof"
	"github.com/mimecast/dtail/internal/server"
	"github.com/mimecast/dtail/internal/user"
	"github.com/mimecast/dtail/internal/version"
)

// The evil begins here.
func main() {
	var cfgFile string
	var debugEnable bool
	var displayVersion bool
	var noColor bool
	var pprofEnable bool
	var shutdownAfter int
	var sshPort int

	userName := user.Name()

	flag.BoolVar(&debugEnable, "debug", false, "Activate debug messages")
	flag.BoolVar(&displayVersion, "version", false, "Display version")
	flag.BoolVar(&noColor, "noColor", false, "Disable ANSII terminal colors")
	flag.BoolVar(&pprofEnable, "pprofEnable", false, "Enable pprof server")
	flag.IntVar(&shutdownAfter, "shutdownAfter", 0, "Automatically shutdown after so many seconds")
	flag.IntVar(&sshPort, "port", 2222, "SSH server port")
	flag.StringVar(&cfgFile, "cfg", "", "Config file path")

	flag.Parse()

	config.Read(cfgFile, sshPort)
	color.Colored = !noColor

	if displayVersion {
		version.PrintAndExit()
	}

	serverEnable := true
	silentEnable := false
	nothingEnable := false
	logger.Start(serverEnable, debugEnable, silentEnable, nothingEnable)
	defer logger.Stop()

	if shutdownAfter > 0 {
		go func() {
			defer os.Exit(1)

			logger.Info("Enabling auto shutdown timer", shutdownAfter)
			time.Sleep(time.Duration(shutdownAfter) * time.Second)
			logger.Info("Auto shutdown timer reached, shutting down now")
		}()
	}

	if pprofEnable || config.Common.PProfEnable {
		pprof.Start()
	}

	logger.Info("Launching server", version.String(), userName)
	sshServer := server.New()
	sshServer.Start()
}
