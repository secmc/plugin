<?php

namespace Dragonfly\PluginLib;

use Df\Plugin\ActionResult;
use Df\Plugin\EventResult;
use Df\Plugin\PluginToHost;

class StreamSender {
    private $call;
    /** @var array<string, array<int, callable>> correlationId => list of handlers */
    private array $actionResultHandlers = [];

    public function __construct($call) {
        $this->call = $call;
    }

    public function enqueue(PluginToHost $message): void {
        $this->call->write($message);
    }

    public function sendEventResult(string $pluginId, EventResult $result): void {
        $resp = new PluginToHost();
        $resp->setPluginId($pluginId);
        $resp->setEventResult($result);
        $this->enqueue($resp);
    }

    /**
     * Build and send an EventResult via a mutator closure.
     * The closure receives the EventResult to populate.
     */
    public function respond(string $pluginId, string $eventId, callable $mutator): void {
        $result = new EventResult();
        $result->setEventId($eventId);
        $mutator($result);
        $this->sendEventResult($pluginId, $result);
    }

    /**
     * Register a handler to be invoked when an ActionResult with the given correlation ID arrives.
     * Handlers are one-shot: all handlers for a correlation ID are invoked once and then cleared.
     *
     * @param string   $correlationId
     * @param callable $handler receives (\Df\Plugin\ActionResult $result): void
     */
    public function onActionResult(string $correlationId, callable $handler): void {
        if ($correlationId === '') {
            return;
        }
        if (!isset($this->actionResultHandlers[$correlationId])) {
            $this->actionResultHandlers[$correlationId] = [];
        }
        $this->actionResultHandlers[$correlationId][] = $handler;
    }

    /**
     * Dispatch an incoming ActionResult to any registered handlers.
     * This is called by the main stream loop when results arrive.
     */
    public function dispatchActionResult(ActionResult $result): void {
        $cid = $result->getCorrelationId();
        if ($cid === '' || !isset($this->actionResultHandlers[$cid])) {
            return;
        }
        $handlers = $this->actionResultHandlers[$cid];
        unset($this->actionResultHandlers[$cid]);
        foreach ($handlers as $h) {
            try {
                $h($result);
            } catch (\Throwable $e) {
                fwrite(STDERR, "[php] action result handler error: {$e->getMessage()}\n");
            }
        }
    }

}
