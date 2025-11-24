<?php
namespace ExamplePhp;
// Example Dragonfly plugin implemented in PHP (client mode).
// Requires: pecl install grpc protobuf
//
// ✅ This works with the standard PECL gRPC extension.
// The plugin connects to the Dragonfly server as a gRPC client and exchanges
// bidirectional messages over the EventStream RPC.

require_once __DIR__ . '/../vendor/autoload.php';

use Df\Plugin\PlayerJoinEvent;
use Df\Plugin\ItemStack;
use Df\Plugin\ItemCategory;
use Df\Plugin\ChatEvent;
use Df\Plugin\CommandEvent;
use Df\Plugin\PlayerAttackEntityEvent;
use Df\Plugin\PlayerAttackEntityMutation;
use Df\Plugin\PlayerJumpEvent;
use Df\Plugin\Sound;
use Dragonfly\PluginLib\PluginBase;
use Dragonfly\PluginLib\Events\EventContext;
use Dragonfly\PluginLib\Events\Listener;
use ExamplePhp\EffectCommand;
use ExamplePhp\CircleCommand;

class HelloPlugin extends PluginBase implements Listener {

    protected string $name = 'example-php';
    protected string $version = '0.1.0';

    public function onEnable(): void {
        $this->registerCommandClass(new EffectCommand());
        $this->registerCommandClass(new CircleCommand());
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
        $player = $ctx->getPlayer();
        $stack = new ItemStack();
        $stack->setName('vasar:pokemon');
        $stack->setMeta(0);
        $stack->setCount(1);
        $player->giveItem($stack);
    }

    public function onChat(ChatEvent $chat, EventContext $ctx): void {
        $text = $chat->getMessage();

        if (stripos($text, 'spoiler') !== false) {
            $ctx->cancel();
            return;
        }

        if (str_starts_with($text, '!cheer ')) {
            $ctx->getPlayer()->sendMessage('🥂 ' . substr($text, 7));
            return;
        }
    }

    public function onCommand(CommandEvent $command, EventContext $ctx): void {
        $player = $ctx->getPlayer();
        switch ($command->getRaw()) {
            case '/cheers':
                $player->sendMessage('Cheers from the PHP plugin!');
                break;
            case '/pokemon':
                $player->sendMessage('You have been given a Pokemon item!');
                $stack = new ItemStack();
                $stack->setName('vasar:pokemon');
                $stack->setMeta(0);
                $stack->setCount(1);
                $player->giveItem($stack);
                break;
        }
    }

    public function onPlayerAttackEntity(PlayerAttackEntityEvent $e, EventContext $ctx): void {
        $mutation = new PlayerAttackEntityMutation();
        $mutation->setForce(0.6);
        $mutation->setHeight(0.6);
        $ctx->playerAttackEntity($mutation);
    }

    public function onPlayerJump(PlayerJumpEvent $e, EventContext $ctx): void {
        $player = $ctx->getPlayer();
        $position = $e->getPosition();
        $player->playSound(Sound::EXPLOSION, $position, 1.0, 1.0);
    }

}

$plugin = new HelloPlugin();
$plugin->run();
