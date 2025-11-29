import { getDb } from '../db.js';

export class PlayerData {
    uuid: string;
    money: number;
    combatTagEndTime: number; // Unix timestamp when combat tag ends
    lastKnownPositionX: number;
    lastKnownPositionY: number;
    lastKnownPositionZ: number;
    lastKnownSpawnStatus: boolean; // true if player was last known to be in spawn

    constructor(
        uuid: string,
        money: number = 0,
        combatTagEndTime: number = 0,
        lastKnownPositionX: number = 0,
        lastKnownPositionY: number = -60, // Default spawn Y
        lastKnownPositionZ: number = 0,
        lastKnownSpawnStatus: boolean = true // Assume new players start in spawn
    ) {
        this.uuid = uuid;
        this.money = money;
        this.combatTagEndTime = combatTagEndTime;
        this.lastKnownPositionX = lastKnownPositionX;
        this.lastKnownPositionY = lastKnownPositionY;
        this.lastKnownPositionZ = lastKnownPositionZ;
        this.lastKnownSpawnStatus = lastKnownSpawnStatus;
    }

    static async get(uuid: string): Promise<PlayerData | null> {
        const db = getDb();
        const row = db.query(`SELECT * FROM players WHERE uuid = ?`).get(uuid) as any | null;
        if (row) {
            return new PlayerData(
                row.uuid,
                row.money,
                row.combat_tag_end_time,
                row.last_known_position_x,
                row.last_known_position_y,
                row.last_known_position_z,
                Boolean(row.last_known_spawn_status)
            );
        }
        return null;
    }

    static async getOrCreate(uuid: string): Promise<PlayerData> {
        let player = await PlayerData.get(uuid);
        if (!player) {
            player = new PlayerData(uuid);
            await player.save();
        }
        return player;
    }

    async save(): Promise<void> {
        const db = getDb();
        db.query(`
            INSERT INTO players (uuid, money, combat_tag_end_time, last_known_position_x, last_known_position_y, last_known_position_z, last_known_spawn_status)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(uuid) DO UPDATE SET
                money = EXCLUDED.money,
                combat_tag_end_time = EXCLUDED.combat_tag_end_time,
                last_known_position_x = EXCLUDED.last_known_position_x,
                last_known_position_y = EXCLUDED.last_known_position_y,
                last_known_position_z = EXCLUDED.last_known_position_z,
                last_known_spawn_status = EXCLUDED.last_known_spawn_status;
        `).run(
            this.uuid,
            this.money,
            this.combatTagEndTime,
            this.lastKnownPositionX,
            this.lastKnownPositionY,
            this.lastKnownPositionZ,
            Number(this.lastKnownSpawnStatus) // Convert boolean to number for SQLite
        );
    }

    async addMoney(amount: number): Promise<void> {
        if (amount <= 0) return;
        this.money += amount;
        await this.save();
    }

    async removeMoney(amount: number): Promise<boolean> {
        if (amount <= 0 || this.money < amount) return false;
        this.money -= amount;
        await this.save();
        return true;
    }

    async setCombatTag(durationSeconds: number): Promise<void> {
        this.combatTagEndTime = Math.floor(Date.now() / 1000) + durationSeconds;
        await this.save();
    }

    isCombatTagged(): boolean {
        return this.combatTagEndTime > Math.floor(Date.now() / 1000);
    }

    getCombatTagRemaining(): number {
        const remaining = this.combatTagEndTime - Math.floor(Date.now() / 1000);
        return Math.max(0, remaining);
    }
}
