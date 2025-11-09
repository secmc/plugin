<?php
// Example Dragonfly plugin implemented in PHP.
// Requires: pecl install grpc protobuf

use Grpc\ChannelCredentials;

define('PROTO_PATH', __DIR__ . '/../../../plugin/proto/types/plugin.proto');
require_once __DIR__ . '/vendor/autoload.php';

$pluginId = getenv('DF_PLUGIN_ID') ?: 'php-plugin';
$address = getenv('DF_PLUGIN_GRPC_ADDRESS') ?: '127.0.0.1:50052';

/**
 * IMPORTANT: All events MUST receive an eventResult response to avoid timeout warnings.
 * Even if your plugin doesn't modify or cancel an event, send an acknowledgment with cancel: false.
 */

$client = new \Df\Plugin\PluginClient($address, [
    'credentials' => ChannelCredentials::createInsecure(),
]);

$stream = $client->EventStream();

try {
    foreach ($stream->responses() as $message) {
        if ($message->hasHello()) {
            $hello = new \DF\Plugin\PluginToHost();
            $hello->setPluginId($pluginId);
            $pluginHello = new \DF\Plugin\PluginHello();
            $pluginHello->setName('example-php');
            $pluginHello->setVersion('0.1.0');
            $pluginHello->setApiVersion($message->getHello()->getApiVersion());
            $command = new \DF\Plugin\CommandSpec();
            $command->setName('/cheers');
            $command->setDescription('Send a toast from PHP');
            $pluginHello->setCommands([$command]);
            $hello->setHello($pluginHello);
            $stream->write($hello);

            $sub = new \DF\Plugin\PluginToHost();
            $sub->setPluginId($pluginId);
            $subscribe = new \DF\Plugin\EventSubscribe();
            $subscribe->setEvents(['PLAYER_JOIN', 'COMMAND', 'CHAT']);
            $sub->setSubscribe($subscribe);
            $stream->write($sub);
            continue;
        }

        if ($message->hasEvent()) {
            $event = $message->getEvent();
            
            // Handle PLAYER_JOIN events
            if ($event->getType() === 'PLAYER_JOIN' && $event->hasPlayerJoin()) {
                // Acknowledge the event
                $result = new \DF\Plugin\EventResult();
                $result->setEventId($event->getEventId());
                $result->setCancel(false);
                $resp = new \DF\Plugin\PluginToHost();
                $resp->setPluginId($pluginId);
                $resp->setEventResult($result);
                $stream->write($resp);
                continue;
            }
            
            // Handle CHAT events
            if ($event->getType() === 'CHAT' && $event->hasChat()) {
                $chat = $event->getChat();
                if (stripos($chat->getMessage(), 'spoiler') !== false) {
                    $result = new \DF\Plugin\EventResult();
                    $result->setEventId($event->getEventId());
                    $result->setCancel(true);
                    $resp = new \DF\Plugin\PluginToHost();
                    $resp->setPluginId($pluginId);
                    $resp->setEventResult($result);
                    $stream->write($resp);
                    continue;
                }

                if (str_starts_with($chat->getMessage(), '!cheer ')) {
                    $mutation = new \DF\Plugin\ChatMutation();
                    $mutation->setMessage('🥂 ' . substr($chat->getMessage(), 7));
                    $result = new \DF\Plugin\EventResult();
                    $result->setEventId($event->getEventId());
                    $result->setChat($mutation);
                    $resp = new \DF\Plugin\PluginToHost();
                    $resp->setPluginId($pluginId);
                    $resp->setEventResult($result);
                    $stream->write($resp);
                    continue;
                }
                
                // Acknowledge regular chat messages
                $result = new \DF\Plugin\EventResult();
                $result->setEventId($event->getEventId());
                $result->setCancel(false);
                $resp = new \DF\Plugin\PluginToHost();
                $resp->setPluginId($pluginId);
                $resp->setEventResult($result);
                $stream->write($resp);
                continue;
            }
            
            // Handle COMMAND events
            if ($event->getType() === 'COMMAND' && $event->hasCommand()) {
                if ($event->getCommand()->getRaw() === '/cheers') {
                    $action = new \DF\Plugin\Action();
                    $send = new \DF\Plugin\SendChatAction();
                    $send->setTargetUuid($event->getCommand()->getPlayerUuid());
                    $send->setMessage('🍻 Cheers from the PHP plugin!');
                    $action->setSendChat($send);
                    $batch = new \DF\Plugin\ActionBatch();
                    $batch->setActions([$action]);
                    $resp = new \DF\Plugin\PluginToHost();
                    $resp->setPluginId($pluginId);
                    $resp->setActions($batch);
                    $stream->write($resp);
                }
                
                // Always acknowledge command events
                $result = new \DF\Plugin\EventResult();
                $result->setEventId($event->getEventId());
                $result->setCancel(false);
                $resp = new \DF\Plugin\PluginToHost();
                $resp->setPluginId($pluginId);
                $resp->setEventResult($result);
                $stream->write($resp);
                continue;
            }
        }

        if ($message->hasShutdown()) {
            break;
        }
    }
} catch (Exception $e) {
    echo "[php] Error: " . $e->getMessage() . "\n";
} finally {
    $stream->writesDone();
}

print "[php] plugin connected to {$address}\n";
