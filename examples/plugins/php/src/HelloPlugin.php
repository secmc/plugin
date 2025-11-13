<?php
namespace ExamplePhp;
// Example Dragonfly plugin implemented in PHP (client mode).
// Requires: pecl install grpc protobuf
//
// ✅ This works with the standard PECL gRPC extension.
// The plugin connects to the Dragonfly server as a gRPC client and exchanges
// bidirectional messages over the EventStream RPC.

require_once __DIR__ . '/../vendor/autoload.php';

use Df\Plugin\EventType;
use Df\Plugin\PlayerJoinEvent;
use Df\Plugin\ItemStack;
use Df\Plugin\ItemCategory;
use Df\Plugin\ChatEvent;
use Df\Plugin\CommandEvent;
use Df\Plugin\PlayerAttackEntityEvent;
use Df\Plugin\PlayerAttackEntityMutation;
use Dragonfly\PluginLib\PluginBase;
use Dragonfly\PluginLib\Events\EventContext;
use Dragonfly\PluginLib\Events\Listener;
use ExamplePhp\EffectCommand;

class HelloPlugin extends PluginBase implements Listener {

    protected string $name = 'example-php';
    protected string $version = '0.1.0';

    public function onEnable(): void {
        $this->registerCommandClass(new EffectCommand());
        $this->registerCommand('/cheers', 'Send a toast from PHP');
        $this->registerCommand('/pokemon', 'Give a Pokemon item');
        // Register custom items
        $this->registerCustomItemFromFile(
            'vasar:pokemon',
            'Pokemon Item',
            __DIR__ . '/../assets/daco.png',
            ItemCategory::ITEM_CATEGORY_ITEMS,
            null,
            0
        );
        $this->registerListener($this);
    }

    public function onPlayerJoin(PlayerJoinEvent $e, EventContext $ctx): void {
        $stack = new ItemStack();
        $stack->setName('vasar:pokemon');
        $stack->setMeta(0);
        $stack->setCount(1);
        $ctx->giveItemUuid($e->getPlayerUuid(), $stack);
    }

    public function onChat(ChatEvent $chat, EventContext $ctx): void {
        $text = $chat->getMessage();

        if (stripos($text, 'spoiler') !== false) {
            $ctx->cancel();
            return;
        }

        if (str_starts_with($text, '!cheer ')) {
            $ctx->chat('🥂 ' . substr($text, 7));
            return;
        }
    }

    public function onCommand(CommandEvent $command, EventContext $ctx): void {
        switch ($command->getRaw()) {
            case '/cheers':
                $ctx->chatToUuid($command->getPlayerUuid(), 'Cheers from the PHP plugin!');
                break;
            case '/pokemon':
                $ctx->chatToUuid($command->getPlayerUuid(), 'You have been given a Pokemon item!');
                $stack = new ItemStack();
                $stack->setName('vasar:pokemon');
                $stack->setMeta(0);
                $stack->setCount(1);
                $ctx->giveItemUuid($command->getPlayerUuid(), $stack);
                break;
        }
    }

    public function onPlayerAttackEntity(PlayerAttackEntityEvent $e, EventContext $ctx): void {
        $mutation = new PlayerAttackEntityMutation();
        $mutation->setForce(0.6);
        $mutation->setHeight(0.6);
        $ctx->playerAttackEntity($mutation);
    }
}

$plugin = new HelloPlugin();
$plugin->run();
