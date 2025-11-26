package plugin

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/customblock"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	pb "github.com/secmc/plugin/proto/generated/go"
)

type customBlock struct {
	id          string
	displayName string
	geometry    []byte
	textures    map[string]image.Image
	props       customblock.Properties
	baseHash    uint64
	// Optional permutable data (pack-only).
	states map[string][]any
	perms  []customblock.Permutation
	// Tracks whether a collision box was explicitly provided (nil in proto means 'no collision').
	hasCollisionBox bool
}

var _ world.CustomBlockBuildable = &customBlock{}

func (c *customBlock) EncodeBlock() (string, map[string]any) {
	return c.id, map[string]any{}
}

func (c *customBlock) Hash() (uint64, uint64) {
	if c.baseHash == 0 {
		c.baseHash = block.NextHash()
	}
	return c.baseHash, 0
}

// propertiesModel derives collision from a configured bounding box if provided.
type propertiesModel struct {
	box          cube.BBox
	hasCollision bool
}

func (m propertiesModel) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	if !m.hasCollision || m.box == (cube.BBox{}) {
		// No collision box provided: no collision.
		return nil
	}
	return []cube.BBox{m.box}
}

func (propertiesModel) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return true
}

func (c *customBlock) Model() world.BlockModel {
	// Always use propertiesModel so we can respect 'no collision' when the proto field is omitted.
	return propertiesModel{box: c.props.CollisionBox, hasCollision: c.hasCollisionBox}
}

func (c *customBlock) Properties() customblock.Properties {
	return c.props
}

func (c *customBlock) Name() string {
	return c.displayName
}

func (c *customBlock) Geometry() []byte {
	return c.geometry
}

func (c *customBlock) Textures() map[string]image.Image {
	return c.textures
}

// States implements block.Permutable (optional).
func (c *customBlock) States() map[string][]any {
	return c.states
}

// Permutations implements block.Permutable (optional).
func (c *customBlock) Permutations() []customblock.Permutation {
	return c.perms
}

// registerCustomBlocks registers custom blocks declared in PluginHello.
func (m *Manager) registerCustomBlocks(p *pluginProcess, defs []*pb.CustomBlockDefinition) {
	if len(defs) == 0 {
		return
	}
	pluginName := p.id
	if hello := p.helloInfo(); hello != nil && hello.Name != "" {
		pluginName = hello.Name
	}
	for _, def := range defs {
		if def == nil {
			continue
		}
		if err := m.registerSingleCustomBlock(def); err != nil {
			p.log.Error("failed to register custom block", "id", def.Id, "error", err)
			continue
		}
		m.log.Info("registered custom block", "plugin", pluginName, "id", def.Id, "name", def.DisplayName)
	}
}

func (m *Manager) registerSingleCustomBlock(def *pb.CustomBlockDefinition) error {
	if def.Id == "" {
		return fmt.Errorf("custom block ID cannot be empty")
	}
	if def.DisplayName == "" {
		return fmt.Errorf("custom block display name cannot be empty")
	}
	if def.Properties == nil {
		return fmt.Errorf("custom block properties cannot be empty")
	}

	// Decode textures
	images := make(map[string]image.Image, len(def.Textures))
	for _, t := range def.Textures {
		if t == nil || t.Name == "" || len(t.ImagePng) == 0 {
			continue
		}
		img, err := png.Decode(bytes.NewReader(t.ImagePng))
		if err != nil {
			return fmt.Errorf("decode texture %q: %w", t.Name, err)
		}
		images[t.Name] = img
	}

	// Build customblock.Properties
	props := customblock.Properties{
		Cube:        def.Properties.Cube,
		Geometry:    "",
		MapColour:   "",
		Rotation:    cube.Pos{},
		Scale:       mgl64.Vec3{},
		Textures:    nil,
		Translation: mgl64.Vec3{},
	}
	hasCollision := false
	if def.Properties.GeometryIdentifier != nil {
		props.Geometry = *def.Properties.GeometryIdentifier
	}
	if def.Properties.MapColour != nil {
		props.MapColour = *def.Properties.MapColour
	}
	if def.Properties.CollisionBox != nil {
		if bb, ok := bboxFromProto(def.Properties.CollisionBox); ok {
			props.CollisionBox = bb
			hasCollision = true
		}
	}
	if def.Properties.SelectionBox != nil {
		if bb, ok := bboxFromProto(def.Properties.SelectionBox); ok {
			props.SelectionBox = bb
		}
	}
	if def.Properties.Rotation != nil {
		if v, ok := vec3FromProto(def.Properties.Rotation); ok {
			props.Rotation = cube.Pos{int(v[0]), int(v[1]), int(v[2])}
		}
	}
	if def.Properties.Translation != nil {
		if v, ok := vec3FromProto(def.Properties.Translation); ok {
			props.Translation = mgl64.Vec3{v[0], v[1], v[2]}
		}
	}
	if def.Properties.Scale != nil {
		if v, ok := vec3FromProto(def.Properties.Scale); ok {
			props.Scale = mgl64.Vec3{v[0], v[1], v[2]}
		}
	}
	if len(def.Properties.Materials) > 0 {
		props.Textures = make(map[string]customblock.Material, len(def.Properties.Materials))
		for _, mat := range def.Properties.Materials {
			if mat == nil || mat.Target == "" || mat.TextureName == "" {
				continue
			}
			method := convertRenderMethod(mat.RenderMethod)
			material := customblock.NewMaterial(mat.TextureName, method)
			if mat.FaceDimming != nil && !*mat.FaceDimming {
				material = material.WithoutFaceDimming()
			}
			if mat.AmbientOcclusion != nil {
				if *mat.AmbientOcclusion {
					material = material.WithAmbientOcclusion()
				} else {
					material = material.WithoutAmbientOcclusion()
				}
			}
			props.Textures[mat.Target] = material
		}
	}

	cb := &customBlock{
		id:              def.Id,
		displayName:     def.DisplayName,
		geometry:        nil,
		textures:        images,
		props:           props,
		hasCollisionBox: hasCollision,
	}
	if def.GeometryJson != nil {
		cb.geometry = def.GeometryJson
	}

	// Permutable-only: parse client states and permutations from proto.
	if def.Properties != nil && len(def.Properties.States) > 0 {
		cb.states = make(map[string][]any, len(def.Properties.States))
		for name, list := range def.Properties.States {
			if list == nil || len(list.Values) == 0 {
				continue
			}
			values := make([]any, 0, len(list.Values))
			for _, raw := range list.Values {
				values = append(values, parsePropertyValue(raw))
			}
			cb.states[name] = values
		}
	}
	if def.Properties != nil && len(def.Properties.Permutations) > 0 {
		cb.perms = make([]customblock.Permutation, 0, len(def.Properties.Permutations))
		for _, perm := range def.Properties.Permutations {
			if perm == nil {
				continue
			}
			cb.perms = append(cb.perms, customblock.Permutation{
				Properties: convertCustomBlockProperties(perm.Properties),
				Condition:  perm.Condition,
			})
		}
	}

	world.RegisterBlock(cb)
	return nil
}

