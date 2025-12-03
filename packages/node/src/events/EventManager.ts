import { PluginBase } from '../plugin/PluginBase.js';
import { HostToPlugin, EventType } from '../generated/plugin.js';
import { EVENT_HANDLERS } from './decorators.js';

export class EventContext<T> {
    private handled = false;

    constructor(
        public readonly event: NonNullable<HostToPlugin['event']>,
        public readonly data: T,
        private plugin: PluginBase
    ) {}

    public get client() {
        return this.plugin.getStream();
    }

    public async cancel(): Promise<void> {
        if (this.handled) return;
        this.handled = true;

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
        if (this.handled) return;
        this.handled = true;

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

    public async ackIfUnhandled(): Promise<void> {
        if (!this.handled) {
            await this.ack();
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
        
        for (const key in event) {
            if (key === 'eventId' || key === 'type' || key === 'expectsResponse' || key === 'immediate') continue;
            const val = (event as any)[key];
            if (val !== undefined) {
                data = val;
                break;
            }
        }

        const context = new EventContext(event, data, this.plugin);

        try {
            for (const handler of handlers) {
                await handler(data, context);
            }
        } finally {
            await context.ackIfUnhandled();
        }
    }
}
