<?php

namespace Dragonfly\PluginLib\Commands;

/**
 * Interface for anything that can execute commands (players, console, etc.)
 */
interface CommandSender {
    public function sendMessage(string $message): void;
    public function getName(): string;
}
