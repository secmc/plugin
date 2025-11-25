package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/didntpot/pregdk"
	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/secmc/plugin/plugin/adapters/handlers"
	"github.com/secmc/plugin/plugin/adapters/plugin"
	pcfg "github.com/secmc/plugin/plugin/config"
	"github.com/secmc/plugin/plugin/ports"
)

func main() {
	logLevel := slog.LevelInfo
	if os.Getenv("DEBUG") != "" {
		logLevel = slog.LevelDebug
	}
	slog.SetLogLoggerLevel(logLevel)

	chat.Global.Subscribe(chat.StdoutSubscriber{})
	conf, err := readConfig(slog.Default())
	if err != nil {
		panic(err)
	}

	manager := plugin.NewManager(
		nil,
		slog.Default(),
		func(e ports.EventManager) player.Handler {
			return handlers.NewPlayerHandler(e)
		},
		func(e ports.EventManager) world.Handler {
			return handlers.NewWorldHandler(e)
		},
	)
	cfgPlugins, err := pcfg.LoadConfig("plugins/plugins.yaml")
	if err != nil {
		log.Fatalf("failed loading plugin config: %v", err)
	}

	if err := manager.StartWithConfig(cfgPlugins); err != nil {
		log.Fatalf("failed starting plugin manager: %v", err)
	}
	if ok := manager.WaitForPlugins(cfgPlugins.RequiredPlugins, time.Duration(cfgPlugins.HelloTimeoutMs)*time.Millisecond); !ok {
		if len(cfgPlugins.RequiredPlugins) > 0 {
			slog.Warn("required plugins did not load before timeout; custom items may not be included in resource pack")
		} else {
			slog.Warn("no plugin hello received before timeout; custom items registered later won't reach clients until restart")
		}
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()
	manager.SetServer(srv)
	manager.AttachWorld(srv.World())
	manager.AttachWorld(srv.Nether())
	manager.AttachWorld(srv.End())
	defer manager.Close()

	srv.Listen()
	for p := range srv.Accept() {
		manager.AttachPlayer(p)
	}
}

// readConfig reads the configuration from the config.toml file, or creates the
// file if it does not yet exist.
func readConfig(log *slog.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	var zero server.Config
	if _, err := os.Stat("config.toml"); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return zero, fmt.Errorf("stat config: %w", err)
		}
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %w", err)
		}
		if err := os.WriteFile("config.toml", data, 0o644); err != nil {
			return zero, fmt.Errorf("create default config: %w", err)
		}
	} else {
		data, err := os.ReadFile("config.toml")
		if err != nil {
			return zero, fmt.Errorf("read config: %w", err)
		}
		if err := toml.Unmarshal(data, &c); err != nil {
			return zero, fmt.Errorf("decode config: %w", err)
		}
	}
	cfg, _ := c.Config(log)
	listenerFunc(&cfg, c.Network.Address, []minecraft.Protocol{
		pregdk.Protocol(false),
		basicProtocol{Protocol: 860, Version: "1.21.124"},
	})
	return cfg, nil
}
