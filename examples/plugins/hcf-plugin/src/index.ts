import {
    PluginBase,
    On,
    EventType,
    EventContext,
    Player,
    RegisterCommand,
    GameMode,
    PlayerJoinEvent,
    PlayerMoveEvent,
    PlayerQuitEvent,
    PlayerAttackEntityEvent,
    BlockBreakEvent,
    PlayerBlockPlaceEvent,
    ParamType,
    Vec3,
} from '@dragonfly/proto';
import { getDb, initializeDatabase } from './db.js';
import { PlayerData } from './models/Player.js';
import { FactionData } from './models/Faction.js';
import {
    WorldService,
    SPAWN_CENTER_X,
    SPAWN_CENTER_Y,
    SPAWN_CENTER_Z,
    SPAWN_RADIUS,
    SPAWN_MAX_Y,
    SPAWN_MIN_Y
} from './services/WorldService.js';


class HCFPlugin extends PluginBase {
    private worldService: WorldService;

    constructor() {
        super();
        this.worldService = new WorldService(this);
    }

    onLoad(): void {
        console.log('[HCFPlugin] Loading...');
        initializeDatabase();
    }

    onEnable(): void {
        console.log('[HCFPlugin] Enabled.');
    }

    onDisable(): void {
        console.log('[HCFPlugin] Disabled.');
        // Clear all active combat walls on disable
        this.worldService.clearAllCombatWalls();
    }

    // Helper method to set world time
    private async setWorldTime(value: number | "day" | "night"): Promise<void> {
        let timeValue: number;
        if (typeof value === "string") {
            switch (value.toLowerCase()) {
                case "day": timeValue = 1000; break; // Minecraft day time
                case "night": timeValue = 13000; break; // Minecraft night time
                default: timeValue = 0; // Default to sunrise if unknown string
            }
        } else {
            timeValue = value;
        }

        await this.sendAction('worldSetTime', {
            time: timeValue,
            world: { name: '', dimension: 'overworld', id: '' },
        });
    }

    @On(EventType.PLAYER_JOIN)
    async onPlayerJoin(event: PlayerJoinEvent, context: EventContext<PlayerJoinEvent>) {
        const player = new Player(this, event.playerUuid);
        const playerData = await PlayerData.getOrCreate(event.playerUuid);
        
        // Always set initial spawn point
        await player.teleport(SPAWN_CENTER_X, SPAWN_CENTER_Y, SPAWN_CENTER_Z); 
        await player.sendMessage(`§aWelcome to the HCF server, ${event.name}!`);
        await player.sendMessage(`§aYour balance: $${playerData.money}.`);
        
        // Update player's spawn status on join
        playerData.lastKnownSpawnStatus = this.worldService.isPositionInSpawn(SPAWN_CENTER_X, SPAWN_CENTER_Y, SPAWN_CENTER_Z);
        // Also update last known position to spawn
        playerData.lastKnownPositionX = SPAWN_CENTER_X;
        playerData.lastKnownPositionY = SPAWN_CENTER_Y;
        playerData.lastKnownPositionZ = SPAWN_CENTER_Z;

        await playerData.save();

        await context.ack();
    }

    @On(EventType.PLAYER_QUIT)
    async onPlayerQuit(event: PlayerQuitEvent, context: EventContext<PlayerQuitEvent>) {
        console.log(`[HCFPlugin] Player ${event.name} (${event.playerUuid}) quit.`);
        // Ensure to remove any active wall for this player on quit
        await this.worldService.removeSpawnWall(event.playerUuid);
        await context.ack();
    }

    @On(EventType.PLAYER_ATTACK_ENTITY)
    async onPlayerAttackEntity(event: PlayerAttackEntityEvent, context: EventContext<PlayerAttackEntityEvent>) {
        const attackerPlayer = await PlayerData.getOrCreate(event.playerUuid);
        const victimUuid = event.entity?.uuid;

        if (victimUuid) {
            const victimPlayer = await PlayerData.get(victimUuid);
            if (victimPlayer) {
                await attackerPlayer.setCombatTag(15);
                await victimPlayer.setCombatTag(15);
                new Player(this, attackerPlayer.uuid).sendPopup('§cCombat Tagged!');
                new Player(this, victimPlayer.uuid).sendPopup('§cCombat Tagged!');
            }
        }
        await context.ack();
    }

