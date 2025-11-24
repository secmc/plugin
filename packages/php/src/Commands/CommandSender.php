<?php

namespace Dragonfly\PluginLib\Commands;

use Df\Plugin\WorldRef;
use Dragonfly\PluginLib\Actions\Actions;
use Dragonfly\PluginLib\Entity\Player;

class CommandSender extends Player {
    public function __construct(
        public string $uuid,
        public string $name,
        Actions $actions,
        ?WorldRef $world = null,
    ) {
        parent::__construct($uuid, $name, $actions, $world);
    }
}
