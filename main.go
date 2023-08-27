package main

import (
	"flag"
	"fmt"
	"github.com/Dreamacro/clash/constant/features"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/Dreamacro/clash/config"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/hub"
	"github.com/Dreamacro/clash/hub/executor"
	"github.com/Dreamacro/clash/log"

	"go.uber.org/automaxprocs/maxprocs"
)

var (
	flagset            map[string]bool
	version            bool
	initConfig         bool
	testConfig         bool
	geodataMode        bool
	homeDir            string
	configFile         string
	externalUI         string
	externalController string
	secret             string
)

func init() {
	flag.StringVar(&homeDir, "d", "", "set configuration directory")
	flag.BoolVar(&initConfig, "init-config", false, "initialize configuration directory and files")
	flag.StringVar(&configFile, "f", "", "specify configuration file")
	flag.StringVar(&externalUI, "ext-ui", "", "override external ui directory")
	flag.StringVar(&externalController, "ext-ctl", "", "override external controller address")
	flag.StringVar(&secret, "secret", "", "override secret for RESTful API")
	flag.BoolVar(&geodataMode, "m", false, "set geodata mode")
	flag.BoolVar(&version, "v", false, "show current version of clash")
	flag.BoolVar(&testConfig, "t", false, "test configuration and exit")
	flag.Parse()

	flagset = map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		flagset[f.Name] = true
	})
}

func main() {
	_, _ = maxprocs.Set(maxprocs.Logger(func(string, ...any) {}))
	if version {
		fmt.Printf("Clash Meta %s %s %s with %s %s\n",
			C.Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), C.BuildTime)
		if len(features.TAGS) != 0 {
			fmt.Printf("Use tags: %s\n", strings.Join(features.TAGS, ", "))
		}

		return
	}

	if homeDir != "" {
		if !filepath.IsAbs(homeDir) {
			currentDir, _ := os.Getwd()
			homeDir = filepath.Join(currentDir, homeDir)
		}
		C.SetHomeDir(homeDir)
	}

	if configFile != "" {
		if !filepath.IsAbs(configFile) {
			currentDir, _ := os.Getwd()
			configFile = filepath.Join(currentDir, configFile)
		}
		C.SetConfig(configFile)
	} else {
		configFile = filepath.Join(C.Path.HomeDir(), C.Path.Config())
		C.SetConfig(configFile)
	}

	if geodataMode {
		C.GeodataMode = true
	}

	if !initConfig {
		if _, err := os.Stat(C.Path.HomeDir()); os.IsNotExist(err) {
			fmt.Printf("Configuration directory does not exist:\n  %s\n",
						C.Path.HomeDir())
			fmt.Println("Setting with `-d` or use -init-config to initialize.")
			return
		}

		if _, err := os.Stat(C.Path.Config()); os.IsNotExist(err) {
			fmt.Printf("Configuration file does not exist:\n  %s\n",
						C.Path.Config())
			fmt.Println("Setting with `-f` or use -init-config to initialize.")
			return
		}
	}

	if err := config.Init(); err != nil {
		log.Fatalln("Initialize configuration directory error: %s", err.Error())
	}

	if testConfig {
		if _, err := executor.Parse(); err != nil {
			log.Errorln(err.Error())
			fmt.Printf("configuration file %s test failed\n", C.Path.Config())
			os.Exit(1)
		}
		fmt.Printf("configuration file %s test is successful\n", C.Path.Config())
		return
	}

	var options []hub.Option
	if flagset["ext-ui"] {
		options = append(options, hub.WithExternalUI(externalUI))
	}
	if flagset["ext-ctl"] {
		options = append(options, hub.WithExternalController(externalController))
	}
	if flagset["secret"] {
		options = append(options, hub.WithSecret(secret))
	}

	if err := hub.Parse(options...); err != nil {
		log.Fatalln("Parse config error: %s", err.Error())
	}

	defer executor.Shutdown()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
