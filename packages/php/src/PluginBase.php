<?php

namespace Dragonfly\PluginLib;

use Df\Plugin\EventEnvelope;
use Df\Plugin\CommandSpec;
use Df\Plugin\EventSubscribe;
use Df\Plugin\EventType;
use Df\Plugin\PluginClient;
use Df\Plugin\PluginHello;
use Df\Plugin\PluginToHost;
use Df\Plugin\CustomItemDefinition;
use Df\Plugin\ParamSpec as PbParamSpec;
use Df\Plugin\ParamType as PbParamType;
use Dragonfly\PluginLib\Actions\Actions;
use Dragonfly\PluginLib\Commands\Command;
use Dragonfly\PluginLib\Events\EventContext;
use Dragonfly\PluginLib\Events\Listener;
use Dragonfly\PluginLib\Server\Server;
use ReflectionClass;
use ReflectionMethod;
use ReflectionNamedType;

abstract class PluginBase {
    protected string $pluginId;
    protected string $serverAddress;

    protected string $name = 'example-php';
    protected string $version = '0.1.0';
    protected string $apiVersion = 'v1';

    /** @var array<int, array<int, callable>> */
    private array $handlers = [];

    /** @var int[] */
    private array $subscriptions = [];

    /** @var PluginClient */
    private PluginClient $client;

    /** @var mixed */
    private $call;

    private StreamSender $sender;

    private Server $server;

    private bool $running = false;