func convertRenderMethod(m pb.CustomBlockRenderMethod) customblock.Method {
	switch m {
	case pb.CustomBlockRenderMethod_CUSTOM_BLOCK_RENDER_METHOD_ALPHA_TEST:
		return customblock.AlphaTestRenderMethod()
	case pb.CustomBlockRenderMethod_CUSTOM_BLOCK_RENDER_METHOD_BLEND:
		return customblock.BlendRenderMethod()
	case pb.CustomBlockRenderMethod_CUSTOM_BLOCK_RENDER_METHOD_DOUBLE_SIDED:
		return customblock.DoubleSidedRenderMethod()
	default:
		return customblock.OpaqueRenderMethod()
	}
}

// convertCustomBlockProperties converts proto CustomBlockProperties to customblock.Properties.
func convertCustomBlockProperties(in *pb.CustomBlockProperties) customblock.Properties {
	if in == nil {
		return customblock.Properties{}
	}
	props := customblock.Properties{
		Cube:        in.Cube,
		Geometry:    "",
		MapColour:   "",
		Rotation:    cube.Pos{},
		Scale:       mgl64.Vec3{},
		Textures:    nil,
		Translation: mgl64.Vec3{},
	}
	if in.GeometryIdentifier != nil {
		props.Geometry = *in.GeometryIdentifier
	}
	if in.MapColour != nil {
		props.MapColour = *in.MapColour
	}
	if in.CollisionBox != nil {
		if bb, ok := bboxFromProto(in.CollisionBox); ok {
			props.CollisionBox = bb
		}
	}
	if in.SelectionBox != nil {
		if bb, ok := bboxFromProto(in.SelectionBox); ok {
			props.SelectionBox = bb
		}
	}
	if in.Rotation != nil {
		if v, ok := vec3FromProto(in.Rotation); ok {
			props.Rotation = cube.Pos{int(v[0]), int(v[1]), int(v[2])}
		}
	}
	if in.Translation != nil {
		if v, ok := vec3FromProto(in.Translation); ok {
			props.Translation = mgl64.Vec3{v[0], v[1], v[2]}
		}
	}
	if in.Scale != nil {
		if v, ok := vec3FromProto(in.Scale); ok {
			props.Scale = mgl64.Vec3{v[0], v[1], v[2]}
		}
	}
	if len(in.Materials) > 0 {
		props.Textures = make(map[string]customblock.Material, len(in.Materials))
		for _, mat := range in.Materials {
			if mat == nil || mat.Target == "" || mat.TextureName == "" {
				continue
			}
			method := convertRenderMethod(mat.RenderMethod)
			material := customblock.NewMaterial(mat.TextureName, method)
			if mat.FaceDimming != nil && !*mat.FaceDimming {
				material = material.WithoutFaceDimming()
			}
			if mat.AmbientOcclusion != nil {
				if *mat.AmbientOcclusion {
					material = material.WithAmbientOcclusion()
				} else {
					material = material.WithoutAmbientOcclusion()
				}
			}
			props.Textures[mat.Target] = material
		}
	}
	return props
}
