import { PluginBase } from '../plugin/PluginBase.js';
import { HostToPlugin, EventType } from '../generated/plugin.js';
import { EVENT_HANDLERS } from './decorators.js';

export class EventContext<T> {
    constructor(
        public readonly event: NonNullable<HostToPlugin['event']>,
        public readonly data: T,
        private plugin: PluginBase
    ) {}

    public get client() {
        return this.plugin.getStream();
    }

    public async cancel(): Promise<void> {
         if (this.event.expectsResponse) {
            await this.plugin.send({
                pluginId: this.plugin.pluginId,
                eventResult: {
                    eventId: this.event.eventId,
                    cancel: true
                }
            });
        }
    }
    
    public async ack(): Promise<void> {
         if (this.event.expectsResponse) {
            await this.plugin.send({
                pluginId: this.plugin.pluginId,
                eventResult: {
                    eventId: this.event.eventId,
                    cancel: false
                }
            });
        }
    }
}

export class EventManager {
    private handlers: Map<EventType, Function[]> = new Map();

    constructor(private plugin: PluginBase) {
        this.registerPluginHandlers();
    }

    private registerPluginHandlers() {
        const proto = Object.getPrototypeOf(this.plugin);
        const handlers = proto[EVENT_HANDLERS] || [];
        for (const handler of handlers) {
            this.registerHandler(handler.type, (this.plugin as any)[handler.method].bind(this.plugin));
        }
    }

    public registerHandler(type: EventType, handler: Function) {
        if (!this.handlers.has(type)) {
            this.handlers.set(type, []);
        }
        this.handlers.get(type)!.push(handler);
    }

    public getSubscribedEvents(): EventType[] {
        return Array.from(this.handlers.keys());
    }

    public async handleEvent(event: NonNullable<HostToPlugin['event']>) {
        const handlers = this.handlers.get(event.type);
        if (!handlers) return;

        let data: any = event;
        switch(event.type) {
            case EventType.PLAYER_JOIN: data = event.playerJoin; break;
            case EventType.PLAYER_QUIT: data = event.playerQuit; break;
            case EventType.PLAYER_MOVE: data = event.playerMove; break;
            case EventType.CHAT: data = event.chat; break;
            case EventType.PLAYER_BLOCK_BREAK: data = event.blockBreak; break;
            case EventType.COMMAND: data = event.command; break;
            // Add other event types here as needed
        }

        const context = new EventContext(event, data, this.plugin);

        for (const handler of handlers) {
            await handler(data, context);
        }
    }
}
