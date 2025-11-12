// Example Dragonfly plugin implemented in TypeScript with generated protobuf types.
// Run: npm install && npm run dev

import * as grpc from '@grpc/grpc-js';
import {
    HostToPlugin,
    PluginToHost,
    EventType,
} from '@dragonfly/proto';
import { GameMode } from '@dragonfly/proto/common';

const pluginId = process.env.DF_PLUGIN_ID || 'typescript-plugin';
const rawAddress = process.env.DF_PLUGIN_SERVER_ADDRESS || 'unix:///tmp/dragonfly_plugin.sock';

// Ensure Unix socket paths have the unix: prefix for Node.js gRPC
let serverAddress = rawAddress;
if (rawAddress.startsWith('/') && !rawAddress.startsWith('unix:')) {
    serverAddress = 'unix://' + rawAddress;
}

const API_VERSION = 'v1';

// Helper function to send a message to a player
function sendMessage(
    call: grpc.ClientDuplexStream<PluginToHost, HostToPlugin>,
    targetUuid: string,
    message: string,
    correlationId?: string
) {
    const response: PluginToHost = {
        pluginId,
        actions: {
            actions: [
                {
                    correlationId: correlationId || `msg-${Date.now()}`,
                    sendChat: {
                        targetUuid,
                        message,
                    },
                },
            ],
        },
    };
    call.write(response);
}

/**
 * IMPORTANT: All events MUST receive an eventResult response to avoid timeout warnings.
 * Even if your plugin doesn't modify or cancel an event, send an acknowledgment with cancel: false.
 */