    @On(EventType.PLAYER_MOVE)
    async onPlayerMove(event: PlayerMoveEvent, context: EventContext<PlayerMoveEvent>) {
        const player = new Player(this, event.playerUuid);
        const playerData = await PlayerData.getOrCreate(event.playerUuid);
        
        const newPosX = event.position?.x ?? playerData.lastKnownPositionX;
        const newPosY = event.position?.y ?? playerData.lastKnownPositionY;
        const newPosZ = event.position?.z ?? playerData.lastKnownPositionZ;

        const currentPos = { x: newPosX, y: newPosY, z: newPosZ };
        const isCurrentlyInSpawn = this.worldService.isPositionInSpawn(currentPos.x, currentPos.y, currentPos.z);
        const wasLastInSpawn = playerData.lastKnownSpawnStatus;

        // 1. Combat Tag Information (Console Log)
        if (playerData.isCombatTagged()) {
            const remaining = playerData.getCombatTagRemaining();
        }

        // 2. Spawn Entry/Exit Messages
        if (isCurrentlyInSpawn && !wasLastInSpawn) {
            await player.sendMessage('§aYou have entered the spawn area.');
            // Remove wall if they somehow made it in
            await this.worldService.removeSpawnWall(event.playerUuid);
        } else if (!isCurrentlyInSpawn && wasLastInSpawn) {
            await player.sendMessage('§cYou have left the spawn area and entered the wilderness.');
            // Remove wall if they were combat tagged but now left spawn
            await this.worldService.removeSpawnWall(event.playerUuid);
        }

        // 3. Combat Tag Spawn Restriction and Dynamic Wall
        if (playerData.isCombatTagged()) {
            if (isCurrentlyInSpawn && !wasLastInSpawn) { // Combat tagged trying to enter spawn
                await player.sendMessage('§cYou cannot enter spawn while combat tagged!');
                context.cancel(); // Cancel the move
                await player.teleport(
                    playerData.lastKnownPositionX,
                    playerData.lastKnownPositionY,
                    playerData.lastKnownPositionZ
                );
                await this.worldService.removeSpawnWall(event.playerUuid); // Ensure wall is removed if they fail to enter
                await context.ack();
                return; // Prevent further processing
            } else if (this.worldService.shouldShowSpawnWall(currentPos.x, currentPos.y, currentPos.z)) {
                // Player is combat tagged, outside spawn, and close to the border, show/update wall
                 await this.worldService.buildSpawnWall(event.playerUuid, currentPos);
            } else { // Player is combat tagged, but not near spawn border or in spawn
                await this.worldService.removeSpawnWall(event.playerUuid);
            }
        } else {
            // Player is not combat tagged, ensure no wall is shown
            await this.worldService.removeSpawnWall(event.playerUuid);
        }

        // Update player's last known position and spawn status
        if (event.position) {
            playerData.lastKnownPositionX = newPosX;
            playerData.lastKnownPositionY = newPosY;
            playerData.lastKnownPositionZ = newPosZ;
        }
        playerData.lastKnownSpawnStatus = isCurrentlyInSpawn;
        await playerData.save();

        await context.ack();
    }

    // --- Commands ---

    @RegisterCommand({
        name: 'money',
        description: 'Check your money balance.',
        aliases: ['bal', 'balance'],
    })
    async onMoneyCommand(uuid: string, args: string[], context: EventContext<any>) {
        const player = new Player(this, uuid);
        const playerData = await PlayerData.getOrCreate(uuid);
        await player.sendMessage(`§aYour balance: $${playerData.money}.`);
        await context.ack();
    }

    @RegisterCommand({
        name: 'time',
        description: 'Set the world time.',
        params: [
            {
                name: 'value',
                description: 'Time value (day, night, or number)',
                type: ParamType.PARAM_STRING,
                optional: false,
                enumValues: ['day', 'night', '0', '6000', '12000', '18000'], // Common time values
            },
        ],
    })
    async onTimeCommand(uuid: string, args: string[], context: EventContext<any>) {
        const player = new Player(this, uuid);
        if (args.length === 0) {
            await player.sendMessage('§cUsage: /time set <value>');
            await context.ack();
            return;
        }

        const timeArg = args[0].toLowerCase();
        let timeValue: number | string;

        if (timeArg === 'day' || timeArg === 'night') {
            timeValue = timeArg;
        } else {
            const parsedTime = parseInt(timeArg);
            if (isNaN(parsedTime)) {
                await player.sendMessage('§cInvalid time value. Use "day", "night", or a number.');
                await context.ack();
                return;
            }
            timeValue = parsedTime;
        }

        await this.setWorldTime(timeValue as any);
        await player.sendMessage(`§aWorld time set to ${timeArg}.`);
        await context.ack();
    }

