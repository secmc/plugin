<?php

namespace Dragonfly\PluginLib\Actions;

use Df\Plugin\Action;
use Df\Plugin\AddEffectAction;
use Df\Plugin\ClearInventoryAction;
use Df\Plugin\ExecuteCommandAction;
use Df\Plugin\GiveItemAction;
use Df\Plugin\ItemStack;
use Df\Plugin\PlaySoundAction;
use Df\Plugin\RemoveEffectAction;
use Df\Plugin\SendChatAction;
use Df\Plugin\SendPopupAction;
use Df\Plugin\SendTipAction;
use Df\Plugin\SendTitleAction;
use Df\Plugin\SetExperienceAction;
use Df\Plugin\SetFoodAction;
use Df\Plugin\TeleportAction;
use Df\Plugin\KickAction;
use Df\Plugin\SetGameModeAction;
use Df\Plugin\SetHealthAction;
use Df\Plugin\SetHeldItemAction;
use Df\Plugin\SetVelocityAction;
use Df\Plugin\Vec3;
use Df\Plugin\WorldSetDefaultGameModeAction;
use Df\Plugin\WorldSetDifficultyAction;
use Df\Plugin\WorldSetTickRangeAction;
use Df\Plugin\WorldSetBlockAction;
use Df\Plugin\WorldPlaySoundAction;
use Df\Plugin\WorldAddParticleAction;
use Df\Plugin\WorldQueryEntitiesAction;
use Df\Plugin\WorldQueryPlayersAction;
use Df\Plugin\WorldQueryEntitiesWithinAction;
use Df\Plugin\WorldRef;
use Df\Plugin\BlockPos;
use Df\Plugin\BlockState;
use Df\Plugin\BBox;
use Dragonfly\PluginLib\StreamSender;

final class Actions {
    public function __construct(
        private StreamSender $sender,
    ) {}

    public function sendAction(Action $action): void {
        $this->sender->queueAction($action);
    }

    /**
     * Flush any queued actions immediately.
     */
    public function flush(): void {
        $this->sender->flushPendingActions();
    }

    public function chatToUuid(string $targetUuid, string $message): void {
        $action = new Action();
        $send = new SendChatAction();
        $send->setTargetUuid($targetUuid);
        $send->setMessage($message);
        $action->setSendChat($send);
        $this->sendAction($action);
    }

    public function teleportUuid(string $playerUuid, ?Vec3 $position = null, ?Vec3 $rotation = null): void {
        $action = new Action();
        $teleport = new TeleportAction();
        $teleport->setPlayerUuid($playerUuid);
        if ($position !== null) {
            $teleport->setPosition($position);
        }
        if ($rotation !== null) {
            $teleport->setRotation($rotation);
        }
        $action->setTeleport($teleport);
        $this->sendAction($action);
    }

    public function kickUuid(string $playerUuid, string $reason): void {
        $action = new Action();
        $kick = new KickAction();
        $kick->setPlayerUuid($playerUuid);
        $kick->setReason($reason);
        $action->setKick($kick);
        $this->sendAction($action);
    }

    public function setGameModeUuid(string $playerUuid, int $gameMode): void {
        $action = new Action();
        $set = new SetGameModeAction();
        $set->setPlayerUuid($playerUuid);
        $set->setGameMode($gameMode);
        $action->setSetGameMode($set);
        $this->sendAction($action);
    }

    public function giveItemUuid(string $playerUuid, ItemStack $item): void {
        $action = new Action();
        $msg = new GiveItemAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setItem($item);
        $action->setGiveItem($msg);
        $this->sendAction($action);
    }

    public function clearInventoryUuid(string $playerUuid): void {
        $action = new Action();
        $msg = new ClearInventoryAction();
        $msg->setPlayerUuid($playerUuid);
        $action->setClearInventory($msg);
        $this->sendAction($action);
    }

