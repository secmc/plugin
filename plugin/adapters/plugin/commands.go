package plugin

import (
	"slices"
	"sort"
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	pb "github.com/secmc/plugin/proto/generated/go"
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
		case pb.ParamType_PARAM_TARGET, pb.ParamType_PARAM_TARGETS:
			value = []cmd.Target{}
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

// resolveTargetArg resolves a single selector or name to a player UUID string.
// Only player targets are supported in this initial implementation.
// TODO: support entity targets.
func (m *Manager) resolveTargetArg(src *player.Player, arg string) (string, bool) {
	uuids := m.resolveSelectorUUIDs(src, arg)
	if len(uuids) == 0 {
		return "", false
	}
	return uuids[0], true
}

// resolveTargetsArg resolves a selector or name to a list of player UUIDs.
func (m *Manager) resolveTargetsArg(src *player.Player, arg string) []string {
	return m.resolveSelectorUUIDs(src, arg)
}

// resolveSelectorUUIDs normalizes a selector/name and returns UUIDs per selector semantics:
// - direct UUID returns that UUID
// - @s -> self
// - @a -> all players (self first)
// - @p -> nearest single (returned as single-item list)
// - @r -> deterministic single (sorted by name, first)
// - @e -> approximated as all players (self first)
// - name -> matching player (single)
func (m *Manager) resolveSelectorUUIDs(src *player.Player, arg string) []string {
	if src == nil {
		return nil
	}
	a := strings.ToLower(strings.TrimSpace(arg))
	if a == "" {
		return nil
	}
	// Direct UUID pass-through if already resolved.
	if len(a) == 36 && strings.Count(a, "-") == 4 {
		return []string{a}
	}
	// Gather players in the same world as the source.
	tx := src.Tx()
	var players []*player.Player
	for pl := range tx.Players() {
		if pp, ok := pl.(*player.Player); ok {
			players = append(players, pp)
		}
	}
	if len(players) == 0 {
		return nil
	}
	switch a {
	case "@s":
		return []string{src.UUID().String()}
	case "@a":
		out := make([]string, 0, len(players))
		// Ensure self is first for downstream single-target callers.
		out = append(out, src.UUID().String())
		for _, pl := range players {
			if pl == nil || pl.UUID() == src.UUID() {
				continue
			}
			out = append(out, pl.UUID().String())
		}
		return out
	case "@p":
		// Nearest player to the source.
		pos := src.Position()
		type distPl struct {
			d float64
			p *player.Player
		}
		ds := make([]distPl, 0, len(players))
		for _, pl := range players {
			ds = append(ds, distPl{d: pl.Position().Sub(pos).Len(), p: pl})
		}
		sort.Slice(ds, func(i, j int) bool { return ds[i].d < ds[j].d })
		return []string{ds[0].p.UUID().String()}
	case "@r":
		// Random player: deterministic fallback — sort by name and take first.
		slices.SortFunc(players, func(a, b *player.Player) int {
			return strings.Compare(strings.ToLower(a.Name()), strings.ToLower(b.Name()))
		})
		return []string{players[0].UUID().String()}
	case "@e":
		// Entities not supported; approximate with all players, self first.
		out := make([]string, 0, len(players))
		out = append(out, src.UUID().String())
		for _, pl := range players {
			if pl == nil || pl.UUID() == src.UUID() {
				continue
			}
			out = append(out, pl.UUID().String())
		}
		return out
	default:
		// Try by exact or case-insensitive name.
		for _, pl := range players {
			if strings.EqualFold(pl.Name(), arg) {
				return []string{pl.UUID().String()}
			}
		}
		return nil
	}
}
