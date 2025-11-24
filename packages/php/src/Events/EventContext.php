<?php

namespace Dragonfly\PluginLib\Events;

use Df\Plugin\EventResult;
use Dragonfly\PluginLib\Actions\Actions;
use Dragonfly\PluginLib\Actions\ActionsTrait;
use Dragonfly\PluginLib\Entity\Player;
use Dragonfly\PluginLib\Server\Server;
use Dragonfly\PluginLib\StreamSender;

final class EventContext {
    use MutationsTrait;
    use ActionsTrait;

    private bool $handled = false;
    private ?Actions $actions = null;
    private ?Player $player = null;

    public function __construct(
        private string $pluginId,
        private string $eventId,
        private StreamSender $sender,
        private Server $server,
        private bool $expectsResponse,
        private ?object $payload = null,
    ) {}

    protected function getActions(): Actions {
        return $this->actions ??= new Actions($this->sender, $this->pluginId);
    }

    public function respond(callable $mutator): void {
        if ($this->handled) {
            throw new \LogicException('Event already handled.');
        }
        $this->sender->respond($this->pluginId, $this->eventId, $mutator);
        $this->handled = true;
    }

    public function cancel(): void {
        if ($this->handled) {
            throw new \LogicException('Event already handled.');
        }
        $this->sender->respond($this->pluginId, $this->eventId, function (EventResult $r): void {
            $r->setCancel(true);
        });
        $this->handled = true;
    }

    public function ackIfUnhandled(): void {
        if (!$this->handled) {
            if (!$this->expectsResponse) {
                $this->handled = true;
                return;
            }
            $this->sender->respond($this->pluginId, $this->eventId, function (EventResult $r): void {
                $r->setCancel(false);
            });
            $this->handled = true;
        }
    }

    public function respondWith(object $mutation): void {
        $class = get_class($mutation);
        if (!str_starts_with($class, 'Df\\Plugin\\') || !str_ends_with($class, 'Mutation')) {
            throw new \InvalidArgumentException('Mutation must be a \\Df\\Plugin\\*Mutation message.');
        }
        $short = ($pos = strrpos($class, '\\')) !== false ? substr($class, $pos + 1) : $class;
        $base = substr($short, 0, -strlen('Mutation'));
        if ($base === '') {
            throw new \InvalidArgumentException('Invalid mutation class name: ' . $class);
        }
        $setter = 'set' . $base;
        $this->respond(function (EventResult $r) use ($setter, $mutation): void {
            if (!method_exists($r, $setter)) {
                throw new \InvalidArgumentException('EventResult does not support mutation via ' . $setter);
            }
            $r->{$setter}($mutation);
        });
    }

    /**
     * Register a callback to be invoked when an ActionResult with the given correlation ID is received.
     * The callback receives (\Df\Plugin\ActionResult $result).
     */
    public function onActionResult(string $correlationId, callable $handler): void {
        $this->sender->onActionResult($correlationId, $handler);
    }

    public function player(string $uuid, string $name = ''): Player {
        return new Player($uuid, $name, $this->getActions());
    }

    public function commandSender(string $uuid): ?Player {
        return $this->server->getPlayer($uuid);
    }

    /**
     * Get the Server instance for accessing online players.
     */
    public function getServer(): Server {
        return $this->server;
    }

    /**
     * Get the player associated with this event.
     * Returns null if the event doesn't have a player (e.g., world events).
     */
    public function getPlayer(): ?Player {
        if ($this->player !== null) {
            return $this->player;
        }
        if ($this->payload === null || !method_exists($this->payload, 'getPlayerUuid')) {
            return null;
        }
        $uuid = $this->payload->getPlayerUuid();
        $name = method_exists($this->payload, 'getName') ? $this->payload->getName() : '';
        $this->player = new Player($uuid, $name, $this->getActions());
        return $this->player;
    }
}
