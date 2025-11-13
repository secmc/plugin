<?php

namespace Dragonfly\PluginLib\Commands;

use Dragonfly\PluginLib\Events\EventContext;
use ReflectionClass;
use ReflectionNamedType;
use ReflectionProperty;
use ReflectionType;
use RuntimeException;

/**
 * Base command class with reflection-based argument parsing.
 *
 * Supported parameter types:
 * - int, float, bool, string
 * - Varargs (must be last)
 * - Optional (wrapper; optional params must be last, may be multiple)
 *
 * Define parameters as public properties on a subclass, in the order they
 * should be parsed. Example:
 *
 *   class TpCommand extends Command {
 *       protected string $name = 'tpc';
 *       protected string $description = 'Teleport to coordinates';
 *       public float $x;
 *       /** @var Optional<float> *\/
 *       public Optional $y;
 *       public float $z;
 *       public function execute(CommandSender $sender, EventContext $ctx): void { ... }
 *   }
 */
abstract class Command {
    // Metadata
    protected string $name = '';
    protected string $description = '';
    /** @var string[] */
    protected array $aliases = [];

    abstract public function execute(CommandSender $sender, EventContext $ctx): void;

    public function getName(): string {
        return $this->name;
    }

    public function getDescription(): string {
        return $this->description;
    }

    /**
     * @return string[]
     */
    public function getAliases(): array {
        return $this->aliases;
    }

    /**
     * Parse command arguments. Returns true on success, false on usage error.
     *
     * @param string[] $rawArgs
     */
    public function parseArgs(array $rawArgs): bool {
        try {
            $this->validateSignature();
        } catch (\Throwable) {
            return false;
        }
        $schema = $this->inspectParameters();
        $ref = new ReflectionClass($this);
        $props = $this->getCommandProperties($ref);
        $propMap = [];
        foreach ($props as $p) {
            $p->setAccessible(true);
            $propMap[$p->getName()] = $p;
        }

        $argIndex = 0;
        $argCount = count($rawArgs);
        $paramCount = count($schema);

        foreach ($schema as $idx => $param) {
            $name = $param['name'];
            $type = $param['type'];      // int|float|bool|string|varargs
            $optional = !empty($param['optional']);

            $prop = $propMap[$name] ?? null;
            if (!$prop) {
                return false;
            }

            if ($type === 'varargs') {
                if ($idx !== $paramCount - 1) {
                    return false;
                }
                $remaining = array_slice($rawArgs, $argIndex);
                $prop->setValue($this, new Varargs(implode(' ', $remaining)));
                return true;
            }

            if ($argIndex >= $argCount) {
                if ($optional) {
                    if ($this->getTypeName($prop->getType()) === Optional::class) {
                        $prop->setValue($this, new Optional());
                        continue;
                    }
                    return false;
                }
                return false;
            }

            $parsed = $this->parseTypedValue($rawArgs[$argIndex], $type);
            if ($parsed === null && $type !== 'string') {
                return false;
            }

            if ($this->getTypeName($prop->getType()) === Optional::class) {
                $opt = new Optional();
                $opt->set($parsed);
                $prop->setValue($this, $opt);
            } else {
                $prop->setValue($this, $parsed);
            }
            $argIndex++;
        }
        if ($argIndex < $argCount) {
            return false;
        }
        return true;
    }

    /**
     * Validate parameter ordering rules:
     * - Optional parameters may only appear at the end (can be multiple).
     * - Varargs must be the final parameter.
     */
    public function validateSignature(): void {
        $ref = new ReflectionClass($this);
        $props = $this->getCommandProperties($ref);

        $seenOptional = false;
        foreach ($props as $index => $prop) {
            $typeName = $this->getTypeName($prop->getType());
            if ($typeName === Varargs::class) {
                if ($index !== count($props) - 1) {
                    throw new RuntimeException('Varargs must be the last parameter.');
                }
                continue;
            }
            if ($seenOptional && $typeName !== Optional::class) {
                throw new RuntimeException('Optional parameters must be at the end.');
            }
            if ($typeName === Optional::class) {
                $seenOptional = true;
            }
        }
    }

    /**
     * Generate a human-friendly usage string.
     */
    public function generateUsage(): string {
        $parts = ['/' . $this->name];
        foreach ($this->inspectParameters() as $p) {
            $name = $p['name'];
            $type = $p['type'];
            $optional = !empty($p['optional']);
            if ($type === 'varargs') {
                $parts[] = '<' . $name . '...>';
            } elseif ($optional) {
                $parts[] = '[' . $name . ']';
            } else {
                $parts[] = '<' . $name . '>';
            }
        }
        return implode(' ', $parts);
    }

