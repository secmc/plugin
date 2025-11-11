<?php
// Example Dragonfly plugin implemented in PHP (client mode).
// Requires: pecl install grpc protobuf
//
// ✅ This works with the standard PECL gRPC extension.
// The plugin connects to the Dragonfly server as a gRPC client and exchanges
// bidirectional messages over the EventStream RPC.

require_once __DIR__ . '/../vendor/autoload.php';

use Df\Plugin\Action;
use Df\Plugin\ActionBatch;
use Df\Plugin\ChatMutation;
use Df\Plugin\CommandSpec;
use Df\Plugin\EventResult;
use Df\Plugin\EventSubscribe;
use Df\Plugin\EventType;
use Df\Plugin\PluginClient;
use Df\Plugin\PluginHello;
use Df\Plugin\PluginToHost;
use Df\Plugin\SendChatAction;
use Grpc\ChannelCredentials;

define('DF_PLUGIN_API_VERSION', 'v1');

$pluginId = getenv('DF_PLUGIN_ID') ?: 'php-plugin';
$serverAddress = getenv('DF_PLUGIN_SERVER_ADDRESS') ?: '127.0.0.1:50050';

fwrite(STDOUT, "[php] connecting to {$serverAddress}...\n");

$client = new PluginClient($serverAddress, [
    'credentials' => ChannelCredentials::createInsecure(),
]);

$call = $client->EventStream();

fwrite(STDOUT, "[php] connected, sending handshake\n");

$hello = new PluginToHost();
$hello->setPluginId($pluginId);
$pluginHello = new PluginHello();
$pluginHello->setName('example-php');
$pluginHello->setVersion('0.1.0');
$pluginHello->setApiVersion(DF_PLUGIN_API_VERSION);
$command = new CommandSpec();
$command->setName('/cheers');
$command->setDescription('Send a toast from PHP');
$pluginHello->setCommands([$command]);
$hello->setHello($pluginHello);
$call->write($hello);

$subscribeMsg = new PluginToHost();
$subscribeMsg->setPluginId($pluginId);
$subscribe = new EventSubscribe();
$subscribe->setEvents([
    EventType::PLAYER_JOIN,
    EventType::COMMAND,
    EventType::CHAT,
]);
$subscribeMsg->setSubscribe($subscribe);
$call->write($subscribeMsg);

try {
    while (true) {
        $message = $call->read();
        if ($message === null) {
            fwrite(STDOUT, "[php] stream ended by host\n");
            break;
        }

        if ($message->hasHello()) {
            $hostHello = $message->getHello();
            fwrite(STDOUT, "[php] host hello api=" . $hostHello->getApiVersion() . "\n");
            if ($hostHello->getApiVersion() !== DF_PLUGIN_API_VERSION) {
                fwrite(STDOUT, "[php] WARNING: API version mismatch (host={$hostHello->getApiVersion()}, plugin=" . DF_PLUGIN_API_VERSION . ")\n");
            }
            continue;
        }

        if ($message->hasEvent()) {
            $event = $message->getEvent();
            $eventId = $event->getEventId();

            if ($event->getType() === EventType::PLAYER_JOIN && $event->hasPlayerJoin()) {
                acknowledgeEvent($call, $pluginId, $eventId);
                continue;
            }

            if ($event->getType() === EventType::CHAT && $event->hasChat()) {
                $chat = $event->getChat();
                $text = $chat->getMessage();

                if (stripos($text, 'spoiler') !== false) {
                    cancelEvent($call, $pluginId, $eventId);
                    continue;
                }

                if (str_starts_with($text, '!cheer ')) {
                    $mutation = new ChatMutation();
                    $mutation->setMessage('🥂 ' . substr($text, 7));
                    mutateChat($call, $pluginId, $eventId, $mutation);
                    continue;
                }

                acknowledgeEvent($call, $pluginId, $eventId);
                continue;
            }

            if ($event->getType() === EventType::COMMAND && $event->hasCommand()) {
                $commandEvent = $event->getCommand();
                if ($commandEvent->getRaw() === '/cheers') {
                    $action = new Action();
                    $send = new SendChatAction();
                    $send->setTargetUuid($commandEvent->getPlayerUuid());
                    $send->setMessage('🍻 Cheers from the PHP plugin!');
                    $action->setSendChat($send);

                    $batch = new ActionBatch();
                    $batch->setActions([$action]);

                    $resp = new PluginToHost();
                    $resp->setPluginId($pluginId);
                    $resp->setActions($batch);
                    $call->write($resp);
                }

                acknowledgeEvent($call, $pluginId, $eventId);
                continue;
            }

            acknowledgeEvent($call, $pluginId, $eventId);
            continue;
        }

        if ($message->hasShutdown()) {
            fwrite(STDOUT, "[php] shutdown received\n");
            break;
        }
    }
} catch (\Throwable $e) {
    fwrite(STDERR, "[php] error: {$e->getMessage()}\n");
    fwrite(STDERR, $e->getTraceAsString() . "\n");
    exit(1);
} finally {
    $call->writesDone();
    fwrite(STDOUT, "[php] connection closed\n");
}

function acknowledgeEvent($call, string $pluginId, string $eventId): void
{
    $result = new EventResult();
    $result->setEventId($eventId);
    $result->setCancel(false);

    $resp = new PluginToHost();
    $resp->setPluginId($pluginId);
    $resp->setEventResult($result);
    $call->write($resp);
}

function cancelEvent($call, string $pluginId, string $eventId): void
{
    $result = new EventResult();
    $result->setEventId($eventId);
    $result->setCancel(true);

    $resp = new PluginToHost();
    $resp->setPluginId($pluginId);
    $resp->setEventResult($result);
    $call->write($resp);
}

function mutateChat($call, string $pluginId, string $eventId, ChatMutation $mutation): void
{
    $result = new EventResult();
    $result->setEventId($eventId);
    $result->setChat($mutation);

    $resp = new PluginToHost();
    $resp->setPluginId($pluginId);
    $resp->setEventResult($result);
    $call->write($resp);
}