function handleEvent(
    call: grpc.ClientDuplexStream<PluginToHost, HostToPlugin>,
    event: NonNullable<HostToPlugin['event']>
) {
    switch (event.type) {
        case EventType.PLAYER_JOIN: {
            const player = event.playerJoin;
            if (!player) break;

            console.log(`[ts] player joined ${player.name} (${player.playerUuid})`);

            // Use helper to send welcome message
            sendMessage(
                call,
                player.playerUuid,
                `§aWelcome to the server, §e${player.name}§a! (from TypeScript)`,
                `join-${player.playerUuid}`
            );

            // Acknowledge the event
            const ackResponse: PluginToHost = {
                pluginId,
                eventResult: {
                    eventId: event.eventId,
                    cancel: false,
                },
            };
            call.write(ackResponse);
            break;
        }

        case EventType.PLAYER_QUIT: {
            const player = event.playerQuit;
            if (!player) break;
            console.log(`[ts] player left ${player.name}`);

            // Acknowledge the event
            const ackResponse: PluginToHost = {
                pluginId,
                eventResult: {
                    eventId: event.eventId,
                    cancel: false,
                },
            };
            call.write(ackResponse);
            break;
        }

        case EventType.COMMAND: {
            const cmd = event.command;
            if (!cmd) break;

            // Now we get structured command name and args instead of parsing raw string!
            console.log(`[ts] command: ${cmd.command}, args:`, cmd.args);

            // Handle /greet command
            if (cmd.command === 'greet') {
                sendMessage(
                    call,
                    cmd.playerUuid,
                    `§6Hello §b${cmd.name}§6! This is a TypeScript plugin with full type safety! 🚀`
                );
                return;
            }

            // Handle /tp command with optional coordinates
            if (cmd.command === 'tp') {
                let x = 0, y = 100, z = 0;

                // Parse coordinates from args if provided: /tp <x> <y> <z>
                if (cmd.args && cmd.args.length === 3) {
                    x = parseFloat(cmd.args[0]) || 0;
                    y = parseFloat(cmd.args[1]) || 100;
                    z = parseFloat(cmd.args[2]) || 0;
                }

                const response: PluginToHost = {
                    pluginId,
                    actions: {
                        actions: [
                            {
                                correlationId: `tp-${Date.now()}`,
                                teleport: {
                                    playerUuid: cmd.playerUuid,
                                    x,
                                    y,
                                    z,
                                    yaw: 0,
                                    pitch: 0,
                                },
                            },
                            {
                                correlationId: `tp-msg-${Date.now()}`,
                                sendChat: {
                                    targetUuid: cmd.playerUuid,
                                    message: `§aTeleported to ${x}, ${y}, ${z}!`,
                                },
                            },
                        ],
                    },
                };
                call.write(response);
                return;
            }

            // Handle /gm command to change game mode
            if (cmd.command === 'gamemode') {
                let gameMode: GameMode;
                let modeName: string;
                if (!cmd.args || cmd.args.length === 0) {
                    gameMode = GameMode.SURVIVAL;
                    modeName = 'Survival';
                    // sendMessage(call, cmd.playerUuid, '§cUsage: /gm <survival|creative|adventure|spectator>');
                    // return;
                }

                const mode = cmd.args[0].toLowerCase();

                switch (mode) {
                    case 'survival':
                    case 's':
                    case '0':
                        gameMode = GameMode.SURVIVAL;
                        modeName = 'Survival';
                        break;
                    case 'creative':
                    case 'c':
                    case '1':
                        gameMode = GameMode.CREATIVE;
                        modeName = 'Creative';
                        break;
                    case 'adventure':
                    case 'a':
                    case '2':
                        gameMode = GameMode.ADVENTURE;
                        modeName = 'Adventure';
                        break;
                    case 'spectator':
                    case 'sp':
                    case '3':
                        gameMode = GameMode.SPECTATOR;
                        modeName = 'Spectator';
                        break;
                    default:
                        sendMessage(
                            call,
                            cmd.playerUuid,
                            '§cInvalid game mode. Use: survival, creative, adventure, or spectator'
                        );
                        return;
                }

                // Use action batch for multiple actions
                const response: PluginToHost = {
                    pluginId,
                    actions: {
                        actions: [
                            {
                                correlationId: `gm-${Date.now()}`,
                                setGameMode: {
                                    playerUuid: cmd.playerUuid,
                                    gameMode: gameMode,
                                },
                            },
                        ],
                    },
                };
                call.write(response);

                // Send success message
                sendMessage(call, cmd.playerUuid, `§aGame mode changed to §e${modeName}§a!`);
                return;
            }

            // For commands we don't handle, send an acknowledgment to avoid timeout warnings
            // This allows the command to pass through to other handlers
            const ackResponse: PluginToHost = {
                pluginId,
                eventResult: {
                    eventId: event.eventId,
                    cancel: false,
                },
            };
            call.write(ackResponse);
            break;
        }

        case EventType.CHAT: {
            const chat = event.chat;
            if (!chat) break;

            // Profanity filter with event cancellation
            const badWords = ['badword', 'spam', 'hack', 'fuck'];
            if (badWords.some(word => chat.message.toLowerCase().includes(word))) {
                const cancelResponse: PluginToHost = {
                    pluginId,
                    eventResult: {
                        eventId: event.eventId,
                        cancel: true,
                    },
                };
                call.write(cancelResponse);

                sendMessage(call, chat.playerUuid, '§cPlease keep the chat friendly');
                break;
            }

            // Chat mutation example
            if (chat.message.startsWith('!shout ')) {
                const updated = chat.message.substring(7).toUpperCase() + '!!!';
                const mutateResponse: PluginToHost = {
                    pluginId,
                    eventResult: {
                        eventId: event.eventId,
                        cancel: false,
                        chat: { message: updated },
                    },
                };
                call.write(mutateResponse);
                break;
            }

            // Rainbow text easter egg
            if (chat.message.startsWith('!rainbow ')) {
                const text = chat.message.substring(9);
                const colors = ['§c', '§6', '§e', '§a', '§b', '§d'];
                const rainbow = text.split('').map((char, i) => colors[i % colors.length] + char).join('');

                const mutateResponse: PluginToHost = {
                    pluginId,
                    eventResult: {
                        eventId: event.eventId,
                        chat: { message: rainbow },
                    },
                };
                call.write(mutateResponse);
                break;
            }

            // Acknowledge regular chat messages
            const ackResponse: PluginToHost = {
                pluginId,
                eventResult: {
                    eventId: event.eventId,
                    cancel: false,
                },
            };
            call.write(ackResponse);
            break;
        }

        case EventType.PLAYER_BLOCK_BREAK: {
            const blockBreak = event.blockBreak;
            if (!blockBreak || !blockBreak.position) break;

            const { x, y, z } = blockBreak.position;
            console.log(`[ts] ${blockBreak.name} broke block at ${x},${y},${z}`);

            // Example: Double drops for diamond ore
            if (x % 10 === 0) { // Just as an example
                const response: PluginToHost = {
                    pluginId,
                    eventResult: {
                        eventId: event.eventId,
                        blockBreak: {
                            drops: {
                                items: [
                                    { name: 'minecraft:diamond', count: 2, meta: 0 },
                                ],
                            },
                            xp: 10,
                        }
                    },
                };
                call.write(response);
            } else {
                // Acknowledge the event even if we don't modify it
                const ackResponse: PluginToHost = {
                    pluginId,
                    eventResult: {
                        eventId: event.eventId,
                        cancel: false,
                    },
                };
                call.write(ackResponse);
            }
            break;
        }

        default:
            console.log('[ts] unhandled event type:', EventType[event.type] ?? event.type);
    }
}

