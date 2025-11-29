import { Database } from "bun:sqlite";

const dbPath = "./plugin.sqlite";
let db: Database | null = null;

export function getDb(): Database {
    if (!db) {
        db = new Database(dbPath);
    }
    return db;
}

export function initializeDatabase() {
    const database = getDb();
    
    // Create players table
    database.run(`
        CREATE TABLE IF NOT EXISTS players (
            uuid TEXT PRIMARY KEY,
            money INTEGER DEFAULT 0,
            combat_tag_end_time INTEGER DEFAULT 0,
            last_known_position_x REAL DEFAULT 0,
            last_known_position_y REAL DEFAULT 0,
            last_known_position_z REAL DEFAULT 0,
            last_known_spawn_status INTEGER DEFAULT 1
        );
    `);

    // Create factions table
    database.run(`
        CREATE TABLE IF NOT EXISTS factions (
            id TEXT PRIMARY KEY,
            name TEXT UNIQUE NOT NULL,
            leader_uuid TEXT NOT NULL,
            members_json TEXT DEFAULT '[]',
            power INTEGER DEFAULT 0,
            claimed_chunks_json TEXT DEFAULT '[]'
        );
    `);

    console.log("Database initialized successfully.");
}
