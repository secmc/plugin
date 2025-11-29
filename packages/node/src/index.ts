export * from './plugin/PluginBase.js';
export * from './events/EventManager.js';
export * from './events/decorators.js';
export * from './commands/CommandManager.js';
export * from './commands/decorators.js';
export * from './entity/Player.js';

// Main plugin types
export * from './generated/plugin.js';

// Common types
export { GameMode, Vec3, Sound, CustomItemDefinition, CustomBlockDefinition } from './generated/common.js';
export { CommandSpec, CommandEvent, ParamType, ParamSpec } from './generated/command.js';

// Player events
export {
  BlockBreakEvent,
  ChatEvent,
  PlayerAttackEntityEvent,
  PlayerBlockPickEvent,
  PlayerBlockPlaceEvent,
  PlayerChangeWorldEvent,
  PlayerDeathEvent,
  PlayerDiagnosticsEvent,
  PlayerExperienceGainEvent,
  PlayerFireExtinguishEvent,
  PlayerFoodLossEvent,
  PlayerHealEvent,
  PlayerHeldSlotChangeEvent,
  PlayerHurtEvent,
  PlayerItemConsumeEvent,
  PlayerItemDamageEvent,
  PlayerItemDropEvent,
  PlayerItemPickupEvent,
  PlayerItemReleaseEvent,
  PlayerItemUseEvent,
  PlayerItemUseOnBlockEvent,
  PlayerItemUseOnEntityEvent,
  PlayerJoinEvent,
  PlayerJumpEvent,
  PlayerLecternPageTurnEvent,
  PlayerMoveEvent,
  PlayerPunchAirEvent,
  PlayerQuitEvent,
  PlayerRespawnEvent,
  PlayerSignEditEvent,
  PlayerSkinChangeEvent,
  PlayerStartBreakEvent,
  PlayerTeleportEvent,
  PlayerToggleSneakEvent,
  PlayerToggleSprintEvent,
  PlayerTransferEvent,
} from "./generated/player_events.js";

// World events
export {
  WorldBlockBurnEvent,
  WorldCloseEvent,
  WorldCropTrampleEvent,
  WorldEntityDespawnEvent,
  WorldEntitySpawnEvent,
  WorldExplosionEvent,
  WorldFireSpreadEvent,
  WorldLeavesDecayEvent,
  WorldLiquidDecayEvent,
  WorldLiquidFlowEvent,
  WorldLiquidHardenEvent,
  WorldSoundEvent,
} from "./generated/world_events.js";
