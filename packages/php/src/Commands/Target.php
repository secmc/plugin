<?php

namespace Dragonfly\PluginLib\Commands;

final class Target {
    public function __construct(
        public readonly string $uuid,
        public readonly string $name = ''
    ) {}
}
