<?php

namespace Dragonfly\PluginLib\Commands;

/**
 * Varargs consumes all remaining command arguments as a single string.
 * Must be the last parameter in a command class.
 */
class Varargs {
    public function __construct(public string $value) {}

    public function __toString(): string {
        return $this->value;
    }
}
