package plugin

import (
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	pb "github.com/secmc/plugin/proto/generated"
)

func (m *Manager) registerCommands(p *pluginProcess, specs []*pb.CommandSpec) {
	for _, spec := range specs {
		if spec == nil || spec.Name == "" {
			continue
		}
		name := strings.TrimPrefix(spec.Name, "/")

		aliases := make([]string, 0, len(spec.Aliases))
		for _, alias := range spec.Aliases {
			alias = strings.TrimPrefix(alias, "/")
			if alias == "" || alias == name {
				continue
			}
			aliases = append(aliases, alias)
		}

		binding := commandBinding{pluginID: p.id, command: name, descriptor: spec}
		m.mu.Lock()
		m.commands[name] = binding
		for _, alias := range aliases {
			m.commands[alias] = binding
		}
		m.mu.Unlock()

		pc := pluginCommand{
			mgr:      m,
			pluginID: p.id,
			name:     name,
			params:   buildParamInfo(spec),
		}
		cmd.Register(cmd.New(name, spec.Description, aliases, pc))
	}
}

type pluginCommand struct {
	mgr      *Manager
	pluginID string
	name     string
	params   []cmd.ParamInfo
	// Args swallows all provided arguments so Dragonfly's parser doesn't reject the command
	// for having leftover args. Actual parsing/dispatch is handled elsewhere via events.
	Args cmd.Varargs
}

func (c pluginCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	_, ok := src.(*player.Player)
	if !ok {
		output.Errorf("command only available to players")
		return
	}
	// No-op: PlayerHandler.HandleCommandExecution emits command events
}

// DescribeParams exposes parameter info to Dragonfly so the client can render usage and enums.
func (c pluginCommand) DescribeParams(src cmd.Source) []cmd.ParamInfo {
	return c.params
}

type staticEnum struct {
	typeName string
	options  []string
}

func (e staticEnum) Type() string                { return e.typeName }
func (e staticEnum) Options(cmd.Source) []string { return e.options }

func buildParamInfo(spec *pb.CommandSpec) []cmd.ParamInfo {
	if spec == nil || len(spec.Params) == 0 {
		return nil
	}
	params := make([]cmd.ParamInfo, 0, len(spec.Params))
	for _, p := range spec.Params {
		if p == nil {
			continue
		}
		var value any
		switch p.Type {
		case pb.ParamType_PARAM_INT:
			value = int(0)
		case pb.ParamType_PARAM_FLOAT:
			value = float64(0)
		case pb.ParamType_PARAM_BOOL:
			value = false
		case pb.ParamType_PARAM_VARARGS:
			value = cmd.Varargs("")
		case pb.ParamType_PARAM_ENUM:
			// Prefer explicit enum values provided by the plugin.
			if len(p.EnumValues) > 0 {
				value = staticEnum{typeName: "Enum:" + p.Name, options: p.EnumValues}
			} else {
				// No values provided; fall back to plain string.
				value = ""
			}
		default: // PARAM_STRING and fallback
			// If enum values provided for a string param, treat as enum.
			if len(p.EnumValues) > 0 {
				value = staticEnum{typeName: "Enum:" + p.Name, options: p.EnumValues}
			} else {
				value = ""
			}
		}
		params = append(params, cmd.ParamInfo{
			Name:     p.Name,
			Value:    value,
			Optional: p.Optional,
			Suffix:   p.Suffix,
		})
	}
	return params
}
