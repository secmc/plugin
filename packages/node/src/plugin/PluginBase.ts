import * as grpc from '@grpc/grpc-js';
import { HostToPlugin, PluginToHost } from '../generated/plugin.js';
import { ActionBatch, Action } from '../generated/actions.js'; // Corrected import for Action and ActionBatch
import { CommandManager } from '../commands/CommandManager.js';
import { EventManager } from '../events/EventManager.js';
import { randomUUID } from 'crypto'; // Import randomUUID

export abstract class PluginBase {
    private client: grpc.Client;
    private stream: grpc.ClientDuplexStream<PluginToHost, HostToPlugin> | null = null;
    public readonly commandManager: CommandManager;
    public readonly eventManager: EventManager;
    public readonly pluginId: string;
    
    constructor(
        private address: string = process.env.DF_PLUGIN_SERVER_ADDRESS || 'unix:///tmp/dragonfly_plugin.sock',
        pluginId: string = process.env.DF_PLUGIN_ID || 'typescript-plugin'
    ) {
        this.pluginId = pluginId;
        if (this.address.startsWith('/') && !this.address.startsWith('unix:')) {
            this.address = 'unix://' + this.address;
        }

        this.client = new grpc.Client(this.address, grpc.credentials.createInsecure());
        this.eventManager = new EventManager(this);
        this.commandManager = new CommandManager(this);

        // Auto-run the plugin on the next tick to allow subclass initialization to complete
        setTimeout(() => this.run(), 0);
    }

    abstract onLoad(): void;
    abstract onEnable(): void;
    abstract onDisable(): void;

    public run(): void {
        this.onLoad();
        console.log(`[${this.pluginId}] Connecting to ${this.address}...`);
        
        this.stream = this.client.makeBidiStreamRequest<PluginToHost, HostToPlugin>(
            '/df.plugin.Plugin/EventStream',
            (msg: PluginToHost) => {
                const writer = PluginToHost.encode(msg);
                return Buffer.from(writer.finish());
            },
            (buf: Buffer) => {
                return HostToPlugin.decode(new Uint8Array(buf));
            }
        ) as grpc.ClientDuplexStream<PluginToHost, HostToPlugin>;

        this.stream.on('data', (message: HostToPlugin) => this.handleMessage(message));
        this.stream.on('end', () => {
            console.log(`[${this.pluginId}] Stream ended`);
            process.exit(0);
        });
        this.stream.on('error', (err) => {
            console.error(`[${this.pluginId}] Stream error:`, err);
            process.exit(1);
        });

        // Send Hello
        const commands = this.commandManager.getCommandDefinitions();
        const hello: PluginToHost = {
            pluginId: this.pluginId,
            hello: {
                name: this.pluginId,
                version: '0.1.0',
                apiVersion: 'v1',
                commands: commands,
                customItems: [],
                customBlocks: [],
            }
        };
        this.stream.write(hello);
        
        // Subscribe
        const events = this.eventManager.getSubscribedEvents();
        if (events.length > 0) {
            this.stream.write({
                pluginId: this.pluginId,
                subscribe: {
                    events: events,
                }
            });
        }

        this.onEnable();
    }

    private handleMessage(message: HostToPlugin): void {
        if (message.hello) return;
        if (message.shutdown) {
            this.onDisable();
            this.stream?.end();
            return;
        }
        if (message.event) {
            this.eventManager.handleEvent(message.event);
        }
        if (message.events) {
            for (const event of message.events.events) {
                this.eventManager.handleEvent(event);
            }
        }
    }

    public send(msg: PluginToHost): Promise<void> {
        return new Promise((resolve, reject) => {
            if (this.stream) {
                this.stream.write(msg, (err?: Error | null) => {
                    if (err) {
                        console.error(`[${this.pluginId}] Error writing to stream:`, err);
                        return reject(err);
                    }
                    resolve();
                });
            } else {
                const error = new Error(`[${this.pluginId}] specific warning: Stream not ready, cannot send message`);
                console.warn(error.message);
                reject(error);
            }
        });
    }

    public async sendAction<T extends keyof Omit<Action, 'correlationId'>>(
        actionType: T,
        actionPayload: NonNullable<Action[T]>,
        correlationId?: string
    ): Promise<void> {
        const action: Action = {
            correlationId: correlationId ?? randomUUID(),
            [actionType]: actionPayload, // Dynamically set the action payload
        };

        const msg: PluginToHost = {
            pluginId: this.pluginId,
            actions: {
                actions: [action],
            },
        };

        return this.send(msg);
    }

    public getStream(): grpc.ClientDuplexStream<PluginToHost, HostToPlugin> | null {
        return this.stream;
    }
}
