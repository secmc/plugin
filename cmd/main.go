package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/didntpot/pregdk"
	"github.com/pelletier/go-toml"
	"github.com/secmc/plugin/plugin/adapters/handlers"
	"github.com/secmc/plugin/plugin/adapters/plugin"
	"github.com/secmc/plugin/plugin/ports"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	chat.Global.Subscribe(chat.StdoutSubscriber{})
	conf, err := readConfig(slog.Default())
	if err != nil {
		panic(err)
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()

	manager := plugin.NewManager(
		srv,
		slog.Default(),
		func(e ports.EventManager) player.Handler {
			return handlers.NewPlayerHandler(e)
		},
		func(e ports.EventManager) world.Handler {
			return handlers.NewWorldHandler(e)
		},
	)
	if err := manager.Start("configs/plugins.yaml"); err != nil {
		slog.Default().Error("start plugins", "error", err)
	}
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
	if _, err := os.Stat("configs/config.toml"); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return zero, fmt.Errorf("stat config: %w", err)
		}
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %w", err)
		}
		if err := os.WriteFile("configs/config.toml", data, 0o644); err != nil {
			return zero, fmt.Errorf("create default config: %w", err)
		}
	} else {
		data, err := os.ReadFile("configs/config.toml")
		if err != nil {
			return zero, fmt.Errorf("read config: %w", err)
		}
		if err := toml.Unmarshal(data, &c); err != nil {
			return zero, fmt.Errorf("decode config: %w", err)
		}
	}
	cfg, _ := c.Config(log)
	listenerFunc(&cfg, c.Network.Address, single(pregdk.Protocol(false)))
	return cfg, nil
}
