package plugin

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	pb "github.com/secmc/plugin/proto/generated/go"
)

type cancelContext interface {
	Cancel()
}

func playerWorldDimension(p *player.Player) string {
	if p == nil {
		return ""
	}
	tx := p.Tx()
	if tx == nil {
		return ""
	}
	w := tx.World()
	if w == nil {
		return ""
	}
	return strings.ToLower(fmt.Sprint(w.Dimension()))
}

func playerWorldRef(p *player.Player) *pb.WorldRef {
	if p == nil {
		return nil
	}
	tx := p.Tx()
	if tx == nil {
		return nil
	}
	return protoWorldRef(tx.World())
}

func worldDimension(w *world.World) string {
	if w == nil {
		return ""
	}
	return strings.ToLower(fmt.Sprint(w.Dimension()))
}

func protoVec3(v mgl64.Vec3) *pb.Vec3 {
	return &pb.Vec3{X: v[0], Y: v[1], Z: v[2]}
}

func protoRotation(r cube.Rotation) *pb.Rotation {
	yaw, pitch := r.Elem()
	return &pb.Rotation{Yaw: float32(yaw), Pitch: float32(pitch)}
}

func protoBBox(box cube.BBox) *pb.BBox {
	return &pb.BBox{Min: protoVec3(box.Min()), Max: protoVec3(box.Max())}
}

func bboxFromProto(box *pb.BBox) (cube.BBox, bool) {
	if box == nil || box.Min == nil || box.Max == nil {
		return cube.BBox{}, false
	}
	return cube.Box(box.Min.X, box.Min.Y, box.Min.Z, box.Max.X, box.Max.Y, box.Max.Z), true
}

func protoBlockPos(pos cube.Pos) *pb.BlockPos {
	return &pb.BlockPos{X: int32(pos.X()), Y: int32(pos.Y()), Z: int32(pos.Z())}
}

func protoBlockState(b world.Block) *pb.BlockState {
	if b == nil {
		return nil
	}
	name, props := b.EncodeBlock()
	out := &pb.BlockState{Name: name}
	if len(props) > 0 {
		out.Properties = make(map[string]string, len(props))
		for k, v := range props {
			out.Properties[k] = fmt.Sprint(v)
		}
	}
	return out
}

func protoLiquidState(l world.Liquid) *pb.LiquidState {
	if l == nil {
		return nil
	}
	return &pb.LiquidState{
		Block:      protoBlockState(l),
		Depth:      int32(l.LiquidDepth()),
		Falling:    l.LiquidFalling(),
		LiquidType: l.LiquidType(),
	}
}

func protoLiquidOrBlockState(b world.Block) *pb.LiquidState {
	if b == nil {
		return nil
	}
	if l, ok := b.(world.Liquid); ok {
		return protoLiquidState(l)
	}
	return &pb.LiquidState{Block: protoBlockState(b)}
}

func protoItemStack(it item.Stack) *pb.ItemStack {
	if it.Empty() {
		return nil
	}
	itm := it.Item()
	if itm == nil {
		return nil
	}
	name, meta := itm.EncodeItem()
	return &pb.ItemStack{
		Name:  name,
		Meta:  int32(meta),
		Count: int32(it.Count()),
	}
}

func protoItemStackPtr(it *item.Stack) *pb.ItemStack {
	if it == nil {
		return nil
	}
	return protoItemStack(*it)
}

func convertProtoItemStackValue(stack *pb.ItemStack) (item.Stack, bool) {
	if stack == nil || stack.Name == "" {
		return item.Stack{}, false
	}
	material, ok := world.ItemByName(stack.Name, int16(stack.Meta))
	if !ok {
		return item.Stack{}, false
	}
	count := int(stack.Count)
	if count <= 0 {
		return item.Stack{}, false
	}
	return item.NewStack(material, count), true
}

func blockFromProto(state *pb.BlockState) (world.Block, bool) {
	if state == nil || state.Name == "" {
		return nil, false
	}
	properties := make(map[string]any, len(state.Properties))
	for k, v := range state.Properties {
		properties[k] = parsePropertyValue(v)
	}
	return world.BlockByName(state.Name, properties)
}

func parsePropertyValue(v string) any {
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return int(i)
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	return v
}

