package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
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

	emitter := plugin.NewEmitter(
		srv,
		slog.Default(),
		func(e ports.EventEmitter) player.Handler {
			return handlers.NewPlayerHandler(e)
		},
		func(e ports.EventEmitter) world.Handler {
			return handlers.NewWorldHandler(e)
		},
	)
	if err := emitter.Start(""); err != nil {
		slog.Default().Error("start plugins", "error", err)
	}
	emitter.AttachWorld(srv.World())
	emitter.AttachWorld(srv.Nether())
	emitter.AttachWorld(srv.End())
	defer emitter.Close()

	srv.Listen()
	for p := range srv.Accept() {
		emitter.AttachPlayer(p)
	}
}

// readConfig reads the configuration from the config.toml file, or creates the
// file if it does not yet exist.
func readConfig(log *slog.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	var zero server.Config
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %v", err)
		}
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			return zero, fmt.Errorf("create default config: %v", err)
		}
		return c.Config(log)
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return zero, fmt.Errorf("read config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return zero, fmt.Errorf("decode config: %v", err)
	}
	return c.Config(log)
}
