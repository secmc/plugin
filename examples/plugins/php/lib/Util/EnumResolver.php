<?php

namespace Dragonfly\PluginLib\Util;

use ReflectionClass;

final class EnumResolver {
    /**
     * Resolve an enum value by accepting either a numeric id or a name (case-insensitive).
     * Hyphens/spaces in names are normalized to underscores.
     */
    public static function value(string $enumClass, string $input): ?int {
        if (ctype_digit($input)) {
            return (int)$input;
        }
        $key = strtoupper(str_replace(['-', ' '], ['_', '_'], $input));
        $consts = self::constants($enumClass);
        if (isset($consts[$key])) {
            return (int)$consts[$key];
        }
        return null;
    }

    /**
     * Get the name for the given enum numeric value. Falls back to the numeric string if not found.
     */
    public static function name(string $enumClass, int $value): string {
        $consts = self::constants($enumClass);
        $name = array_search($value, $consts, true);
        return is_string($name) ? $name : (string)$value;
    }
    
    /**
     * Return all enum constant names, optionally excluding some.
     *
     * @param string[] $excludeNames
     * @return string[]
     */
    public static function names(string $enumClass, array $excludeNames = []): array {
        $names = array_keys(self::constants($enumClass));
        if (!empty($excludeNames)) {
            $exclude = array_flip($excludeNames);
            $names = array_values(array_filter($names, static fn (string $n) => !isset($exclude[$n])));
        }
        return $names;
    }
    
    /**
     * Lowercase variant of name().
     */
    public static function lowerName(string $enumClass, int $value): string {
        return strtolower(self::name($enumClass, $value));
    }
    
    /**
     * Lowercase variant of names().
     *
     * @param string[] $excludeNames
     * @return string[]
     */
    public static function lowerNames(string $enumClass, array $excludeNames = []): array {
        return array_map('strtolower', self::names($enumClass, $excludeNames));
    }

    /**
     * Cache and return enum constants for a given class.
     *
     * @return array<string, int>
     */
    private static function constants(string $enumClass): array {
        static $cache = [];
        if (isset($cache[$enumClass])) {
            return $cache[$enumClass];
        }
        $cache[$enumClass] = (new ReflectionClass($enumClass))->getConstants();
        return $cache[$enumClass];
    }
}