    @RegisterCommand({
        name: 'gamemode',
        description: 'Change your gamemode.',
        aliases: ['gm'],
        params: [
            {
                name: 'mode',
                description: 'The gamemode (survival, creative, adventure, spectator)',
                type: ParamType.PARAM_ENUM,
                optional: false,
                enumValues: ['survival', 'creative', 'adventure', 'spectator', '0', '1', '2', '3']
            },
        ]
    })
    async onGamemodeCommand(uuid: string, args: string[], context: EventContext<any>) {
        const player = new Player(this, uuid);
        if (args.length === 0) {
            await player.sendMessage('§cUsage: /gamemode <mode>');
            await context.ack();
            return;
        }

        const modeArg = args[0].toLowerCase();
        let gamemode: GameMode | undefined;

        switch (modeArg) {
            case 'survival':
            case '0':
                gamemode = GameMode.SURVIVAL;
                break;
            case 'creative':
            case '1':
                gamemode = GameMode.CREATIVE;
                break;
            case 'adventure':
            case '2':
                gamemode = GameMode.ADVENTURE;
                break;
            case 'spectator':
            case '3':
                gamemode = GameMode.SPECTATOR;
                break;
            default:
                await player.sendMessage('§cInvalid gamemode: ' + modeArg);
                await context.ack();
                return;
        }

        if (gamemode !== undefined) {
            await player.setGameMode(gamemode);
            await player.sendMessage(`§aYour gamemode has been set to ${modeArg}.`);
        } else {
            await player.sendMessage('§cAn unexpected error occurred.');
        }
        await context.ack();
    }

    @RegisterCommand({
        name: 'give',
        description: 'Give yourself an item.',
        aliases: ['i'],
        params: [
            { name: 'item', description: 'The item ID (e.g., minecraft:dirt)', type: ParamType.PARAM_STRING, optional: false },
            { name: 'amount', description: 'The amount to give (default: 1)', type: ParamType.PARAM_INT, optional: true },
        ]
    })
    async onGiveCommand(uuid: string, args: string[], context: EventContext<any>) {
        const player = new Player(this, uuid);
        if (args.length === 0) {
            await player.sendMessage('§cUsage: /give <item_id> [amount]');
            await context.ack();
            return;
        }

        const itemId = args[0];
        const amount = args.length > 1 ? parseInt(args[1]) : 1;

        if (isNaN(amount) || amount <= 0) {
            await player.sendMessage('§cInvalid amount. Must be a positive number.');
            await context.ack();
            return;
        }

        try {
            await player.giveItem(itemId, amount);
            await player.sendMessage(`§aGiven ${amount} of ${itemId}.`);
        } catch (error) {
            console.error(`Error giving item ${itemId} to ${uuid}:`, error);
            await player.sendMessage(`§cFailed to give item ${itemId}.`);
        }
        await context.ack();
    }

    @RegisterCommand({
        name: 'combattag',
        description: 'Test command: Gives yourself combat tag for 10 seconds.',
    })
    async onCombatTagCommand(uuid: string, args: string[], context: EventContext<any>) {
        const player = new Player(this, uuid);
        const playerData = await PlayerData.getOrCreate(uuid);
        await playerData.setCombatTag(10);
        await player.sendMessage('§cYou have been combat tagged for 10 seconds!');
        await context.ack();
    }

