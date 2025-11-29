import { PluginBase } from '@dragonfly/proto'; 
import { Player } from '@dragonfly/proto'; // For Player class to send messages
import { Action } from 'node_modules/@dragonfly/proto/dist/generated/actions.js';

// --- HCF Plugin Constants ---
export const SPAWN_CENTER_X = 0;
export const SPAWN_CENTER_Y = -60;
export const SPAWN_CENTER_Z = 0;
export const SPAWN_RADIUS = 50; // A radius of 50 makes a 100x100 area (x: -50 to 50, z: -50 to 50) around the center.
// Max Y for spawn vertical extent. Adjust as needed for your world's build height.
export const SPAWN_MAX_Y = 256; 
export const SPAWN_MIN_Y = -64; // Min Y for spawn vertical extent

export class WorldService {
    private plugin: PluginBase;
    private activeCombatWalls: Map<string, Set<string>> = new Map();

    constructor(plugin: PluginBase) {
        this.plugin = plugin;
    }

    // Helper to check if a position is within the spawn area
    public isPositionInSpawn(x: number, y: number, z: number): boolean {
        const minX = SPAWN_CENTER_X - SPAWN_RADIUS;
        const maxX = SPAWN_CENTER_X + SPAWN_RADIUS;
        const minZ = SPAWN_CENTER_Z - SPAWN_RADIUS;
        const maxZ = SPAWN_CENTER_Z + SPAWN_RADIUS;

        return x >= minX && x <= maxX && z >= minZ && z <= maxZ && y >= SPAWN_MIN_Y && y <= SPAWN_MAX_Y;
    }

    // Clears all active combat walls for all players
    public async clearAllCombatWalls(): Promise<void> {
        for (const playerUuid of this.activeCombatWalls.keys()) {
            await this.removeSpawnWall(playerUuid);
        }
    }

    // Helper to remove the dynamic spawn wall for a player
    public async removeSpawnWall(playerUuid: string): Promise<void> {
        const currentActiveBlocks = this.activeCombatWalls.get(playerUuid);
        if (!currentActiveBlocks) return;

        const actions: Action[] = [];
        for (const blockKey of currentActiveBlocks) {
            const [x, y, z] = blockKey.split(',').map(Number);
            actions.push({
                worldSetBlock: {
                    world: { name: '', dimension: 'overworld', id: '' }, // Assuming overworld
                    position: { x, y, z },
                    block: { name: 'minecraft:air', properties: {} },
                },
            });
        }
        this.activeCombatWalls.delete(playerUuid); // Clear entry

        if (actions.length > 0) {
            await this.plugin.send({
                pluginId: this.plugin.pluginId,
                actions: { actions }
            });
        }
    }

    // Helper to determine if the wall should be shown around the player
    public shouldShowSpawnWall(x: number, y: number, z: number): boolean {
        const WALL_PROXIMITY_THRESHOLD = 20; // Player must be within 20 blocks of the spawn border

        const minX = SPAWN_CENTER_X - SPAWN_RADIUS;
        const maxX = SPAWN_CENTER_X + SPAWN_RADIUS;
        const minZ = SPAWN_CENTER_Z - SPAWN_RADIUS;
        const maxZ = SPAWN_CENTER_Z + SPAWN_RADIUS;

        // Check if player is outside spawn but within proximity to any of its borders
        const outsideSpawn = !this.isPositionInSpawn(x, y, z);
        const nearXBorder = (x >= minX - WALL_PROXIMITY_THRESHOLD && x <= minX) || (x <= maxX + WALL_PROXIMITY_THRESHOLD && x >= maxX);
        const nearZBorder = (z >= minZ - WALL_PROXIMITY_THRESHOLD && z <= minZ) || (z <= maxZ + WALL_PROXIMITY_THRESHOLD && z >= maxZ);
        const inVerticalBounds = y >= SPAWN_MIN_Y && y <= SPAWN_MAX_Y; // Still check vertical bounds

        return outsideSpawn && (nearXBorder || nearZBorder) && inVerticalBounds;
    }

