<?php

namespace Dragonfly\PluginLib\Events;

use Df\Plugin\EventResult;
use Df\Plugin\EventType;
use Dragonfly\PluginLib\Actions\Actions;
use Dragonfly\PluginLib\Actions\ActionsTrait;
use Dragonfly\PluginLib\StreamSender;

final class EventContext {
    use MutationsTrait;
    use ActionsTrait;

    private bool $handled = false;
    private ?Actions $actions = null;

    public function __construct(
        private string $pluginId,
        private string $eventId,
        private StreamSender $sender,
        private bool $expectsResponse,
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
}
