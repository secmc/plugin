<?php

namespace Dragonfly\PluginLib\Entity;

use Df\Plugin\ItemStack;
use Df\Plugin\Vec3;
use Df\Plugin\WorldRef;
use Dragonfly\PluginLib\Actions\Actions;
use Dragonfly\PluginLib\Commands\CommandSender;

final class Player implements CommandSender {
    public function __construct(
        private string $uuid,
        private string $name,
        private Actions $actions,
        private ?WorldRef $world = null,
    ) {}

    public function getUuid(): string {
        return $this->uuid;
    }

    public function getName(): string {
        return $this->name;
    }

    public function getWorld(): ?WorldRef {
        return $this->world;
    }

    public function setWorld(?WorldRef $world): void {
        $this->world = $world;
    }

    public function sendMessage(string $message): void {
        $this->actions->chatToUuid($this->uuid, $message);
    }

    public function teleport(?Vec3 $position = null, ?Vec3 $rotation = null): void {
        $this->actions->teleportUuid($this->uuid, $position, $rotation);
    }

    public function kick(string $reason): void {
        $this->actions->kickUuid($this->uuid, $reason);
    }

    public function setGameMode(int $gameMode): void {
        $this->actions->setGameModeUuid($this->uuid, $gameMode);
    }

    public function giveItem(ItemStack $item): void {
        $this->actions->giveItemUuid($this->uuid, $item);
    }

    public function clearInventory(): void {
        $this->actions->clearInventoryUuid($this->uuid);
    }

    public function setHeldItems(?ItemStack $main = null, ?ItemStack $offhand = null): void {
        $this->actions->setHeldItemsUuid($this->uuid, $main, $offhand);
    }

    public function setHealth(float $health, ?float $maxHealth = null): void {
        $this->actions->setHealthUuid($this->uuid, $health, $maxHealth);
    }

    public function setFood(int $food): void {
        $this->actions->setFoodUuid($this->uuid, $food);
    }

    public function setExperience(?int $level = null, ?float $progress = null, ?int $amount = null): void {
        $this->actions->setExperienceUuid($this->uuid, $level, $progress, $amount);
    }

    public function setVelocity(Vec3 $velocity): void {
        $this->actions->setVelocityUuid($this->uuid, $velocity);
    }

    public function addEffect(int $effectType, int $level, int $durationMs, bool $showParticles = true): void {
        $this->actions->addEffectUuid($this->uuid, $effectType, $level, $durationMs, $showParticles);
    }

    public function removeEffect(int $effectType): void {
        $this->actions->removeEffectUuid($this->uuid, $effectType);
    }

    public function sendTitle(string $title, ?string $subtitle = null, ?int $fadeInMs = null, ?int $durationMs = null, ?int $fadeOutMs = null): void {
        $this->actions->sendTitleUuid($this->uuid, $title, $subtitle, $fadeInMs, $durationMs, $fadeOutMs);
    }

    public function sendPopup(string $message): void {
        $this->actions->sendPopupUuid($this->uuid, $message);
    }

    public function sendTip(string $message): void {
        $this->actions->sendTipUuid($this->uuid, $message);
    }

    public function playSound(int $sound, ?Vec3 $position = null, ?float $volume = null, ?float $pitch = null): void {
        $this->actions->playSoundUuid($this->uuid, $sound, $position, $volume, $pitch);
    }

    public function executeCommand(string $command): void {
        $this->actions->executeCommandUuid($this->uuid, $command);
    }
}
