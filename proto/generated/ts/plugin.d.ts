import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
export declare const protobufPackage = "df.plugin";
export declare enum EventType {
    EVENT_TYPE_UNSPECIFIED = 0,
    EVENT_TYPE_ALL = 1,
    PLAYER_JOIN = 10,
    PLAYER_QUIT = 11,
    PLAYER_MOVE = 12,
    PLAYER_JUMP = 13,
    PLAYER_TELEPORT = 14,
    PLAYER_CHANGE_WORLD = 15,
    PLAYER_TOGGLE_SPRINT = 16,
    PLAYER_TOGGLE_SNEAK = 17,
    CHAT = 18,
    PLAYER_FOOD_LOSS = 19,
    PLAYER_HEAL = 20,
    PLAYER_HURT = 21,
    PLAYER_DEATH = 22,
    PLAYER_RESPAWN = 23,
    PLAYER_SKIN_CHANGE = 24,
    PLAYER_FIRE_EXTINGUISH = 25,
    PLAYER_START_BREAK = 26,
    PLAYER_BLOCK_BREAK = 27,
    PLAYER_BLOCK_PLACE = 28,
    PLAYER_BLOCK_PICK = 29,
    PLAYER_ITEM_USE = 30,
    PLAYER_ITEM_USE_ON_BLOCK = 31,
    PLAYER_ITEM_USE_ON_ENTITY = 32,
    PLAYER_ITEM_RELEASE = 33,
    PLAYER_ITEM_CONSUME = 34,
    PLAYER_ATTACK_ENTITY = 35,
    PLAYER_EXPERIENCE_GAIN = 36,
    PLAYER_PUNCH_AIR = 37,
    PLAYER_SIGN_EDIT = 38,
    PLAYER_LECTERN_PAGE_TURN = 39,
    PLAYER_ITEM_DAMAGE = 40,
    PLAYER_ITEM_PICKUP = 41,
    PLAYER_HELD_SLOT_CHANGE = 42,
    PLAYER_ITEM_DROP = 43,
    PLAYER_TRANSFER = 44,
    COMMAND = 45,
    PLAYER_DIAGNOSTICS = 46,
    WORLD_LIQUID_FLOW = 70,
    WORLD_LIQUID_DECAY = 71,
    WORLD_LIQUID_HARDEN = 72,
    WORLD_SOUND = 73,
    WORLD_FIRE_SPREAD = 74,
    WORLD_BLOCK_BURN = 75,
    WORLD_CROP_TRAMPLE = 76,
    WORLD_LEAVES_DECAY = 77,
    WORLD_ENTITY_SPAWN = 78,
    WORLD_ENTITY_DESPAWN = 79,
    WORLD_EXPLOSION = 80,
    WORLD_CLOSE = 81,
    UNRECOGNIZED = -1
}
export declare function eventTypeFromJSON(object: any): EventType;
export declare function eventTypeToJSON(object: EventType): string;
export declare enum GameMode {
    SURVIVAL = 0,
    CREATIVE = 1,
    ADVENTURE = 2,
    SPECTATOR = 3,
    UNRECOGNIZED = -1
}
export declare function gameModeFromJSON(object: any): GameMode;
export declare function gameModeToJSON(object: GameMode): string;
export interface HostToPlugin {
    pluginId: string;
    hello?: HostHello | undefined;
    shutdown?: HostShutdown | undefined;
    event?: EventEnvelope | undefined;
}
export interface HostHello {
    apiVersion: string;
}
export interface HostShutdown {
    reason: string;
}
export interface EventEnvelope {
    eventId: string;
    type: EventType;
    playerJoin?: PlayerJoinEvent | undefined;
    playerQuit?: PlayerQuitEvent | undefined;
    chat?: ChatEvent | undefined;
    command?: CommandEvent | undefined;
    blockBreak?: BlockBreakEvent | undefined;
    worldClose?: WorldCloseEvent | undefined;
}
export interface PlayerJoinEvent {
    playerUuid: string;
    name: string;
}
export interface PlayerQuitEvent {
    playerUuid: string;
    name: string;
}
export interface ChatEvent {
    playerUuid: string;
    name: string;
    message: string;
}
export interface CommandEvent {
    playerUuid: string;
    name: string;
    /** Full command string like "/tp 100 64 200" */
    raw: string;
    /** Just the command name like "tp" */
    command: string;
    /** Parsed arguments like ["100", "64", "200"] */
    args: string[];
}
export interface BlockBreakEvent {
    playerUuid: string;
    name: string;
    world: string;
    x: number;
    y: number;
    z: number;
}
export interface WorldCloseEvent {
}
export interface PluginToHost {
    pluginId: string;
    hello?: PluginHello | undefined;
    subscribe?: EventSubscribe | undefined;
    actions?: ActionBatch | undefined;
    log?: LogMessage | undefined;
    eventResult?: EventResult | undefined;
}
export interface PluginHello {
    name: string;
    version: string;
    apiVersion: string;
    commands: CommandSpec[];
}
export interface CommandSpec {
    name: string;
    description: string;
    aliases: string[];
}
export interface EventSubscribe {
    events: EventType[];
}
export interface ActionBatch {
    actions: Action[];
}
export interface Action {
    correlationId?: string | undefined;
    sendChat?: SendChatAction | undefined;
    teleport?: TeleportAction | undefined;
    kick?: KickAction | undefined;
    setGameMode?: SetGameModeAction | undefined;
}
export interface SendChatAction {
    targetUuid: string;
    message: string;
}
export interface TeleportAction {
    playerUuid: string;
    x: number;
    y: number;
    z: number;
    yaw: number;
    pitch: number;
}
export interface KickAction {
    playerUuid: string;
    reason: string;
}
export interface SetGameModeAction {
    playerUuid: string;
    gameMode: GameMode;
}
export interface LogMessage {
    level: string;
    message: string;
}
export interface EventResult {
    eventId: string;
    cancel?: boolean | undefined;
    chat?: ChatMutation | undefined;
    blockBreak?: BlockBreakMutation | undefined;
}
export interface ChatMutation {
    message: string;
}
export interface BlockBreakMutation {
    drops: ItemStack[];
    xp?: number | undefined;
}
export interface ItemStack {
    name: string;
    meta: number;
    count: number;
}
export declare const HostToPlugin: MessageFns<HostToPlugin>;
export declare const HostHello: MessageFns<HostHello>;
export declare const HostShutdown: MessageFns<HostShutdown>;
export declare const EventEnvelope: MessageFns<EventEnvelope>;
export declare const PlayerJoinEvent: MessageFns<PlayerJoinEvent>;
export declare const PlayerQuitEvent: MessageFns<PlayerQuitEvent>;
export declare const ChatEvent: MessageFns<ChatEvent>;
export declare const CommandEvent: MessageFns<CommandEvent>;
export declare const BlockBreakEvent: MessageFns<BlockBreakEvent>;
export declare const WorldCloseEvent: MessageFns<WorldCloseEvent>;
export declare const PluginToHost: MessageFns<PluginToHost>;
export declare const PluginHello: MessageFns<PluginHello>;
export declare const CommandSpec: MessageFns<CommandSpec>;
export declare const EventSubscribe: MessageFns<EventSubscribe>;
export declare const ActionBatch: MessageFns<ActionBatch>;
export declare const Action: MessageFns<Action>;
export declare const SendChatAction: MessageFns<SendChatAction>;
export declare const TeleportAction: MessageFns<TeleportAction>;
export declare const KickAction: MessageFns<KickAction>;
export declare const SetGameModeAction: MessageFns<SetGameModeAction>;
export declare const LogMessage: MessageFns<LogMessage>;
export declare const EventResult: MessageFns<EventResult>;
export declare const ChatMutation: MessageFns<ChatMutation>;
export declare const BlockBreakMutation: MessageFns<BlockBreakMutation>;
export declare const ItemStack: MessageFns<ItemStack>;
export type PluginDefinition = typeof PluginDefinition;
export declare const PluginDefinition: {
    readonly name: "Plugin";
    readonly fullName: "df.plugin.Plugin";
    readonly methods: {
        readonly eventStream: {
            readonly name: "EventStream";
            readonly requestType: MessageFns<HostToPlugin>;
            readonly requestStream: true;
            readonly responseType: MessageFns<PluginToHost>;
            readonly responseStream: true;
            readonly options: {};
        };
    };
};
type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;
export type DeepPartial<T> = T extends Builtin ? T : T extends globalThis.Array<infer U> ? globalThis.Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export interface MessageFns<T> {
    encode(message: T, writer?: BinaryWriter): BinaryWriter;
    decode(input: BinaryReader | Uint8Array, length?: number): T;
    fromJSON(object: any): T;
    toJSON(message: T): unknown;
    create(base?: DeepPartial<T>): T;
    fromPartial(object: DeepPartial<T>): T;
}
export {};
