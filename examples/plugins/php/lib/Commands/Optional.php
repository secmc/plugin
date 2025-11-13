<?php

namespace Dragonfly\PluginLib\Commands;

/**
 * Optional parameter wrapper. Optional parameters must come after required ones,
 * and may not be followed by non-optional parameters.
 */
class Optional {
    private mixed $value = null;
    private bool $present = false;

    public function set(mixed $value): void {
        $this->value = $value;
        $this->present = true;
    }

    public function isPresent(): bool {
        return $this->present;
    }

    public function get(): mixed {
        return $this->value;
    }

    public function getOr(mixed $default): mixed {
        return $this->present ? $this->value : $default;
    }
}
