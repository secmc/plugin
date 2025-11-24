<?php

namespace Dragonfly\PluginLib\Commands;

/**
 * Optional parameter wrapper. Optional parameters must come after required ones,
 * and may not be followed by non-optional parameters.
 */
class Optional {
    private mixed $value = null;
    private bool $hasValue = false;

    public function set(mixed $value): void {
        $this->value = $value;
        $this->hasValue = true;
    }

    public function hasValue(): bool {
        return $this->hasValue;
    }

    public function get(): mixed {
        return $this->value;
    }

    public function getOr(mixed $default): mixed {
        return $this->hasValue ? $this->value : $default;
    }
}