    @RegisterCommand({
        name: 'f',
        description: 'Faction commands.',
        aliases: ['faction'],
        params: [
            { name: 'subcommand', description: 'create, claim, info', type: ParamType.PARAM_ENUM, optional: false, enumValues: ['create', 'claim', 'info'] },
            { name: 'name', description: 'Faction name (for create/info)', type: ParamType.PARAM_STRING, optional: true },
        ]
    })
    async onFactionCommand(uuid: string, args: string[], context: EventContext<any>) {
        const player = new Player(this, uuid);
        const playerData = await PlayerData.getOrCreate(uuid);

        if (args.length === 0) {
            await player.sendMessage('§cUsage: /f <create|claim|info> [args]');
            await context.ack();
            return;
        }

        const subCommand = args[0].toLowerCase();

        switch (subCommand) {
            case 'create':
                if (args.length < 2) {
                    await player.sendMessage('§cUsage: /f create <name>');
                    break;
                }
                const factionName = args[1];
                let existingFaction = await FactionData.getByName(factionName);
                if (existingFaction) {
                    await player.sendMessage(`§cFaction '${factionName}' already exists.`);
                    break;
                }
                const newFaction = await FactionData.create(factionName, uuid);
                await player.sendMessage(`§aFaction '${newFaction.name}' created!`);
                break;

            case 'claim':
                const playerPos = { x: playerData.lastKnownPositionX, y: playerData.lastKnownPositionY, z: playerData.lastKnownPositionZ };
                const chunkId = `${Math.floor(playerPos.x / 16)},${Math.floor(playerPos.z / 16)}`;
                
                let playerFaction = await FactionData.getByMember(uuid);
                if (!playerFaction) { 
                    await player.sendMessage('§cYou are not in a faction. Use /f create <name> to make one.'); 
                    break; 
                }
                if (playerFaction.leaderUuid !== uuid) {
                    await player.sendMessage('§cOnly the faction leader can claim land.');
                    break;
                }

                // Check if chunk is already claimed by another faction
                const allFactionsForClaimCheck = await this.getAllFactions();
                for (const otherFaction of allFactionsForClaimCheck) {
                    if (otherFaction.id !== playerFaction.id && otherFaction.isChunkClaimed(chunkId)) {
                        await player.sendMessage(`§cThis chunk is already claimed by ${otherFaction.name}.`);
                        break;
                    }
                }

                if (playerFaction.addClaim(chunkId)) {
                    await playerFaction.save();
                    await player.sendMessage(`§aClaimed chunk ${chunkId} for your faction.`);
                } else {
                    await player.sendMessage(`§cChunk ${chunkId} is already claimed by your faction.`);
                }
                break;

            case 'info':
                const infoFactionName = args.length > 1 ? args[1] : undefined;
                let factionToDisplay: FactionData | null = null;

                if (infoFactionName) {
                    factionToDisplay = await FactionData.getByName(infoFactionName);
                } else {
                    factionToDisplay = await FactionData.getByMember(uuid);
                }

                if (factionToDisplay) {
                    await player.sendMessage(`§b--- Faction Info: ${factionToDisplay.name} ---`);
                    await player.sendMessage(`§bLeader: ${factionToDisplay.leaderUuid}`); // Replace with name if Player API supports it
                    await player.sendMessage(`§bMembers: ${factionToDisplay.members.length}`);
                    await player.sendMessage(`§bPower: ${factionToDisplay.power}`);
                    await player.sendMessage(`§bClaims: ${factionToDisplay.claimedChunks.length}`);
                } else {
                    await player.sendMessage('§cFaction not found or you are not in one.');
                }
                break;

            default:
                await player.sendMessage('§cUnknown faction subcommand.');
        }
        await context.ack();
    }

    // --- Event Handlers for HCF Logic ---

    @On(EventType.PLAYER_BLOCK_BREAK)
    async onBlockBreak(event: BlockBreakEvent, context: EventContext<BlockBreakEvent>) {
        if (!event.position) {
            await context.ack();
            return;
        }
        const chunkId = `${Math.floor(event.position.x! / 16)},${Math.floor(event.position.z! / 16)}`;
        const allFactions = await this.getAllFactions();

        let isClaimed = false;
        let claimingFaction: FactionData | null = null;
        for (const faction of allFactions) {
            if (faction.isChunkClaimed(chunkId)) {
                isClaimed = true;
                claimingFaction = faction;
                break;
            }
        }

        if (isClaimed && claimingFaction) {
            if (!claimingFaction.members.includes(event.playerUuid)) {
                await new Player(this, event.playerUuid).sendMessage(`§cThis land is claimed by ${claimingFaction.name}!`);
                context.cancel();
            }
        }
        await context.ack();
    }

    @On(EventType.PLAYER_BLOCK_PLACE)
    async onBlockPlace(event: BlockBreakEvent, context: EventContext<BlockBreakEvent>) { // Corrected EventContext type
        if (!event.position) {
            await context.ack();
            return;
        }
        const chunkId = `${Math.floor(event.position.x! / 16)},${Math.floor(event.position.z! / 16)}`;
        const allFactions = await this.getAllFactions();

        let isClaimed = false;
        let claimingFaction: FactionData | null = null;
        for (const faction of allFactions) {
            if (faction.isChunkClaimed(chunkId)) {
                isClaimed = true;
                claimingFaction = faction;
                break;
            }
        }

        if (isClaimed && claimingFaction) {
            if (!claimingFaction.members.includes(event.playerUuid)) {
                await new Player(this, event.playerUuid).sendMessage(`§cThis land is claimed by ${claimingFaction.name}!`);
                context.cancel();
            }
        }
        await context.ack();
    }

    // Helper to get all factions - Placeholder
    private async getAllFactions(): Promise<FactionData[]> {
        const db = getDb(); // Corrected to use getDb()
        const rows = db.query(`SELECT * FROM factions`).all() as any[];
        return rows.map(row => new FactionData(
            row.id,
            row.name,
            row.leader_uuid,
            JSON.parse(row.members_json),
            row.power,
            JSON.parse(row.claimed_chunks_json)
        ));
    }
}

new HCFPlugin();