    /** @var array<int, array{name: string, description: string, aliases?: string[]        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    private function getDefaultAddress(): string {
        if (PHP_OS_FAMILY === 'Windows') {
            return 'unix://C:/temp/dragonfly_plugin.sock';
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    private function normalizeServerAddress(string $address): string {
        // Handle bare Unix socket paths: "/tmp/dragonfly_plugin.sock" -> "unix:/tmp/dragonfly_plugin.sock"
        if ($address !== '' && $address[0] === '/') {
            return 'unix:' . $address;
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    // Lifecycle hooks
    public function onLoad(): void {        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
    public function onDisable(): void {        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    // Registration APIs
    public function subscribe(array $eventTypes): void {
        $this->subscriptions = array_values(array_unique($eventTypes));
            }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
        $this->handlers[$eventType][] = $handler;
            }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    } else {
                throw new \InvalidArgumentException('Handler map keys must be int EventType values.');
                    }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
            }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
            }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    /**
     * Register a Command class. Automatically wires command event handling and
     * includes aliases in the handshake.
     */
    public function registerCommandClass(Command $cmd): void {
        $name = $cmd->getName();
        if ($name === '') {
            throw new \InvalidArgumentException('Command name must not be empty.');
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
        $this->commandSpecs[] = $spec;
        // Ensure we are subscribed to command events.
        $this->ensureCommandHandler();
            }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    /**
     * Helper to register a custom item from a PNG file path.
     *
     * @param string $id           Identifier like "example:example_item"
     * @param string $displayName  Display name shown to players
     * @param string $pngPath      Absolute or relative path to PNG file
     * @param int    $category     One of \Df\Plugin\ItemCategory::* constants
     * @param string|null $group   Optional subgroup (e.g. "sword", "food")
     * @param int    $meta         Metadata value (default 0)
     */
    public function registerCustomItemFromFile(string $id, string $displayName, string $pngPath, int $category, ?string $group = null, int $meta = 0): void {
        $data = @file_get_contents($pngPath);
        if ($data === false) {
            throw new \RuntimeException("Failed to read PNG file: {$pngPath        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
        $def = new CustomItemDefinition();
        $def->setId($id);
        $def->setDisplayName($displayName);
        $def->setTextureData($data);
        $def->setCategory($category);
        if ($group !== null && $group !== '') {
            $def->setGroup($group);
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    /**
     * Register a listener object.
     * Public, non-static methods with:
     *  - first parameter typed to a payload class under \Df\Plugin\... ending with "Event"
     *  - optional second parameter typed to EventContext
     * are auto-registered. Method names are arbitrary.
     *
     * The handler is invoked as either:
     *  - (TypedPayload $payload)
     *  - (TypedPayload $payload, EventContext $ctx)
     *
     * Use $ctx->getPlayer() to get the Player wrapper for events that have a player.
     * The context auto-ACKs if the handler returns without respond/cancel.
     */
    public function registerListener(object $listener): void {
        if (!$listener instanceof Listener) {
            throw new \InvalidArgumentException('Listener must implement ' . Listener::class);
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
            $params = $method->getParameters();
            if (count($params) < 1) {
                continue;
                    }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
            $paramClass = $type->getName();
            $binding = $this->resolveEventBinding($paramClass);
            if ($binding === null) {
                continue;
                    }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }();
                $ctx = new EventContext($this->pluginId, $eventId, $this->sender, $this->server, $event->getExpectsResponse(), $payload);
                try {
                    if ($wantsContext) {
                        $listener->{$methodName        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    } else {
                        $listener->{$methodName        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }\n");
                        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                    }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
            }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }|null
     */
    private function resolveEventBinding(string $payloadFqcn): ?array {
        if (!str_starts_with($payloadFqcn, 'Df\\Plugin\\') || !str_ends_with($payloadFqcn, 'Event')) {
            return null;
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
        $getter = 'get' . $base;
        $constName = strtoupper(preg_replace('/(?<!^)[A-Z]/', '_$0', $base));
        $constFq = 'Df\\Plugin\\EventType::' . $constName;
        if (!defined($constFq)) {
            return null;
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    // Action helpers moved to StreamSender and HandlerContext

    // Runner
    public function run(): void {
        if (!\extension_loaded('grpc')) {
            fwrite(STDERR, "[php] gRPC extension (ext-grpc) not loaded. Install via 'pecl install grpc' or run with the bundled PHP binary that includes gRPC.\n");
            throw new \RuntimeException('Missing required PHP extension: ext-grpc');
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }...\n");

        $credClass = '\\Grpc\\ChannelCredentials';
        $options = [];
        if (\class_exists($credClass)) {
            /** @var callable $factory */
            $factory = [$credClass, 'createInsecure'];
            $options['credentials'] = \call_user_func($factory);
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

        // Handshake
        fwrite(STDOUT, "[php] connected, sending handshake\n");
        $hello = new PluginToHost();
        $hello->setPluginId($this->pluginId);
        $pluginHello = new PluginHello();
        $pluginHello->setName($this->name);
        $pluginHello->setVersion($this->version);
        $pluginHello->setApiVersion($this->apiVersion);
        if (!empty($this->commandSpecs)) {
            $cmds = [];
            foreach ($this->commandSpecs as $spec) {
                $c = new CommandSpec();
                $c->setName($spec['name']);
                $c->setDescription($spec['description']);
                if (isset($spec['aliases']) && is_array($spec['aliases']) && !empty($spec['aliases'])) {
                    $c->setAliases($spec['aliases']);
                        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    });
                        if (!empty($p['optional'])) {
                            $pp->setOptional(true);
                                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                        $pbParams[] = $pp;
                            }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
            $pluginHello->setCommands($cmds);
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
        $hello->setHello($pluginHello);
        $this->sender->enqueue($hello);

        // Subscribe
        $subscribeMsg = new PluginToHost();
        $subscribeMsg->setPluginId($this->pluginId);
        $subscribe = new EventSubscribe();
        $subscribe->setEvents($this->subscriptions);
        $subscribeMsg->setSubscribe($subscribe);
        $this->sender->enqueue($subscribeMsg);

        try {
            while ($this->running) {
                $message = $this->call->read();
                if ($message === null) {
                    $status = $this->call->getStatus();
                    fwrite(STDOUT, "[php] stream closed - status: code=" . $status->code . " details=" . $status->details . "\n");
                    $this->running = false;
                    break;
                        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }, plugin=" . $this->apiVersion . ")\n");
                            }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

                                if (->hasEvent()) {
                    ->handleEvent(->getEvent());
                    continue;
                        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                    continue;
                        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

                if ($message->hasShutdown()) {
                    fwrite(STDOUT, "[php] shutdown received\n");
                    $this->running = false;
                    continue;
                        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    } catch (\Throwable $e) {
                fwrite(STDERR, "[php] onDisable error: {$e->getMessage()        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
            $this->call->writesDone();
            fwrite(STDOUT, "[php] client completed\n");
            fwrite(STDOUT, "[php] connection closing\n");
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    /**
     * Run onEnable() at most once per server boot, identified by HostHello.boot_id.
     * Persists the last seen boot ID in a temp file keyed by plugin ID.
     */
    private function maybeRunOnEnableOnce(string $bootId): void {
        if ($this->enabledOnce) {
            return;
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    } catch (\Throwable $e) {
                fwrite(STDERR, "[php] onEnable error: {$e->getMessage()        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
            if ($bootId !== '') {
                @file_put_contents($path, $bootId);
                    }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    } else {
            fwrite(STDOUT, "[php] skipping onEnable (already enabled for this server boot)\n");
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    private function bootIdCachePath(): string {
        $safeId = preg_replace('/[^a-zA-Z0-9._-]+/', '_', $this->pluginId);
        return rtrim(sys_get_temp_dir(), DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR . "df_boot_{$safeId        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

    /**
     * Ensure a command event handler is registered once to parse and execute
     * registered Command classes.
     */
    private function ensureCommandHandler(): void {
        if ($this->commandHandlerRegistered) {
            return;
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
            $commandName = $cmdEvt->getCommand();
            if ($commandName === '' || !isset($this->commandInstances[$commandName])) {
                return;
                    }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }

            try {
                $argsField = $cmdEvt->getArgs();
                // Convert protobuf RepeatedField to a native array.
                if (is_array($argsField)) {
                    $args = $argsField;
                        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    } else {
                    $args = [];
                        }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                $cmd->execute($sender, $ctx);
                // Ensure base command execution is suppressed server-side.
                $ctx->cancel();
                    }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    } finally {
                $ctx->ackIfUnhandled();
                    }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    });
            }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }
                    }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    });
            }
    }

    private function handleEvent(EventEnvelope ): void {
         = ->getEventId();
         = ->getType();

        if (isset(->handlers[])) {
            foreach (->handlers[] as ) {
                try {
                    (, );
                } catch (\Throwable ) {
                    fwrite(STDERR, "[php] handler error: {->getMessage()}\n");
                }
            }
            return;
        }

        // Default ack when unhandled
        (new EventContext(->pluginId, , ->sender, ->server, ->getExpectsResponse()))->ackIfUnhandled();
    }


