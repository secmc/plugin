<?php

namespace Dragonfly\PluginLib\Actions;

use Df\Plugin\Action;
use Df\Plugin\ActionBatch;
use Df\Plugin\PluginToHost;
use Df\Plugin\SendChatAction;
use Df\Plugin\TeleportAction;
use Df\Plugin\KickAction;
use Df\Plugin\SetGameModeAction;
use Df\Plugin\Vec3;
use Dragonfly\PluginLib\StreamSender;

final class Actions {
    public function __construct(
        private StreamSender $sender,
        private string $pluginId,
    ) {}

    public function sendAction(Action $action): void {
        $batch = new ActionBatch();
        $batch->setActions([$action]);

        $resp = new PluginToHost();
        $resp->setPluginId($this->pluginId);
        $resp->setActions($batch);
        $this->sender->enqueue($resp);
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
}
