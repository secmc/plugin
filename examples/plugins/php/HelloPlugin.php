<?php
// Example Dragonfly plugin implemented in PHP.
// Requires: pecl install grpc protobuf

use Grpc\Server;
use Grpc\ServerCredentials;
use Grpc\UnaryCall;

define('PROTO_PATH', __DIR__ . '/../../../plugin/proto/plugin.proto');
require_once __DIR__ . '/vendor/autoload.php';

$pluginId = getenv('DF_PLUGIN_ID') ?: 'php-plugin';
$address = getenv('DF_PLUGIN_GRPC_ADDRESS') ?: '127.0.0.1:50052';

$server = new Server();
$server->addHttp2Port($address, ServerCredentials::createInsecure());
$service = new \DF\Plugin\PluginService();
$service->setEventStreamHandler(function ($stream) use ($pluginId) {
    foreach ($stream->readAll() as $message) {
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
            $subscribe->setEvents(['PLAYER_JOIN', 'COMMAND']);
            $sub->setSubscribe($subscribe);
            $stream->write($sub);
            continue;
        }

        if ($message->hasEvent()) {
            $event = $message->getEvent();
            if ($event->getType() === 'COMMAND' && $event->getCommand()->getRaw() === '/cheers') {
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
        }

        if ($message->hasShutdown()) {
            break;
        }
    }
    $stream->finish();
});

$server->handle($service);
$server->start();
print "[php] plugin listening on {$address}\n";
$server->wait();
