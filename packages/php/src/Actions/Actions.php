<?php

namespace Dragonfly\PluginLib\Actions;

use Df\Plugin\Action;
use Df\Plugin\ActionBatch;
use Df\Plugin\AddEffectAction;
use Df\Plugin\PluginToHost;
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
    private ?ActionBatch  = null;

    public function __construct(
        private StreamSender ,
        private string ,
    ) {}

    public function startBatch(): void {
        ->activeBatch = new ActionBatch();
    }

    public function commitBatch(): void {
        if (->activeBatch !== null && count(->activeBatch->getActions()) > 0) {
             = new PluginToHost();
            ->setPluginId(->pluginId);
            ->setActions(->activeBatch);
            ->sender->enqueue();
        }
        ->activeBatch = null;
    }

    private function sendOrBatch(Action ): void {
        if (->activeBatch !== null) {
             = ->activeBatch->getActions();
            [] = ;
            ->activeBatch->setActions();
        } else {
            ->sendAction();
        }
    }

    public function sendActions(array ): void {
         = new ActionBatch();
        ->setActions();

         = new PluginToHost();
        ->setPluginId(->pluginId);
        ->setActions();
        ->sender->enqueue();
    }

    private function sendAction(Action ): void {
         = new ActionBatch();
        ->setActions([]);

         = new PluginToHost();
        ->setPluginId(->pluginId);
        ->setActions();
        ->sender->enqueue();
    }

    public function chatToUuid(string , string ): void {
         = new Action();
         = new SendChatAction();
        ->setTargetUuid();
        ->setMessage();
        ->setSendChat();
        ->sendOrBatch();
    }

    public function teleportUuid(string , ?Vec3  = null, ?Vec3  = null): void {
         = new Action();
         = new TeleportAction();
        ->setPlayerUuid();
        if ( !== null) {
            ->setPosition();
        }
        if ( !== null) {
            ->setRotation();
        }
        ->setTeleport();
        ->sendOrBatch();
    }

    public function kickUuid(string , string ): void {
         = new Action();
         = new KickAction();
        ->setPlayerUuid();
        ->setReason();
        ->setKick();
        ->sendOrBatch();
    }

    public function setGameModeUuid(string , int ): void {
         = new Action();
         = new SetGameModeAction();
        ->setPlayerUuid();
        ->setGameMode();
        ->setSetGameMode();
        ->sendOrBatch();
    }

    public function giveItemUuid(string , ItemStack ): void {
         = new Action();
         = new GiveItemAction();
        ->setPlayerUuid();
        ->setItem();
        ->setGiveItem();
        ->sendOrBatch();
    }

    public function clearInventoryUuid(string ): void {
         = new Action();
         = new ClearInventoryAction();
        ->setPlayerUuid();
        ->setClearInventory();
        ->sendOrBatch();
    }

    public function setHeldItemsUuid(string , ?ItemStack  = null, ?ItemStack  = null): void {
         = new Action();
         = new SetHeldItemAction();
        ->setPlayerUuid();
        if ( !== null) {
            ->setMain();
        }
        if ( !== null) {
            ->setOffhand();
        }
        ->setSetHeldItem();
        ->sendOrBatch();
    }

    public function setHealthUuid(string , float , ?float  = null): void {
         = new Action();
         = new SetHealthAction();
        ->setPlayerUuid();
        ->setHealth();
        if ( !== null) {
            ->setMaxHealth();
        }
        ->setSetHealth();
        ->sendOrBatch();
    }

    public function setFoodUuid(string , int ): void {
         = new Action();
         = new SetFoodAction();
        ->setPlayerUuid();
        ->setFood();
        ->setSetFood();
        ->sendOrBatch();
    }

    public function setExperienceUuid(string , ?int  = null, ?float  = null, ?int  = null): void {
         = new Action();
         = new SetExperienceAction();
        ->setPlayerUuid();
        if ( !== null) {
            ->setLevel();
        }
        if ( !== null) {
            ->setProgress();
        }
        if ( !== null) {
            ->setAmount();
        }
        ->setSetExperience();
        ->sendOrBatch();
    }

    public function setVelocityUuid(string , Vec3 ): void {
         = new Action();
         = new SetVelocityAction();
        ->setPlayerUuid();
        ->setVelocity();
        ->setSetVelocity();
        ->sendOrBatch();
    }

    public function addEffectUuid(string , int , int , int , bool  = true): void {
         = new Action();
         = new AddEffectAction();
        ->setPlayerUuid();
        ->setEffectType();
        ->setLevel();
        ->setDurationMs();
        ->setShowParticles();
        ->setAddEffect();
        ->sendOrBatch();
    }

    public function removeEffectUuid(string , int ): void {
         = new Action();
         = new RemoveEffectAction();
        ->setPlayerUuid();
        ->setEffectType();
        ->setRemoveEffect();
        ->sendOrBatch();
    }

    public function sendTitleUuid(string , string , ?string  = null, ?int  = null, ?int  = null, ?int  = null): void {
         = new Action();
         = new SendTitleAction();
        ->setPlayerUuid();
        ->setTitle();
        if ( !== null) {
            ->setSubtitle();
        }
        if ( !== null) {
            ->setFadeInMs();
        }
        if ( !== null) {
            ->setDurationMs();
        }
        if ( !== null) {
            ->setFadeOutMs();
        }
        ->setSendTitle();
        ->sendOrBatch();
    }

    public function sendPopupUuid(string , string ): void {
         = new Action();
         = new SendPopupAction();
        ->setPlayerUuid();
        ->setMessage();
        ->setSendPopup();
        ->sendOrBatch();
    }

    public function sendTipUuid(string , string ): void {
         = new Action();
         = new SendTipAction();
        ->setPlayerUuid();
        ->setMessage();
        ->setSendTip();
        ->sendOrBatch();
    }

    public function playSoundUuid(string , int , ?Vec3  = null, ?float  = null, ?float  = null): void {
         = new Action();
         = new PlaySoundAction();
        ->setPlayerUuid();
        ->setSound();
        if ( !== null) {
            ->setPosition();
        }
        if ( !== null) {
            ->setVolume();
        }
        if ( !== null) {
            ->setPitch();
        }
        ->setPlaySound();
        ->sendOrBatch();
    }

    public function executeCommandUuid(string , string ): void {
         = new Action();
         = new ExecuteCommandAction();
        ->setPlayerUuid();
        ->setCommand();
        ->setExecuteCommand();
        ->sendOrBatch();
    }

    public function worldSetDefaultGameMode(WorldRef , int ): void {
         = new Action();
         = new WorldSetDefaultGameModeAction();
        ->setWorld();
        ->setGameMode();
        ->setWorldSetDefaultGameMode();
        ->sendOrBatch();
    }

    public function worldSetDifficulty(WorldRef , int ): void {
         = new Action();
         = new WorldSetDifficultyAction();
        ->setWorld();
        ->setDifficulty();
        ->setWorldSetDifficulty();
        ->sendOrBatch();
    }

    public function worldSetTickRange(WorldRef , int ): void {
         = new Action();
         = new WorldSetTickRangeAction();
        ->setWorld();
        ->setTickRange();
        ->setWorldSetTickRange();
        ->sendOrBatch();
    }

    public function worldSetBlock(WorldRef , BlockPos , ?BlockState  = null): void {
         = new Action();
         = new WorldSetBlockAction();
        ->setWorld();
        ->setPosition();
        if ( !== null) {
            ->setBlock();
        }
        ->setWorldSetBlock();
        ->sendOrBatch();
    }

    public function worldPlaySound(WorldRef , int , Vec3 ): void {
         = new Action();
         = new WorldPlaySoundAction();
        ->setWorld();
        ->setSound();
        ->setPosition();
        ->setWorldPlaySound();
        ->sendOrBatch();
    }

    public function worldAddParticle(WorldRef , Vec3 , int , ?BlockState  = null, ?int  = null): void {
         = new Action();
         = new WorldAddParticleAction();
        ->setWorld();
        ->setPosition();
        ->setParticle();
        if ( !== null) {
            ->setBlock();
        }
        if ( !== null) {
            ->setFace();
        }
        ->setWorldAddParticle();
        ->sendOrBatch();
    }

    public function worldQueryEntities(WorldRef , ?string  = null): void {
         = new Action();
        if ( !== null) {
            ->setCorrelationId();
        }
         = new WorldQueryEntitiesAction();
        ->setWorld();
        ->setWorldQueryEntities();
        ->sendOrBatch();
    }

    public function worldQueryPlayers(WorldRef , ?string  = null): void {
         = new Action();
        if ( !== null) {
            ->setCorrelationId();
        }
         = new WorldQueryPlayersAction();
        ->setWorld();
        ->setWorldQueryPlayers();
        ->sendOrBatch();
    }

    public function worldQueryEntitiesWithin(WorldRef , BBox , ?string  = null): void {
         = new Action();
        if ( !== null) {
            ->setCorrelationId();
        }
         = new WorldQueryEntitiesWithinAction();
        ->setWorld();
        ->setBox();
        ->setWorldQueryEntitiesWithin();
        ->sendOrBatch();
    }
}
