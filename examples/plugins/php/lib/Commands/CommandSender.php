<?php

namespace Dragonfly\PluginLib\Commands;

class CommandSender {
    public function __construct(
        public string $uuid,
        public string $name,
    ) {}
}