// Connect to Dragonfly as a gRPC client
const client = new grpc.Client(
    serverAddress,
    grpc.credentials.createInsecure()
);

// Create bidirectional stream
const call = client.makeBidiStreamRequest<PluginToHost, HostToPlugin>(
    '/df.plugin.Plugin/EventStream',
    (msg: PluginToHost) => {
        const writer = PluginToHost.encode(msg);
        return Buffer.from(writer.finish());
    },
    (buf: Buffer) => {
        return HostToPlugin.decode(new Uint8Array(buf));
    }
) as grpc.ClientDuplexStream<PluginToHost, HostToPlugin>;

console.log(`[ts] connecting to ${serverAddress}...`);

// Send initial handshake
const helloMessage: PluginToHost = {
    pluginId,
    hello: {
        name: 'example-typescript',
        version: '0.1.0',
        apiVersion: API_VERSION,
        commands: [
            { name: '/greet', description: 'Send a greeting from the TypeScript plugin', aliases: [] },
            { name: '/tp', description: 'Teleport to spawn', aliases: [] },
            { name: '/gamemode', description: 'Change game mode (survival, creative, adventure, spectator)', aliases: ['gm'] },
        ],
    },
};
call.write(helloMessage);

const initialSubscribe: PluginToHost = {
    pluginId,
    subscribe: {
        events: [
            EventType.PLAYER_JOIN,
            EventType.PLAYER_QUIT,
            EventType.COMMAND,
            EventType.CHAT,
            EventType.PLAYER_BLOCK_BREAK,
        ],
    },
};
call.write(initialSubscribe);

// Handle incoming messages from server
call.on('data', (message: HostToPlugin) => {
    // Handle hello handshake
    if (message.hello) {
        console.log('[ts] host hello', message.hello);
        if (message.hello.apiVersion !== API_VERSION) {
            console.warn(`[ts] API version mismatch: host=${message.hello.apiVersion}, plugin=${API_VERSION}`);
        }
        return;
    }

    // Handle events
    if (message.event) {
        handleEvent(call, message.event);
    }

    // Handle shutdown
    if (message.shutdown) {
        console.log('[ts] host shutdown:', message.shutdown.reason);
        call.end();
    }
});

call.on('end', () => {
    console.log('[ts] stream ended');
    process.exit(0);
});

call.on('error', (err) => {
    console.error('[ts] stream error:', err);
    process.exit(1);
});

// Graceful shutdown
process.on('SIGINT', () => {
    console.log('[ts] shutting down...');
    call.end();
});

