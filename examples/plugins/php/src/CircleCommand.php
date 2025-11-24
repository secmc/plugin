<?php

namespace ExamplePhp;

use Df\Plugin\ActionResult;
use Df\Plugin\ParticleType;
use Df\Plugin\Vec3;
use Df\Plugin\WorldRef;
use Dragonfly\PluginLib\Commands\Command;
use Dragonfly\PluginLib\Commands\CommandSender;
use Dragonfly\PluginLib\Commands\Optional;
use Dragonfly\PluginLib\Entity\Player;
use Dragonfly\PluginLib\Events\EventContext;
use Dragonfly\PluginLib\Util\EnumResolver;

class CircleCommand extends Command {
    protected string $name = 'circle';
    protected string $description = 'Spawn particles in a circle around all players';

    /** @var Optional<string> */
    public Optional $particle;

    public function execute(CommandSender $sender, EventContext $ctx): void {
        if (!$sender instanceof Player) {
            $sender->sendMessage("§cThis command can only be run by a player.");
            return;
        }

        if ($this->particle->hasValue()) {
            $particleName = $this->particle->get();
            $particleId = $this->resolveParticleId($particleName);
            if ($particleId === null) {
                $sender->sendMessage("§cUnknown particle: {$particleName}");
                return;
            }
        } else {
            $particleId = ParticleType::PARTICLE_FLAME;
            $particleName = 'flame';
        }

        $world = $sender->getWorld();
        if ($world === null) {
            $sender->sendMessage("§cCannot determine your world.");
            return;
        }

        $correlationId = uniqid('circle_', true);
        $ctx->onActionResult($correlationId, function (ActionResult $result) use ($ctx, $world, $particleId) {
            $playersResult = $result->getWorldPlayers();
            if ($playersResult === null) {
                return;
            }

            $radius = 3.0;
            $points = 16;

            foreach ($playersResult->getPlayers() as $player) {
                $pos = $player->getPosition();
                if ($pos === null) {
                    continue;
                }

                $cx = $pos->getX();
                $cy = $pos->getY();
                $cz = $pos->getZ();

                for ($i = 0; $i < $points; $i++) {
                    $angle = (2 * M_PI / $points) * $i;
                    $x = $cx + $radius * cos($angle);
                    $z = $cz + $radius * sin($angle);

                    $particlePos = new Vec3();
                    $particlePos->setX($x);
                    $particlePos->setY($cy + 0.5);
                    $particlePos->setZ($z);

                    $ctx->worldAddParticle($world, $particlePos, $particleId);
                }
            }
        });

        $ctx->worldQueryPlayers($world, $correlationId);
        $sender->sendMessage("§aSpawning {$particleName} circles around all players!");
    }

    /**
     * @return array<int, array{name:string,type:string,optional?:bool,enum_values?:array<int,string>}>
     */
    public function serializeParamSpec(): array {
        $names = EnumResolver::lowerNames(ParticleType::class, ['PARTICLE_TYPE_UNSPECIFIED']);
        return $this->withEnum(parent::serializeParamSpec(), 'particle', $names);
    }

    private function resolveParticleId(string $input): ?int {
        return EnumResolver::value(ParticleType::class, $input, 'PARTICLE_');
    }
}

