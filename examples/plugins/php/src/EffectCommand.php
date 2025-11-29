<?php

namespace ExamplePhp;

use Dragonfly\PluginLib\Commands\Command;
use Dragonfly\PluginLib\Commands\CommandSender;
use Dragonfly\PluginLib\Commands\Optional;
use Dragonfly\PluginLib\Commands\Target;
use Dragonfly\PluginLib\Entity\Player;
use Dragonfly\PluginLib\Events\EventContext;
use Df\Plugin\EffectType;
use Dragonfly\PluginLib\Util\EnumResolver;

class EffectCommand extends Command {
    protected string $name = 'effect';
    protected string $description = 'Apply an effect to a player';

    public Target $target;
    public string $effect;
    /** @var Optional<int> */
    public Optional $level;
    /** @var Optional<int> */
    public Optional $durationSeconds;
    /** @var Optional<bool> */
    public Optional $showParticles;

    public function execute(CommandSender $sender, EventContext $ctx): void {
        $target = $ctx->getServer()->getPlayer($this->target->uuid);
        if ($target === null) {
            $sender->sendMessage("§cPlayer not found or offline.");
            return;
        }

        $effectId = $this->resolveEffectId($this->effect);
        if ($effectId === null) {
            $sender->sendMessage("§cUnknown effect: {$this->effect}");
            return;
        }
        
        $level = max(1, (int)$this->level->getOr(1));
        $seconds = max(0, (int)$this->durationSeconds->getOr(30));
        $show = $this->showParticles->getOr(true);
        $durationMs = $seconds * 1000;

        $target->addEffect($effectId, $level, $durationMs, $show);
        $sender->sendMessage("Applied effect " . $this->enumName($effectId) . " (id {$effectId}) level {$level} for {$seconds}s" . ($show ? '' : ' (hidden)'));
    }

    /**
     * Provide enum values for the 'effect' parameter so the client shows suggestions.
     *
     * @return array<int, array{name:string,type:string,optional?:bool,enum_values?:array<int,string>}>
     */
    public function serializeParamSpec(): array {
        $names = EnumResolver::lowerNames(EffectType::class, ['EFFECT_UNKNOWN']);
        return $this->withEnum(parent::serializeParamSpec(), 'effect', $names);
    }

    private function resolveEffectId(string $input): ?int {
        return EnumResolver::value(EffectType::class, $input);
    }

    private function enumName(int $value): string {
        return EnumResolver::lowerName(EffectType::class, $value);
    }
}