    // Helper to build the dynamic spawn wall for a player
    public async buildSpawnWall(playerUuid: string, playerPos: { x: number, y: number, z: number }): Promise<void> {
        let currentActiveBlocks = this.activeCombatWalls.get(playerUuid);
        if (!currentActiveBlocks) {
            currentActiveBlocks = new Set<string>();
            this.activeCombatWalls.set(playerUuid, currentActiveBlocks);
        }

        const actions: Action[] = [];
        const newBlocks = new Set<string>();
        const WALL_SEGMENT_WIDTH = 20; // Width of the wall segment along the perimeter
        const WALL_SEGMENT_DEPTH = 1; // Depth of the wall segment extending outwards from spawn
        const WALL_HEIGHT_UNITS = 20; // Vertical extent of the wall

        const wallYRangeMin = Math.floor(playerPos.y)
        const wallYRangeMax = Math.min(SPAWN_MAX_Y, Math.floor(playerPos.y) + Math.ceil(WALL_HEIGHT_UNITS / 2));

        const spawnMinX = SPAWN_CENTER_X - SPAWN_RADIUS;
        const spawnMaxX = SPAWN_CENTER_X + SPAWN_RADIUS;
        const spawnMinZ = SPAWN_CENTER_Z - SPAWN_RADIUS;
        const spawnMaxZ = SPAWN_CENTER_Z + SPAWN_RADIUS;
        
        const WALL_PROXIMITY_THRESHOLD = 20; // Player must be within this many blocks of a wall to trigger segment

        const segmentsToBuild: { fixedCoord: number; isXFixed: boolean; startVaryingCoord: number; endVaryingCoord: number; depthDirection: -1 | 1; }[] = [];

        // Check West wall (fixed X = spawnMinX)
        if (playerPos.x >= spawnMinX - WALL_PROXIMITY_THRESHOLD && playerPos.x <= spawnMinX) {
            segmentsToBuild.push({
                fixedCoord: spawnMinX,
                isXFixed: true,
                startVaryingCoord: Math.floor(playerPos.z - WALL_SEGMENT_WIDTH / 2),
                endVaryingCoord: Math.floor(playerPos.z + WALL_SEGMENT_WIDTH / 2),
                depthDirection: -1, // Extend towards negative X (outwards from minX)
            });
        }
        // Check East wall (fixed X = spawnMaxX)
        if (playerPos.x <= spawnMaxX + WALL_PROXIMITY_THRESHOLD && playerPos.x >= spawnMaxX) {
            segmentsToBuild.push({
                fixedCoord: spawnMaxX,
                isXFixed: true,
                startVaryingCoord: Math.floor(playerPos.z - WALL_SEGMENT_WIDTH / 2),
                endVaryingCoord: Math.floor(playerPos.z + WALL_SEGMENT_WIDTH / 2),
                depthDirection: 1, // Extend towards positive X (outwards from maxX)
            });
        }
        // Check South wall (fixed Z = spawnMinZ)
        if (playerPos.z >= spawnMinZ - WALL_PROXIMITY_THRESHOLD && playerPos.z <= spawnMinZ) {
            segmentsToBuild.push({
                fixedCoord: spawnMinZ,
                isXFixed: false,
                startVaryingCoord: Math.floor(playerPos.x - WALL_SEGMENT_WIDTH / 2),
                endVaryingCoord: Math.floor(playerPos.x + WALL_SEGMENT_WIDTH / 2),
                depthDirection: -1, // Extend towards negative Z (outwards from minZ)
            });
        }
        // Check North wall (fixed Z = spawnMaxZ)
        if (playerPos.z <= spawnMaxZ + WALL_PROXIMITY_THRESHOLD && playerPos.z >= spawnMaxZ) {
            segmentsToBuild.push({
                fixedCoord: spawnMaxZ,
                isXFixed: false,
                startVaryingCoord: Math.floor(playerPos.x - WALL_SEGMENT_WIDTH / 2),
                endVaryingCoord: Math.floor(playerPos.x + WALL_SEGMENT_WIDTH / 2),
                depthDirection: 1, // Extend towards positive Z (outwards from maxZ)
            });
        }

        for (const segment of segmentsToBuild) {
            for (let y = wallYRangeMin; y < wallYRangeMax; y++) {
                if (segment.isXFixed) { // North/South oriented wall (fixed X, varying Z)
                    // Clamp varying coord (z) to stay within spawn Z bounds
                    const clampedStartZ = Math.max(segment.startVaryingCoord, spawnMinZ);
                    const clampedEndZ = Math.min(segment.endVaryingCoord, spawnMaxZ);
                    for (let z = clampedStartZ; z <= clampedEndZ; z++) {
                        for (let depth = 0; depth < WALL_SEGMENT_DEPTH; depth++) {
                            const actualX = segment.fixedCoord + (depth * segment.depthDirection);
                            newBlocks.add(`${actualX},${y},${z}`);
                        }
                    }
                } else { // East/West oriented wall (fixed Z, varying X)
                    const clampedStartX = Math.max(segment.startVaryingCoord, spawnMinX);
                    const clampedEndX = Math.min(segment.endVaryingCoord, spawnMaxX);
                    for (let x = clampedStartX; x <= clampedEndX; x++) {
                        for (let depth = 0; depth < WALL_SEGMENT_DEPTH; depth++) {
                            const actualZ = segment.fixedCoord + (depth * segment.depthDirection);
                            newBlocks.add(`${x},${y},${actualZ}`);
                        }
                    }
                }
            }
        }

        // Add blocks that are new to the current view
        for (const blockKey of newBlocks) {
            if (!currentActiveBlocks.has(blockKey)) {
                const [x, y, z] = blockKey.split(',').map(Number);
                actions.push({
                    worldSetBlock: {
                        world: { name: '', dimension: 'overworld', id: '' }, // Assuming overworld
                        position: { x, y, z },
                        block: { name: 'minecraft:red_stained_glass', properties: {} }, // Use red glass for visibility
                    }
                });
                currentActiveBlocks.add(blockKey);
            }
        }

        // Remove blocks that are no longer in view
        for (const blockKey of currentActiveBlocks) {
            if (!newBlocks.has(blockKey)) {
                const [x, y, z] = blockKey.split(',').map(Number);
                actions.push({
                    worldSetBlock: {
                        world: { name: '', dimension: 'overworld', id: '' },
                        position: { x, y, z },
                        block: { name: 'minecraft:air', properties: {} },
                    }
                });
            }
        }
        this.activeCombatWalls.set(playerUuid, newBlocks); // Update active blocks

        if (actions.length > 0) {
            await this.plugin.send({
                pluginId: this.plugin.pluginId,
                actions: { actions }
            });
        }
    }

    // World Set Time action
    public async setTime(value: number | "day" | "night"): Promise<void> {
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

        await this.plugin.sendAction('worldSetTime', {
            time: timeValue,
            world: { name: '', dimension: 'overworld', id: '' },
        });
    }
}