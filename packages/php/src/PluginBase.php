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

    /** @var array<int, array{name: string, description: string, aliases?: string[]}> */
    private array $commandSpecs = [];

    /** @var CustomItemDefinition[] */
    private array $customItems = [];

    /** @var array<string, Command> name/alias => command instance */
    private array $commandInstances = [];
    private bool $commandHandlerRegistered = false;
    private bool $enabledOnce = false;

    public function __construct(?string $pluginId = null, ?string $serverAddress = null) {
        $this->pluginId = $pluginId ?? (getenv('DF_PLUGIN_ID') ?: 'php-plugin');
        $address = $serverAddress ?? (getenv('DF_PLUGIN_SERVER_ADDRESS') ?: $this->getDefaultAddress());
        $this->serverAddress = $this->normalizeServerAddress($address);
    }

    private function getDefaultAddress(): string {
        if (PHP_OS_FAMILY === 'Windows') {
            return 'unix://C:/temp/dragonfly_plugin.sock';
        }
        // PHP gRPC extension format for Unix sockets
        return 'unix:/tmp/dragonfly_plugin.sock';
    }

    private function normalizeServerAddress(string $address): string {
        // Handle bare Unix socket paths: "/tmp/dragonfly_plugin.sock" -> "unix:/tmp/dragonfly_plugin.sock"
        if ($address !== '' && $address[0] === '/') {
            return 'unix:' . $address;
        }
        // Normalize triple-slash form to single-slash: "unix:///path" -> "unix:/path"
        $normalized = preg_replace('#^unix:///#', 'unix:/', $address);
        return $normalized ?? $address;
    }

    // Lifecycle hooks
    public function onLoad(): void {}
    public function onEnable(): void {}
    public function onDisable(): void {}

    /**
     * Get the Server instance for accessing online players.
     */
    public function getServer(): Server {
        return $this->server;
    }

    // Registration APIs
    public function subscribe(array $eventTypes): void {
        $this->subscriptions = array_values(array_unique($eventTypes));
    }

    public function addEventHandler(int $eventType, callable $handler): void {
        if (!isset($this->handlers[$eventType])) {
            $this->handlers[$eventType] = [];
        }
        $this->handlers[$eventType][] = $handler;
    }

    /**
     * Register many handlers at once.
     * Keys must be int EventType values (e.g. EventType::PLAYER_JOIN).
     *
     * Handlers receive (string $eventId, EventEnvelope $event).
     */
    public function registerHandlers(array $map): void {
        foreach ($map as $key => $handler) {
            if (is_int($key)) {
                $this->addEventHandler($key, $handler);
            } else {
                throw new \InvalidArgumentException('Handler map keys must be int EventType values.');
            }
        }
    }

    /**
     * Subscribe to the set of types that have handlers registered.
     */
    public function subscribeToRegisteredHandlers(): void {
        $types = [];
        foreach ($this->handlers as $type => $_) {
            if (is_int($type)) {
                $types[] = $type;
            }
        }
        if (!empty($types)) {
            $this->subscriptions = array_values(array_unique($types));
        }
    }

    public function registerCommand(string $name, string $description): void {
        $this->commandSpecs[] = ['name' => $name, 'description' => $description];
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
        // Store instance by name and aliases for quick lookup.
        $this->commandInstances[$name] = $cmd;
        foreach ($cmd->getAliases() as $alias) {
            if ($alias !== '' && !isset($this->commandInstances[$alias])) {
                $this->commandInstances[$alias] = $cmd;
            }
        }
        // Queue spec for handshake (with aliases).
        $spec = [
            'name' => $name,
            'description' => $cmd->getDescription(),
        ];
        $aliases = $cmd->getAliases();
        if (!empty($aliases)) {
            $spec['aliases'] = array_values(array_unique($aliases));
        }
        $this->commandSpecs[] = $spec;
        // Ensure we are subscribed to command events.
        $this->ensureCommandHandler();
    }

    /**
     * Queue a custom item definition to be sent in PluginHello.
     */
    public function registerCustomItem(CustomItemDefinition $def): void {
        $this->customItems[] = $def;
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
            throw new \RuntimeException("Failed to read PNG file: {$pngPath}");
        }
        $def = new CustomItemDefinition();
        $def->setId($id);
        $def->setDisplayName($displayName);
        $def->setTextureData($data);
        $def->setCategory($category);
        if ($group !== null && $group !== '') {
            $def->setGroup($group);
        }
        $def->setMeta($meta);
        $this->registerCustomItem($def);
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

        $ref = new ReflectionClass($listener);
        foreach ($ref->getMethods(ReflectionMethod::IS_PUBLIC) as $method) {
            if ($method->isStatic() || $method->isConstructor() || $method->isDestructor()) {
                continue;
            }
            $params = $method->getParameters();
            if (count($params) < 1) {
                continue;
            }
            $param = $params[0];
            $type = $param->getType();
            if (!$type instanceof ReflectionNamedType || $type->isBuiltin()) {
                continue;
            }
            $paramClass = $type->getName();
            $binding = $this->resolveEventBinding($paramClass);
            if ($binding === null) {
                continue;
            }

            $eventType = $binding['type'];
            $getter = $binding['getter'];
            $methodName = $method->getName();

            $wantsContext = $method->getNumberOfParameters() >= 2;
            $this->addEventHandler($eventType, function (string $eventId, EventEnvelope $event) use ($listener, $methodName, $getter, $wantsContext): void {
                $payload = $event->{$getter}();
                $ctx = new EventContext($this->pluginId, $eventId, $this->sender, $this->server, $event->getExpectsResponse(), $payload);
                try {
                    if ($wantsContext) {
                        $listener->{$methodName}($payload, $ctx);
                    } else {
                        $listener->{$methodName}($payload);
                    }
                } catch (\Throwable $e) {
                    fwrite(STDERR, "[php] listener error: {$e->getMessage()}\n");
                } finally {
                    // Avoid double-ACK for command events when a Command class handles the same command.
                    // If a registered Command instance exists for this command name, the internal command
                    // handler will already respond/cancel. In that case, skip the auto-ACK here.
                    $shouldAck = true;
                    if ($event->getType() === EventType::COMMAND && $payload !== null && method_exists($payload, 'getCommand')) {
                        $cmdName = (string)$payload->getCommand();
                        if ($cmdName !== '' && isset($this->commandInstances[$cmdName])) {
                            $shouldAck = false;
                        }
                    }
                    if ($shouldAck) {
                        $ctx->ackIfUnhandled();
                    }
                }
            });
        }
    }

    /**
     * Resolve event type constant and Event getter name from a payload FQCN.
     * Example: \Df\Plugin\PlayerJoinEvent -> ['type' => EventType::PLAYER_JOIN, 'getter' => 'getPlayerJoin']
     *
     * @return array{type:int,getter:string}|null
     */
    private function resolveEventBinding(string $payloadFqcn): ?array {
        if (!str_starts_with($payloadFqcn, 'Df\\Plugin\\') || !str_ends_with($payloadFqcn, 'Event')) {
            return null;
        }
        $short = ($pos = strrpos($payloadFqcn, '\\')) !== false ? substr($payloadFqcn, $pos + 1) : $payloadFqcn;
        $base = substr($short, 0, -strlen('Event'));
        if ($base === '') {
            return null;
        }
        $getter = 'get' . $base;
        $constName = strtoupper(preg_replace('/(?<!^)[A-Z]/', '_$0', $base));
        $constFq = 'Df\\Plugin\\EventType::' . $constName;
        if (!defined($constFq)) {
            return null;
        }
        /** @var int $type */
        $type = constant($constFq);
        return ['type' => $type, 'getter' => $getter];
    }

    // Runner
    public function run(): void {
        if (!\extension_loaded('grpc')) {
            fwrite(STDERR, "[php] gRPC extension (ext-grpc) not loaded. Install via 'pecl install grpc' or run with the bundled PHP binary that includes gRPC.\n");
            throw new \RuntimeException('Missing required PHP extension: ext-grpc');
        }
        fwrite(STDOUT, "[php] connecting to {$this->serverAddress}...\n");

        $credClass = '\\Grpc\\ChannelCredentials';
        $options = [];
        if (\class_exists($credClass)) {
            /** @var callable $factory */
            $factory = [$credClass, 'createInsecure'];
            $options['credentials'] = \call_user_func($factory);
        }
        $this->client = new PluginClient($this->serverAddress, $options);
        $this->call = $this->client->EventStream();
        $this->sender = new StreamSender($this->call, $this->pluginId);
        $this->server = new Server(new Actions($this->sender, $this->pluginId));
        $this->running = true;

        // Register internal handlers to track online players
        $this->registerPlayerTracking();

        // Lifecycle
        // onLoad: runs on every plugin process start/reload.
        $this->onLoad();
        // onEnable: run once per server boot BEFORE handshake so that one-time
        // resources (e.g., custom items) are present in PluginHello on first boot.
        $bootIdEnv = getenv('DF_HOST_BOOT_ID') ?: '';
        $this->maybeRunOnEnableOnce($bootIdEnv);

        // Defaults if not set
        if (empty($this->subscriptions)) {
            // Prefer subscriptions matching registered handlers if present.
            $this->subscribeToRegisteredHandlers();
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
                // If protobuf has params field, populate it from the registered command class.
                if (method_exists($c, 'setParams') && isset($this->commandInstances[$spec['name']])) {
                    $cmd = $this->commandInstances[$spec['name']];
                    $schema = $cmd->serializeParamSpec();
                    $pbParams = [];
                    foreach ($schema as $p) {
                        $pp = new PbParamSpec();
                        $pp->setName($p['name']);
                        // Map string type to enum.
                        $type = $p['type'] ?? 'string';
                        $pp->setType(match ($type) {
                            'int' => PbParamType::PARAM_INT,
                            'float' => PbParamType::PARAM_FLOAT,
                            'bool' => PbParamType::PARAM_BOOL,
                            'enum' => PbParamType::PARAM_ENUM,
                            'varargs' => PbParamType::PARAM_VARARGS,
                            'target' => PbParamType::PARAM_TARGET,
                            'targets' => PbParamType::PARAM_TARGETS,
                            default => PbParamType::PARAM_STRING,
                        });
                        if (!empty($p['optional'])) {
                            $pp->setOptional(true);
                        }
                        if (!empty($p['enum_values']) && method_exists($pp, 'setEnumValues')) {
                            $pp->setEnumValues($p['enum_values']);
                        }
                        $pbParams[] = $pp;
                    }
                    if (!empty($pbParams)) {
                        $c->setParams($pbParams);
                    }
                }
                $cmds[] = $c;
            }
            $pluginHello->setCommands($cmds);
        }
        if (!empty($this->customItems)) {
            $pluginHello->setCustomItems($this->customItems);
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

                if ($message->hasHello()) {
                    $hostHello = $message->getHello();
                    fwrite(STDOUT, "[php] host hello api=" . $hostHello->getApiVersion() . "\n");
                    if ($hostHello->getApiVersion() !== $this->apiVersion) {
                        fwrite(STDOUT, "[php] WARNING: API version mismatch (host={$hostHello->getApiVersion()}, plugin=" . $this->apiVersion . ")\n");
                    }
                } elseif ($message->hasEvent()) {
                    $event = $message->getEvent();
                    $eventId = $event->getEventId();
                    $type = $event->getType();

                    if (isset($this->handlers[$type])) {
                        foreach ($this->handlers[$type] as $handler) {
                            try {
                                $handler($eventId, $event);
                            } catch (\Throwable $e) {
                                fwrite(STDERR, "[php] handler error: {$e->getMessage()}\n");
                            }
                        }
                    } else {
                        // Default ack when unhandled
                        (new EventContext($this->pluginId, $eventId, $this->sender, $this->server, $event->getExpectsResponse()))->ackIfUnhandled();
                    }
                } elseif ($message->hasActionResult()) {
                    $result = $message->getActionResult();
                    $this->sender->dispatchActionResult($result);
                } elseif ($message->hasShutdown()) {
                    fwrite(STDOUT, "[php] shutdown received\n");
                    $this->running = false;
                }
                $this->sender->tick();
            }
        } finally {
            try {
                $this->onDisable();
            } catch (\Throwable $e) {
                fwrite(STDERR, "[php] onDisable error: {$e->getMessage()}\n");
            }
            // Ensure any buffered actions are sent before closing the write side.
            $this->sender->flushPendingActions();
            $this->call->writesDone();
            fwrite(STDOUT, "[php] client completed\n");
            fwrite(STDOUT, "[php] connection closing\n");
        }
    }

    /**
     * Run onEnable() at most once per server boot, identified by HostHello.boot_id.
     * Persists the last seen boot ID in a temp file keyed by plugin ID.
     */
    private function maybeRunOnEnableOnce(string $bootId): void {
        if ($this->enabledOnce) {
            return;
        }
        $path = $this->bootIdCachePath();
        $last = @file_exists($path) ? @trim((string)@file_get_contents($path)) : '';
        if ($bootId === '' || $bootId !== $last) {
            // First time for this server boot: run onEnable and store boot ID.
            try {
                $this->onEnable();
            } catch (\Throwable $e) {
                fwrite(STDERR, "[php] onEnable error: {$e->getMessage()}\n");
            }
            if ($bootId !== '') {
                @file_put_contents($path, $bootId);
            }
        } else {
            fwrite(STDOUT, "[php] skipping onEnable (already enabled for this server boot)\n");
        }
        $this->enabledOnce = true;
    }

    private function bootIdCachePath(): string {
        $safeId = preg_replace('/[^a-zA-Z0-9._-]+/', '_', $this->pluginId);
        return rtrim(sys_get_temp_dir(), DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR . "df_boot_{$safeId}.txt";
    }

    /**
     * Ensure a command event handler is registered once to parse and execute
     * registered Command classes.
     */
    private function ensureCommandHandler(): void {
        if ($this->commandHandlerRegistered) {
            return;
        }
        $this->commandHandlerRegistered = true;
        $this->addEventHandler(EventType::COMMAND, function (string $eventId, EventEnvelope $event): void {
            $cmdEvt = $event->getCommand();
            if ($cmdEvt === null) {
                return;
            }
            $commandName = $cmdEvt->getCommand();
            if ($commandName === '' || !isset($this->commandInstances[$commandName])) {
                return;
            }
            // Work with a fresh instance per execution.
            $template = $this->commandInstances[$commandName];
            $cmd = clone $template;

            $senderUuid = $cmdEvt->getPlayerUuid();
            $ctx = new EventContext($this->pluginId, $eventId, $this->sender, $this->server, $event->getExpectsResponse());
            $sender = $ctx->commandSender($senderUuid);
            if ($sender === null) {
                // Can't resolve sender; let server handle normally
                $ctx->ackIfUnhandled();
                return;
            }

            try {
                $argsField = $cmdEvt->getArgs();
                // Convert protobuf RepeatedField to a native array.
                if (is_array($argsField)) {
                    $args = $argsField;
                } elseif ($argsField instanceof \Traversable) {
                    $args = iterator_to_array($argsField);
                } else {
                    $args = [];
                }
                if (!$cmd->parseArgs($args)) {
                    $usage = method_exists($cmd, 'generateUsage') ? $cmd->generateUsage() : ('/' . $commandName);
                    $ctx->chatToUuid($senderUuid, "§cUsage: " . $usage);
                    $ctx->cancel();
                    return;
                }
                $cmd->execute($sender, $ctx);
                // Ensure base command execution is suppressed server-side.
                $ctx->cancel();
            } catch (\Throwable $e) {
                $ctx->chatToUuid($senderUuid, "§cCommand error: " . $e->getMessage());
                // Suppress base command execution even on error to avoid duplicate messages.
                $ctx->cancel();
            } finally {
                $ctx->ackIfUnhandled();
            }
        });
    }

    /**
     * Register internal handlers to track online players via join/quit events.
     */
    private function registerPlayerTracking(): void {
        // Track player joins
        $this->addEventHandler(EventType::PLAYER_JOIN, function (string $eventId, EventEnvelope $event): void {
            $payload = $event->getPlayerJoin();
            if ($payload !== null) {
                $world = method_exists($payload, 'getWorld') ? $payload->getWorld() : null;
                $this->server->addPlayer($payload->getPlayerUuid(), $payload->getName(), $world);
            }
        });

        // Track player quits
        $this->addEventHandler(EventType::PLAYER_QUIT, function (string $eventId, EventEnvelope $event): void {
            $payload = $event->getPlayerQuit();
            if ($payload !== null) {
                $this->server->removePlayer($payload->getPlayerUuid());
            }
        });

        // Track world changes
        $this->addEventHandler(EventType::PLAYER_CHANGE_WORLD, function (string $eventId, EventEnvelope $event): void {
            $payload = $event->getPlayerChangeWorld();
            if ($payload !== null && method_exists($payload, 'getAfter')) {
                $world = $payload->getAfter();
                if ($world !== null) {
                    $this->server->setPlayerWorld($payload->getPlayerUuid(), $world);
                }
            }
        });
    }
}