    public function setHeldItemsUuid(string $playerUuid, ?ItemStack $main = null, ?ItemStack $offhand = null): void {
        $action = new Action();
        $msg = new SetHeldItemAction();
        $msg->setPlayerUuid($playerUuid);
        if ($main !== null) {
            $msg->setMain($main);
        }
        if ($offhand !== null) {
            $msg->setOffhand($offhand);
        }
        $action->setSetHeldItem($msg);
        $this->sendAction($action);
    }

    public function setHealthUuid(string $playerUuid, float $health, ?float $maxHealth = null): void {
        $action = new Action();
        $msg = new SetHealthAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setHealth($health);
        if ($maxHealth !== null) {
            $msg->setMaxHealth($maxHealth);
        }
        $action->setSetHealth($msg);
        $this->sendAction($action);
    }

    public function setFoodUuid(string $playerUuid, int $food): void {
        $action = new Action();
        $msg = new SetFoodAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setFood($food);
        $action->setSetFood($msg);
        $this->sendAction($action);
    }

    public function setExperienceUuid(string $playerUuid, ?int $level = null, ?float $progress = null, ?int $amount = null): void {
        $action = new Action();
        $msg = new SetExperienceAction();
        $msg->setPlayerUuid($playerUuid);
        if ($level !== null) {
            $msg->setLevel($level);
        }
        if ($progress !== null) {
            $msg->setProgress($progress);
        }
        if ($amount !== null) {
            $msg->setAmount($amount);
        }
        $action->setSetExperience($msg);
        $this->sendAction($action);
    }

    public function setVelocityUuid(string $playerUuid, Vec3 $velocity): void {
        $action = new Action();
        $msg = new SetVelocityAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setVelocity($velocity);
        $action->setSetVelocity($msg);
        $this->sendAction($action);
    }

    public function addEffectUuid(string $playerUuid, int $effectType, int $level, int $durationMs, bool $showParticles = true): void {
        $action = new Action();
        $msg = new AddEffectAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setEffectType($effectType);
        $msg->setLevel($level);
        $msg->setDurationMs($durationMs);
        $msg->setShowParticles($showParticles);
        $action->setAddEffect($msg);
        $this->sendAction($action);
    }

    public function removeEffectUuid(string $playerUuid, int $effectType): void {
        $action = new Action();
        $msg = new RemoveEffectAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setEffectType($effectType);
        $action->setRemoveEffect($msg);
        $this->sendAction($action);
    }

    public function sendTitleUuid(string $playerUuid, string $title, ?string $subtitle = null, ?int $fadeInMs = null, ?int $durationMs = null, ?int $fadeOutMs = null): void {
        $action = new Action();
        $msg = new SendTitleAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setTitle($title);
        if ($subtitle !== null) {
            $msg->setSubtitle($subtitle);
        }
        if ($fadeInMs !== null) {
            $msg->setFadeInMs($fadeInMs);
        }
        if ($durationMs !== null) {
            $msg->setDurationMs($durationMs);
        }
        if ($fadeOutMs !== null) {
            $msg->setFadeOutMs($fadeOutMs);
        }
        $action->setSendTitle($msg);
        $this->sendAction($action);
    }

    public function sendPopupUuid(string $playerUuid, string $message): void {
        $action = new Action();
        $msg = new SendPopupAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setMessage($message);
        $action->setSendPopup($msg);
        $this->sendAction($action);
    }

    public function sendTipUuid(string $playerUuid, string $message): void {
        $action = new Action();
        $msg = new SendTipAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setMessage($message);
        $action->setSendTip($msg);
        $this->sendAction($action);
    }

    public function playSoundUuid(string $playerUuid, int $sound, ?Vec3 $position = null, ?float $volume = null, ?float $pitch = null): void {
        $action = new Action();
        $msg = new PlaySoundAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setSound($sound);
        if ($position !== null) {
            $msg->setPosition($position);
        }
        if ($volume !== null) {
            $msg->setVolume($volume);
        }
        if ($pitch !== null) {
            $msg->setPitch($pitch);
        }
        $action->setPlaySound($msg);
        $this->sendAction($action);
    }

