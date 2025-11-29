<?php

namespace Dragonfly\PluginLib;

use Df\Plugin\Action;
use Df\Plugin\ActionBatch;
use Df\Plugin\ActionResult;
use Df\Plugin\EventResult;
use Df\Plugin\PluginToHost;

class StreamSender {
    private $call;
    private string $pluginId;

    /** @var array<string, array<int, callable>> correlationId => list of handlers */
    private array $actionResultHandlers = [];

    /** @var Action[] */
    private array $pendingActions = [];
    private ?int $nextFlushAtNs = null;
    private int $maxBatchSize = 512;

    const FLUSH_WINDOW_NS = 1_000_000; // 1ms

    public function __construct($call, string $pluginId) {
        $this->call = $call;
        $this->pluginId = $pluginId;
    }

    public function enqueue(PluginToHost $message): void {
        $this->flushIfDue();
        $this->call->write($message);
    }

    /**
     * Queue an action to be batched and flushed within ~1ms.
     */
    public function queueAction(Action $action): void {
        // If a previous window has expired, flush it before starting a new one.
        $this->flushIfDue();
        $this->pendingActions[] = $action;
        if ($this->nextFlushAtNs === null) {
            $this->nextFlushAtNs = hrtime(true) + self::FLUSH_WINDOW_NS;
        }
        if (count($this->pendingActions) >= $this->maxBatchSize) {
            // Flush immediately when batch is large to keep latency bounded.
            $this->flushPendingActions();
        }
    }

    /**
     * Flush pending actions immediately if the flush window has elapsed.
     */
    private function flushIfDue(): void {
        if ($this->nextFlushAtNs !== null && hrtime(true) >= $this->nextFlushAtNs) {
            $this->flushPendingActions();
        }
    }

    /**
     * Public tick hook to opportunistically flush if due.
     */
    public function tick(): void {
        $this->flushIfDue();
    }

    /**
     * Flush pending actions immediately.
     */
    public function flushPendingActions(): void {
        if (empty($this->pendingActions)) {
            $this->nextFlushAtNs = null;
            return;
        }
        $batch = new ActionBatch();
        $batch->setActions($this->pendingActions);
        $this->pendingActions = [];
        $this->nextFlushAtNs = null;

        $resp = new PluginToHost();
        $resp->setPluginId($this->pluginId);
        $resp->setActions($batch);
        $this->enqueue($resp);
    }

    public function sendEventResult(string $pluginId, EventResult $result): void {
        $this->flushPendingActions();
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
