use crate::{types, Server};
use tokio::sync::mpsc;
impl Server {
    ///Sends a `SendChat` action to the server.
    pub async fn send_chat(
        &self,
        target_uuid: String,
        message: String,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::SendChat(types::SendChatAction {
                    target_uuid: target_uuid.into(),
                    message: message.into(),
                }),
            )
            .await
    }
    ///Sends a `Teleport` action to the server.
    pub async fn teleport(
        &self,
        player_uuid: String,
        position: Option<types::Vec3>,
        rotation: Option<types::Vec3>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::Teleport(types::TeleportAction {
                    player_uuid: player_uuid.into(),
                    position: position.map(|v| v.into()),
                    rotation: rotation.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `Kick` action to the server.
    pub async fn kick(
        &self,
        player_uuid: String,
        reason: String,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::Kick(types::KickAction {
                    player_uuid: player_uuid.into(),
                    reason: reason.into(),
                }),
            )
            .await
    }
    ///Sends a `SetGameMode` action to the server.
    pub async fn set_game_mode(
        &self,
        player_uuid: String,
        game_mode: i32,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::SetGameMode(types::SetGameModeAction {
                    player_uuid: player_uuid.into(),
                    game_mode: game_mode.into(),
                }),
            )
            .await
    }
    ///Sends a `GiveItem` action to the server.
    pub async fn give_item(
        &self,
        player_uuid: String,
        item: Option<types::ItemStack>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::GiveItem(types::GiveItemAction {
                    player_uuid: player_uuid.into(),
                    item: item.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `ClearInventory` action to the server.
    pub async fn clear_inventory(
        &self,
        player_uuid: String,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::ClearInventory(types::ClearInventoryAction {
                    player_uuid: player_uuid.into(),
                }),
            )
            .await
    }
    ///Sends a `SetHeldItem` action to the server.
    pub async fn set_held_item(
        &self,
        player_uuid: String,
        main: Option<types::ItemStack>,
        offhand: Option<types::ItemStack>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::SetHeldItem(types::SetHeldItemAction {
                    player_uuid: player_uuid.into(),
                    main: main.map(|v| v.into()),
                    offhand: offhand.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `SetHealth` action to the server.
    pub async fn set_health(
        &self,
        player_uuid: String,
        health: f64,
        max_health: Option<f64>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::SetHealth(types::SetHealthAction {
                    player_uuid: player_uuid.into(),
                    health: health.into(),
                    max_health: max_health.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `SetFood` action to the server.
    pub async fn set_food(
        &self,
        player_uuid: String,
        food: i32,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::SetFood(types::SetFoodAction {
                    player_uuid: player_uuid.into(),
                    food: food.into(),
                }),
            )
            .await
    }
    ///Sends a `SetExperience` action to the server.
    pub async fn set_experience(
        &self,
        player_uuid: String,
        level: Option<i32>,
        progress: Option<f32>,
        amount: Option<i32>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::SetExperience(types::SetExperienceAction {
                    player_uuid: player_uuid.into(),
                    level: level.map(|v| v.into()),
                    progress: progress.map(|v| v.into()),
                    amount: amount.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `SetVelocity` action to the server.
    pub async fn set_velocity(
        &self,
        player_uuid: String,
        velocity: Option<types::Vec3>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::SetVelocity(types::SetVelocityAction {
                    player_uuid: player_uuid.into(),
                    velocity: velocity.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `AddEffect` action to the server.
    pub async fn add_effect(
        &self,
        player_uuid: String,
        effect_type: i32,
        level: i32,
        duration_ms: i64,
        show_particles: bool,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::AddEffect(types::AddEffectAction {
                    player_uuid: player_uuid.into(),
                    effect_type: effect_type.into(),
                    level: level.into(),
                    duration_ms: duration_ms.into(),
                    show_particles: show_particles.into(),
                }),
            )
            .await
    }
    ///Sends a `RemoveEffect` action to the server.
    pub async fn remove_effect(
        &self,
        player_uuid: String,
        effect_type: i32,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::RemoveEffect(types::RemoveEffectAction {
                    player_uuid: player_uuid.into(),
                    effect_type: effect_type.into(),
                }),
            )
            .await
    }
    ///Sends a `SendTitle` action to the server.
    pub async fn send_title(
        &self,
        player_uuid: String,
        title: String,
        subtitle: Option<String>,
        fade_in_ms: Option<i64>,
        duration_ms: Option<i64>,
        fade_out_ms: Option<i64>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::SendTitle(types::SendTitleAction {
                    player_uuid: player_uuid.into(),
                    title: title.into(),
                    subtitle: subtitle.map(|v| v.into()),
                    fade_in_ms: fade_in_ms.map(|v| v.into()),
                    duration_ms: duration_ms.map(|v| v.into()),
                    fade_out_ms: fade_out_ms.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `SendPopup` action to the server.
    pub async fn send_popup(
        &self,
        player_uuid: String,
        message: String,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::SendPopup(types::SendPopupAction {
                    player_uuid: player_uuid.into(),
                    message: message.into(),
                }),
            )
            .await
    }
    ///Sends a `SendTip` action to the server.
    pub async fn send_tip(
        &self,
        player_uuid: String,
        message: String,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::SendTip(types::SendTipAction {
                    player_uuid: player_uuid.into(),
                    message: message.into(),
                }),
            )
            .await
    }
    ///Sends a `PlaySound` action to the server.
    pub async fn play_sound(
        &self,
        player_uuid: String,
        sound: i32,
        position: Option<types::Vec3>,
        volume: Option<f32>,
        pitch: Option<f32>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::PlaySound(types::PlaySoundAction {
                    player_uuid: player_uuid.into(),
                    sound: sound.into(),
                    position: position.map(|v| v.into()),
                    volume: volume.map(|v| v.into()),
                    pitch: pitch.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `ExecuteCommand` action to the server.
    pub async fn execute_command(
        &self,
        player_uuid: String,
        command: String,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::ExecuteCommand(types::ExecuteCommandAction {
                    player_uuid: player_uuid.into(),
                    command: command.into(),
                }),
            )
            .await
    }
    ///Sends a `WorldSetDefaultGameMode` action to the server.
    pub async fn world_set_default_game_mode(
        &self,
        world: Option<types::WorldRef>,
        game_mode: i32,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::WorldSetDefaultGameMode(types::WorldSetDefaultGameModeAction {
                    world: world.map(|v| v.into()),
                    game_mode: game_mode.into(),
                }),
            )
            .await
    }
    ///Sends a `WorldSetDifficulty` action to the server.
    pub async fn world_set_difficulty(
        &self,
        world: Option<types::WorldRef>,
        difficulty: i32,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::WorldSetDifficulty(types::WorldSetDifficultyAction {
                    world: world.map(|v| v.into()),
                    difficulty: difficulty.into(),
                }),
            )
            .await
    }
    ///Sends a `WorldSetTickRange` action to the server.
    pub async fn world_set_tick_range(
        &self,
        world: Option<types::WorldRef>,
        tick_range: i32,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::WorldSetTickRange(types::WorldSetTickRangeAction {
                    world: world.map(|v| v.into()),
                    tick_range: tick_range.into(),
                }),
            )
            .await
    }
    ///Sends a `WorldSetBlock` action to the server.
    pub async fn world_set_block(
        &self,
        world: Option<types::WorldRef>,
        position: Option<types::BlockPos>,
        block: Option<types::BlockState>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::WorldSetBlock(types::WorldSetBlockAction {
                    world: world.map(|v| v.into()),
                    position: position.map(|v| v.into()),
                    block: block.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `WorldPlaySound` action to the server.
    pub async fn world_play_sound(
        &self,
        world: Option<types::WorldRef>,
        sound: i32,
        position: Option<types::Vec3>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::WorldPlaySound(types::WorldPlaySoundAction {
                    world: world.map(|v| v.into()),
                    sound: sound.into(),
                    position: position.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `WorldAddParticle` action to the server.
    pub async fn world_add_particle(
        &self,
        world: Option<types::WorldRef>,
        position: Option<types::Vec3>,
        particle: i32,
        block: Option<types::BlockState>,
        face: Option<i32>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::WorldAddParticle(types::WorldAddParticleAction {
                    world: world.map(|v| v.into()),
                    position: position.map(|v| v.into()),
                    particle: particle.into(),
                    block: block.map(|v| v.into()),
                    face: face.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `WorldQueryEntities` action to the server.
    pub async fn world_query_entities(
        &self,
        world: Option<types::WorldRef>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::WorldQueryEntities(types::WorldQueryEntitiesAction {
                    world: world.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `WorldQueryPlayers` action to the server.
    pub async fn world_query_players(
        &self,
        world: Option<types::WorldRef>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::WorldQueryPlayers(types::WorldQueryPlayersAction {
                    world: world.map(|v| v.into()),
                }),
            )
            .await
    }
    ///Sends a `WorldQueryEntitiesWithin` action to the server.
    pub async fn world_query_entities_within(
        &self,
        world: Option<types::WorldRef>,
        r#box: Option<types::BBox>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.send_action(
                types::action::Kind::WorldQueryEntitiesWithin(types::WorldQueryEntitiesWithinAction {
                    world: world.map(|v| v.into()),
                    r#box: r#box.map(|v| v.into()),
                }),
            )
            .await
    }
}