func convertProtoBlockPositionsToCube(blocks []*pb.BlockPos) []cube.Pos {
	if len(blocks) == 0 {
		return nil
	}
	converted := make([]cube.Pos, 0, len(blocks))
	for _, blk := range blocks {
		if blk == nil {
			continue
		}
		converted = append(converted, cube.Pos{int(blk.X), int(blk.Y), int(blk.Z)})
	}
	if len(converted) == 0 {
		return nil
	}
	return converted
}

func vec3FromProto(vec *pb.Vec3) (mgl64.Vec3, bool) {
	if vec == nil {
		return mgl64.Vec3{}, false
	}
	return mgl64.Vec3{float64(vec.X), float64(vec.Y), float64(vec.Z)}, true
}

func parseProtoAddress(addr *pb.Address) *net.UDPAddr {
	if addr == nil {
		return nil
	}
	parsed := net.ParseIP(addr.Host)
	return &net.UDPAddr{IP: parsed, Port: int(addr.Port)}
}

func protoWorldRef(w *world.World) *pb.WorldRef {
	if w == nil {
		return nil
	}
	ref := &pb.WorldRef{
		Name:      w.Name(),
		Dimension: worldDimension(w),
	}
	return ref
}

func protoDamageSource(src world.DamageSource) *pb.DamageSource {
	if src == nil {
		return nil
	}
	desc := fmt.Sprint(src)
	ds := &pb.DamageSource{Type: fmt.Sprintf("%T", src)}
	if desc != "" {
		ds.Description = &desc
	}
	return ds
}

func protoHealingSource(src world.HealingSource) *pb.HealingSource {
	if src == nil {
		return nil
	}
	desc := fmt.Sprint(src)
	hs := &pb.HealingSource{Type: fmt.Sprintf("%T", src)}
	if desc != "" {
		hs.Description = &desc
	}
	return hs
}

func protoEntityRef(e world.Entity) *pb.EntityRef {
	if e == nil {
		return nil
	}
	ref := &pb.EntityRef{Type: fmt.Sprintf("%T", e)}
	if handle := e.H(); handle != nil {
		ref.Uuid = handle.UUID().String()
	}
	ref.Position = protoVec3(e.Position())
	ref.Rotation = protoRotation(e.Rotation())
	return ref
}

func protoEntityRefs(list []world.Entity) []*pb.EntityRef {
	if len(list) == 0 {
		return nil
	}
	refs := make([]*pb.EntityRef, 0, len(list))
	for _, e := range list {
		if ref := protoEntityRef(e); ref != nil {
			refs = append(refs, ref)
		}
	}
	if len(refs) == 0 {
		return nil
	}
	return refs
}

func protoBlockPositions(list []cube.Pos) []*pb.BlockPos {
	if len(list) == 0 {
		return nil
	}
	positions := make([]*pb.BlockPos, 0, len(list))
	for _, pos := range list {
		positions = append(positions, protoBlockPos(pos))
	}
	return positions
}

func protoAddress(addr *net.UDPAddr) *pb.Address {
	if addr == nil {
		return nil
	}
	return &pb.Address{Host: addr.IP.String(), Port: int32(addr.Port)}
}

func protoSkinSummary(sk *skin.Skin) (fullID, playFabID string, persona bool) {
	if sk == nil {
		return "", "", false
	}
	return sk.FullID, sk.PlayFabID, sk.Persona
}

func applyDiagnosticsFields(evt *pb.PlayerDiagnosticsEvent, d session.Diagnostics) {
	evt.AverageFramesPerSecond = d.AverageFramesPerSecond
	evt.AverageServerSimTickTime = d.AverageServerSimTickTime
	evt.AverageClientSimTickTime = d.AverageClientSimTickTime
	evt.AverageBeginFrameTime = d.AverageBeginFrameTime
	evt.AverageInputTime = d.AverageInputTime
	evt.AverageRenderTime = d.AverageRenderTime
	evt.AverageEndFrameTime = d.AverageEndFrameTime
	evt.AverageRemainderTimePercent = d.AverageRemainderTimePercent
	evt.AverageUnaccountedTimePercent = d.AverageUnaccountedTimePercent
}

func worldFromContext(ctx *world.Context) *world.World {
	if ctx == nil {
		return nil
	}
	tx := ctx.Val()
	if tx == nil {
		return nil
	}
	return tx.World()
}
