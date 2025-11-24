<?php

namespace Dragonfly\PluginLib\Server;

use Df\Plugin\WorldRef;
use Dragonfly\PluginLib\Actions\Actions;
use Dragonfly\PluginLib\Entity\Player;

/**
 * Server provides access to online players and server-wide operations.
 * 
 * Player registry is automatically populated via PlayerJoin/PlayerQuit events
 * when using PluginBase.
 */
final class Server {
    /** @var array<string, Player> uuid => Player instance */
    private array $players = [];

    /** @var array<string, string> lowercase name => uuid */
    private array $nameIndex = [];

    public function __construct(
        private Actions $actions,
    ) {}

    /**
     * Register a player as online. Called automatically on PlayerJoin.
     */
    public function addPlayer(string $uuid, string $name, ?WorldRef $world = null): void {
        $this->players[$uuid] = new Player($uuid, $name, $this->actions, $world);
        $this->nameIndex[strtolower($name)] = $uuid;
    }

    /**
     * Update a player's world. Called on world change events.
     */
    public function setPlayerWorld(string $uuid, WorldRef $world): void {
        if (isset($this->players[$uuid])) {
            $this->players[$uuid]->setWorld($world);
        }
    }

    /**
     * Get a player's current world.
     */
    public function getPlayerWorld(string $uuid): ?WorldRef {
        return $this->players[$uuid]?->getWorld();
    }

    /**
     * Remove a player from the registry. Called automatically on PlayerQuit.
     */
    public function removePlayer(string $uuid): void {
        if (isset($this->players[$uuid])) {
            $name = $this->players[$uuid]->getName();
            unset($this->nameIndex[strtolower($name)]);
            unset($this->players[$uuid]);
        }
    }

    /**
     * Get a player by UUID. Returns null if not online.
     */
    public function getPlayer(string $uuid): ?Player {
        return $this->players[$uuid] ?? null;
    }

    /**
     * Get a player by name (case-insensitive). Returns null if not online.
     */
    public function getPlayerByName(string $name): ?Player {
        $lower = strtolower($name);
        if (!isset($this->nameIndex[$lower])) {
            return null;
        }
        return $this->players[$this->nameIndex[$lower]] ?? null;
    }

    /**
     * Get all online players.
     * 
     * @return Player[]
     */
    public function getOnlinePlayers(): array {
        return array_values($this->players);
    }

    /**
     * Get the count of online players.
     */
    public function getOnlineCount(): int {
        return count($this->players);
    }

    /**
     * Check if a player is online by UUID.
     */
    public function isOnline(string $uuid): bool {
        return isset($this->players[$uuid]);
    }

    /**
     * Check if a player is online by name (case-insensitive).
     */
    public function isOnlineByName(string $name): bool {
        return isset($this->nameIndex[strtolower($name)]);
    }

    /**
     * Broadcast a message to all online players.
     */
    public function broadcastMessage(string $message): void {
        foreach ($this->players as $player) {
            $player->sendMessage($message);
        }
    }
}
