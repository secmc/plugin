use crate::types;
use crate::event::EventContext;
impl<'a> EventContext<'a, types::ChatEvent> {
    ///Sets the `message` for this event.
    pub fn set_message(&mut self, message: String) {
        let mutation = types::ChatMutation {
            message: Some(message.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::Chat(mutation));
    }
}
impl<'a> EventContext<'a, types::BlockBreakEvent> {
    ///Sets the `drops` for this event.
    pub fn set_drops(&mut self, drops: Vec<types::ItemStack>) {
        let mutation = types::BlockBreakMutation {
            drops: Some(types::ItemStackList {
                items: drops.into_iter().map(|s| s.into()).collect(),
            }),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::BlockBreak(mutation));
    }
    ///Sets the `xp` for this event.
    pub fn set_xp(&mut self, xp: i32) {
        let mutation = types::BlockBreakMutation {
            xp: Some(xp.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::BlockBreak(mutation));
    }
}
impl<'a> EventContext<'a, types::PlayerFoodLossEvent> {
    ///Sets the `to` for this event.
    pub fn set_to(&mut self, to: i32) {
        let mutation = types::PlayerFoodLossMutation {
            to: Some(to.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerFoodLoss(mutation));
    }
}
impl<'a> EventContext<'a, types::PlayerHealEvent> {
    ///Sets the `amount` for this event.
    pub fn set_amount(&mut self, amount: f64) {
        let mutation = types::PlayerHealMutation {
            amount: Some(amount.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerHeal(mutation));
    }
}
impl<'a> EventContext<'a, types::PlayerHurtEvent> {
    ///Sets the `damage` for this event.
    pub fn set_damage(&mut self, damage: f64) {
        let mutation = types::PlayerHurtMutation {
            damage: Some(damage.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerHurt(mutation));
    }
    ///Sets the `attack_immunity_ms` for this event.
    pub fn set_attack_immunity_ms(&mut self, attack_immunity_ms: i64) {
        let mutation = types::PlayerHurtMutation {
            attack_immunity_ms: Some(attack_immunity_ms.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerHurt(mutation));
    }
}
impl<'a> EventContext<'a, types::PlayerDeathEvent> {
    ///Sets the `keep_inventory` for this event.
    pub fn set_keep_inventory(&mut self, keep_inventory: bool) {
        let mutation = types::PlayerDeathMutation {
            keep_inventory: Some(keep_inventory.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerDeath(mutation));
    }
}
impl<'a> EventContext<'a, types::PlayerRespawnEvent> {
    ///Sets the `position` for this event.
    pub fn set_position(&mut self, position: types::Vec3) {
        let mutation = types::PlayerRespawnMutation {
            position: Some(position.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerRespawn(mutation));
    }
    ///Sets the `world` for this event.
    pub fn set_world(&mut self, world: types::WorldRef) {
        let mutation = types::PlayerRespawnMutation {
            world: Some(world.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerRespawn(mutation));
    }
}
impl<'a> EventContext<'a, types::PlayerAttackEntityEvent> {
    ///Sets the `force` for this event.
    pub fn set_force(&mut self, force: f64) {
        let mutation = types::PlayerAttackEntityMutation {
            force: Some(force.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerAttackEntity(mutation));
    }
    ///Sets the `height` for this event.
    pub fn set_height(&mut self, height: f64) {
        let mutation = types::PlayerAttackEntityMutation {
            height: Some(height.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerAttackEntity(mutation));
    }
    ///Sets the `critical` for this event.
    pub fn set_critical(&mut self, critical: bool) {
        let mutation = types::PlayerAttackEntityMutation {
            critical: Some(critical.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerAttackEntity(mutation));
    }
}
impl<'a> EventContext<'a, types::PlayerExperienceGainEvent> {
    ///Sets the `amount` for this event.
    pub fn set_amount(&mut self, amount: i32) {
        let mutation = types::PlayerExperienceGainMutation {
            amount: Some(amount.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerExperienceGain(mutation));
    }
}
impl<'a> EventContext<'a, types::PlayerLecternPageTurnEvent> {
    ///Sets the `new_page` for this event.
    pub fn set_new_page(&mut self, new_page: i32) {
        let mutation = types::PlayerLecternPageTurnMutation {
            new_page: Some(new_page.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerLecternPageTurn(mutation));
    }
}
impl<'a> EventContext<'a, types::PlayerItemPickupEvent> {
    ///Sets the `item` for this event.
    pub fn set_item(&mut self, item: types::ItemStack) {
        let mutation = types::PlayerItemPickupMutation {
            item: Some(item.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerItemPickup(mutation));
    }
}
impl<'a> EventContext<'a, types::PlayerTransferEvent> {
    ///Sets the `address` for this event.
    pub fn set_address(&mut self, address: types::Address) {
        let mutation = types::PlayerTransferMutation {
            address: Some(address.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::PlayerTransfer(mutation));
    }
}
impl<'a> EventContext<'a, types::WorldExplosionEvent> {
    ///Sets the `entity_uuids` for this event.
    pub fn set_entity_uuids(&mut self, entity_uuids: Vec<String>) {
        let mutation = types::WorldExplosionMutation {
            entity_uuids: Some(types::StringList {
                values: entity_uuids.into_iter().map(|s| s.into()).collect(),
            }),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::WorldExplosion(mutation));
    }
    ///Sets the `blocks` for this event.
    pub fn set_blocks(&mut self, blocks: types::BlockPosList) {
        let mutation = types::WorldExplosionMutation {
            blocks: Some(blocks.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::WorldExplosion(mutation));
    }
    ///Sets the `item_drop_chance` for this event.
    pub fn set_item_drop_chance(&mut self, item_drop_chance: f64) {
        let mutation = types::WorldExplosionMutation {
            item_drop_chance: Some(item_drop_chance.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::WorldExplosion(mutation));
    }
    ///Sets the `spawn_fire` for this event.
    pub fn set_spawn_fire(&mut self, spawn_fire: bool) {
        let mutation = types::WorldExplosionMutation {
            spawn_fire: Some(spawn_fire.into()),
            ..Default::default()
        };
        self.set_mutation(types::EventResultUpdate::WorldExplosion(mutation));
    }
}
