<?php

namespace Dragonfly\PluginLib\Actions;

use Df\Plugin\Action;
use Df\Plugin\Vec3;

trait ActionsTrait {
    abstract protected function getActions(): Actions;

    public function sendAction(Action $action): void {
        $this->getActions()->sendAction($action);
    }

    public function chatToUuid(string $targetUuid, string $message): void {
        $this->getActions()->chatToUuid($targetUuid, $message);
    }

    public function teleportUuid(string $playerUuid, ?Vec3 $position = null, ?Vec3 $rotation = null): void {
        $this->getActions()->teleportUuid($playerUuid, $position, $rotation);
    }

    public function kickUuid(string $playerUuid, string $reason): void {
        $this->getActions()->kickUuid($playerUuid, $reason);
    }

    public function setGameModeUuid(string $playerUuid, int $gameMode): void {
        $this->getActions()->setGameModeUuid($playerUuid, $gameMode);
    }
}