    /**
     * Export parameter specification for transport to the host (Go) side.
     * Format: list of ['name' => string, 'type' => string, 'optional' => bool]
     * Types: int|float|bool|string|varargs
     *
     * @return array<int, array{name:string,type:string,optional?:bool}>
     */
    public function serializeParamSpec(): array {
        return $this->inspectParameters();
    }

    /**
     * @return ReflectionProperty[]
     */
    private function getCommandProperties(ReflectionClass $ref): array {
        $props = $ref->getProperties(ReflectionProperty::IS_PUBLIC);
        $filtered = [];
        foreach ($props as $p) {
            $n = $p->getName();
            if ($n === 'name' || $n === 'description' || $n === 'aliases') {
                continue;
            }
            $filtered[] = $p;
        }
        return $filtered;
    }

    private function getTypeName(?ReflectionType $type): ?string {
        if ($type instanceof ReflectionNamedType) {
            return $type->getName();
        }
        return null;
    }

    private function parseTypedValue(string $arg, ?string $typeName): mixed {
        return match ($typeName) {
            'int' => filter_var($arg, FILTER_VALIDATE_INT),
            'float' => filter_var($arg, FILTER_VALIDATE_FLOAT),
            'bool' => $this->parseBool($arg),
            null, 'string' => $arg,
            default => null,
        };
    }

    private function parseBool(string $arg): ?bool {
        $v = strtolower($arg);
        return match ($v) {
            'true', '1', 'yes', 'on' => true,
            'false', '0', 'no', 'off' => false,
            default => null,
        };
    }

    /**
     * Build a normalized parameter schema from the command's public properties.
     * @return array<int, array{name:string,type:string,optional?:bool}>
     */
    private function inspectParameters(): array {
        $ref = new ReflectionClass($this);
        $props = $this->getCommandProperties($ref);
        $out = [];
        foreach ($props as $prop) {
            $name = $prop->getName();
            $typeName = $this->getTypeName($prop->getType());
            if ($typeName === Varargs::class) {
                $out[] = ['name' => $name, 'type' => 'varargs'];
                break;
            }
            if ($typeName === Optional::class) {
                $t = $this->getOptionalWrappedType($prop);
                $out[] = ['name' => $name, 'type' => $t, 'optional' => true];
                continue;
            }
            $mapped = match ($typeName) {
                'int' => 'int',
                'float', 'double' => 'float',
                'bool' => 'bool',
                default => 'string',
            };
            $out[] = ['name' => $name, 'type' => $mapped];
        }
        return $out;
    }

    /**
     * Convenience: attach enum values to a parameter in a schema.
     *
     * @param array<int, array{name:string,type:string,optional?:bool,enum_values?:array<int,string>}> $schema
     * @param string $paramName
     * @param string[] $values
     * @return array
     */
    protected function withEnum(array $schema, string $paramName, array $values): array {
        foreach ($schema as &$p) {
            if ($p['name'] === $paramName) {
                $p['enum_values'] = array_values($values);
                $p['type'] = 'enum';
                break;
            }
        }
        return $schema;
    }

    /**
     * Convenience: enum names from a class' constants, with optional excludes.
     *
     * @param string $class Fully-qualified class name
     * @param string[] $excludeNames
     * @return string[]
     */
    protected function enumNamesFromClass(string $class, array $excludeNames = []): array {
        $names = array_keys((new \ReflectionClass($class))->getConstants());
        if (!empty($excludeNames)) {
            $names = array_values(array_filter($names, fn ($n) => !in_array($n, $excludeNames, true)));
        }
        return $names;
    }

    /**
     * Attempt to infer the wrapped type for Optional<T> from @var docblock.
     */
    private function getOptionalWrappedType(ReflectionProperty $prop): string {
        $doc = $prop->getDocComment() ?: '';
        if (preg_match('/@var\s+Optional<\s*([A-Za-z_][A-Za-z0-9_]*)\s*>/i', $doc, $m)) {
            $t = strtolower($m[1]);
            return match ($t) {
                'int' => 'int',
                'float', 'double' => 'float',
                'bool', 'boolean' => 'bool',
                'string' => 'string',
                default => 'string',
            };
        }
        // Default to string if not annotated.
        return 'string';
    }
}