    public function executeCommandUuid(string $playerUuid, string $command): void {
        $action = new Action();
        $msg = new ExecuteCommandAction();
        $msg->setPlayerUuid($playerUuid);
        $msg->setCommand($command);
        $action->setExecuteCommand($msg);
        $this->sendAction($action);
    }

    public function worldSetDefaultGameMode(WorldRef $world, int $gameMode): void {
        $action = new Action();
        $msg = new WorldSetDefaultGameModeAction();
        $msg->setWorld($world);
        $msg->setGameMode($gameMode);
        $action->setWorldSetDefaultGameMode($msg);
        $this->sendAction($action);
    }

    public function worldSetDifficulty(WorldRef $world, int $difficulty): void {
        $action = new Action();
        $msg = new WorldSetDifficultyAction();
        $msg->setWorld($world);
        $msg->setDifficulty($difficulty);
        $action->setWorldSetDifficulty($msg);
        $this->sendAction($action);
    }

    public function worldSetTickRange(WorldRef $world, int $tickRange): void {
        $action = new Action();
        $msg = new WorldSetTickRangeAction();
        $msg->setWorld($world);
        $msg->setTickRange($tickRange);
        $action->setWorldSetTickRange($msg);
        $this->sendAction($action);
    }

    public function worldSetBlock(WorldRef $world, BlockPos $position, ?BlockState $block = null): void {
        $action = new Action();
        $msg = new WorldSetBlockAction();
        $msg->setWorld($world);
        $msg->setPosition($position);
        if ($block !== null) {
            $msg->setBlock($block);
        }
        $action->setWorldSetBlock($msg);
        $this->sendAction($action);
    }

    public function worldPlaySound(WorldRef $world, int $sound, Vec3 $position): void {
        $action = new Action();
        $msg = new WorldPlaySoundAction();
        $msg->setWorld($world);
        $msg->setSound($sound);
        $msg->setPosition($position);
        $action->setWorldPlaySound($msg);
        $this->sendAction($action);
    }

    public function worldAddParticle(WorldRef $world, Vec3 $position, int $particle, ?BlockState $block = null, ?int $face = null): void {
        $action = new Action();
        $msg = new WorldAddParticleAction();
        $msg->setWorld($world);
        $msg->setPosition($position);
        $msg->setParticle($particle);
        if ($block !== null) {
            $msg->setBlock($block);
        }
        if ($face !== null) {
            $msg->setFace($face);
        }
        $action->setWorldAddParticle($msg);
        $this->sendAction($action);
    }

    public function worldQueryEntities(WorldRef $world, ?string $correlationId = null): void {
        $action = new Action();
        if ($correlationId !== null) {
            $action->setCorrelationId($correlationId);
        }
        $msg = new WorldQueryEntitiesAction();
        $msg->setWorld($world);
        $action->setWorldQueryEntities($msg);
        $this->sendAction($action);
    }

    public function worldQueryPlayers(WorldRef $world, ?string $correlationId = null): void {
        $action = new Action();
        if ($correlationId !== null) {
            $action->setCorrelationId($correlationId);
        }
        $msg = new WorldQueryPlayersAction();
        $msg->setWorld($world);
        $action->setWorldQueryPlayers($msg);
        $this->sendAction($action);
    }

    public function worldQueryEntitiesWithin(WorldRef $world, BBox $box, ?string $correlationId = null): void {
        $action = new Action();
        if ($correlationId !== null) {
            $action->setCorrelationId($correlationId);
        }
        $msg = new WorldQueryEntitiesWithinAction();
        $msg->setWorld($world);
        $msg->setBox($box);
        $action->setWorldQueryEntitiesWithin($msg);
        $this->sendAction($action);
    }
}
