import { getDb } from '../db';
import { randomUUID } from 'crypto'; // Bun has crypto built-in

export class FactionData {
    id: string;
    name: string;
    leaderUuid: string;
    members: string[]; // Array of player UUIDs
    power: number;
    claimedChunks: string[]; // Array of chunk identifiers (e.g., "x,z")

    constructor(id: string, name: string, leaderUuid: string, members: string[] = [], power: number = 0, claimedChunks: string[] = []) {
        this.id = id;
        this.name = name;
        this.leaderUuid = leaderUuid;
        this.members = members;
        this.power = power;
        this.claimedChunks = claimedChunks;
    }

    static async getById(id: string): Promise<FactionData | null> {
        const db = getDb();
        const row = db.query(`SELECT * FROM factions WHERE id = ?`).get(id) as any;
        if (row) {
            return new FactionData(
                row.id,
                row.name,
                row.leader_uuid,
                JSON.parse(row.members_json),
                row.power,
                JSON.parse(row.claimed_chunks_json)
            );
        }
        return null;
    }

    static async getByName(name: string): Promise<FactionData | null> {
        const db = getDb();
        const row = db.query(`SELECT * FROM factions WHERE name = ?`).get(name) as any;
        if (row) {
            return new FactionData(
                row.id,
                row.name,
                row.leader_uuid,
                JSON.parse(row.members_json),
                row.power,
                JSON.parse(row.claimed_chunks_json)
            );
        }
        return null;
    }

    static async getByMember(playerUuid: string): Promise<FactionData | null> {
        const db = getDb();
        const rows = db.query(`SELECT * FROM factions`).all() as any[];
        for (const row of rows) {
            const faction = new FactionData(
                row.id,
                row.name,
                row.leader_uuid,
                JSON.parse(row.members_json),
                row.power,
                JSON.parse(row.claimed_chunks_json)
            );
            if (faction.members.includes(playerUuid)) {
                return faction;
            }
        }
        return null;
    }

    static async create(name: string, leaderUuid: string): Promise<FactionData> {
        const db = getDb();
        const id = randomUUID();
        const faction = new FactionData(id, name, leaderUuid, [leaderUuid], 1); // Leader is first member, initial power 1
        await faction.save();
        return faction;
    }

    async save(): Promise<void> {
        const db = getDb();
        db.query(`
            INSERT INTO factions (id, name, leader_uuid, members_json, power, claimed_chunks_json)
            VALUES (?, ?, ?, ?, ?, ?)
            ON CONFLICT(id) DO UPDATE SET
                name = EXCLUDED.name,
                leader_uuid = EXCLUDED.leader_uuid,
                members_json = EXCLUDED.members_json,
                power = EXCLUDED.power,
                claimed_chunks_json = EXCLUDED.claimed_chunks_json;
        `).run(
            this.id,
            this.name,
            this.leaderUuid,
            JSON.stringify(this.members),
            this.power,
            JSON.stringify(this.claimedChunks)
        );
    }

    addMember(uuid: string): void {
        if (!this.members.includes(uuid)) {
            this.members.push(uuid);
            this.power++; // Increase power for each member
        }
    }

    removeMember(uuid: string): void {
        const index = this.members.indexOf(uuid);
        if (index > -1) {
            this.members.splice(index, 1);
            this.power--; // Decrease power
        }
    }

    addClaim(chunkId: string): boolean {
        if (this.claimedChunks.includes(chunkId)) {
            return false; // Already claimed
        }
        this.claimedChunks.push(chunkId);
        this.power++; // Increase power for each claim
        return true;
    }

    removeClaim(chunkId: string): boolean {
        const index = this.claimedChunks.indexOf(chunkId);
        if (index > -1) {
            this.claimedChunks.splice(index, 1);
            this.power--; // Decrease power
            return true;
        }
        return false;
    }

    isChunkClaimed(chunkId: string): boolean {
        return this.claimedChunks.includes(chunkId);
    }
}
