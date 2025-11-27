<?php

namespace Dragonfly\PluginLib\Actions;

use Df\Plugin\Action;
use Df\Plugin\Vec3;
use Df\Plugin\ItemStack;
use Df\Plugin\WorldRef;
use Df\Plugin\BlockPos;
use Df\Plugin\BlockState;
use Df\Plugin\BBox;

trait ActionsTrait {
    abstract protected function getActions(): Actions;

    public function startBatch(): void {
        ->getActions()->startBatch();
    }

    public function commitBatch(): void {
        ->getActions()->commitBatch();
    }

    public function sendActions(array ): void {
        ->getActions()->sendActions();
    }

    public function chatToUuid(string , string ): void {
        ->getActions()->chatToUuid(, );
    }

    public function teleportUuid(string , ?Vec3  = null, ?Vec3  = null): void {
        ->getActions()->teleportUuid(, , );
    }

    public function kickUuid(string , string ): void {
        ->getActions()->kickUuid(, );
    }

    public function setGameModeUuid(string , int ): void {
        ->getActions()->setGameModeUuid(, );
    }

    public function giveItemUuid(string , ItemStack ): void {
        ->getActions()->giveItemUuid(, );
    }

    public function clearInventoryUuid(string ): void {
        ->getActions()->clearInventoryUuid();
    }

    public function setHeldItemsUuid(string , ?ItemStack  = null, ?ItemStack  = null): void {
        ->getActions()->setHeldItemsUuid(, , );
    }

    public function setHealthUuid(string , float , ?float  = null): void {
        ->getActions()->setHealthUuid(, , );
    }

    public function setFoodUuid(string , int ): void {
        ->getActions()->setFoodUuid(, );
    }

    public function setExperienceUuid(string , ?int  = null, ?float  = null, ?int  = null): void {
        ->getActions()->setExperienceUuid(, , , );
    }

    public function setVelocityUuid(string , Vec3 ): void {
        ->getActions()->setVelocityUuid(, );
    }

    public function addEffectUuid(string , int , int , int , bool  = true): void {
        ->getActions()->addEffectUuid(, , , , );
    }

    public function removeEffectUuid(string , int ): void {
        ->getActions()->removeEffectUuid(, );
    }

    public function sendTitleUuid(string , string , ?string  = null, ?int  = null, ?int  = null, ?int  = null): void {
        ->getActions()->sendTitleUuid(, , , , , );
    }

    public function sendPopupUuid(string , string ): void {
        ->getActions()->sendPopupUuid(, );
    }

    public function sendTipUuid(string , string ): void {
        ->getActions()->sendTipUuid(, );
    }

    public function playSoundUuid(string , int , ?Vec3  = null, ?float  = null, ?float  = null): void {
        ->getActions()->playSoundUuid(, , , , );
    }

    public function executeCommandUuid(string , string ): void {
        ->getActions()->executeCommandUuid(, );
    }

    public function worldSetDefaultGameMode(WorldRef , int ): void {
        ->getActions()->worldSetDefaultGameMode(, );
    }

    public function worldSetDifficulty(WorldRef , int ): void {
        ->getActions()->worldSetDifficulty(, );
    }

    public function worldSetTickRange(WorldRef , int ): void {
        ->getActions()->worldSetTickRange(, );
    }

    public function worldSetBlock(WorldRef , BlockPos , ?BlockState  = null): void {
        ->getActions()->worldSetBlock(, , );
    }

    public function worldPlaySound(WorldRef , int , Vec3 ): void {
        ->getActions()->worldPlaySound(, , );
    }

    public function worldAddParticle(WorldRef , Vec3 , int , ?BlockState  = null, ?int  = null): void {
        ->getActions()->worldAddParticle(, , , , );
    }

    public function worldQueryEntities(WorldRef , ?string  = null): void {
        ->getActions()->worldQueryEntities(, );
    }

    public function worldQueryPlayers(WorldRef , ?string  = null): void {
        ->getActions()->worldQueryPlayers(, );
    }

    public function worldQueryEntitiesWithin(WorldRef , BBox , ?string  = null): void {
        ->getActions()->worldQueryEntitiesWithin(, , );
    }
}
