import { PluginBase } from '../plugin/PluginBase.js';
import { CommandSpec, ParamSpec } from '../generated/command.js';
import { COMMAND_HANDLERS, CommandOptions } from './decorators.js';
import { EventType } from '../generated/plugin.js';
import { EventContext } from '../events/EventManager.js';

export class CommandManager {
    private handlers: Map<string, Function> = new Map();
    private definitions: CommandSpec[] = [];

    constructor(private plugin: PluginBase) {
        this.registerPluginCommands();
        this.plugin.eventManager.registerHandler(EventType.COMMAND, this.handleCommandEvent.bind(this));
    }

    private registerPluginCommands() {
        const proto = Object.getPrototypeOf(this.plugin);
        const commands: { method: string; options: CommandOptions }[] = proto[COMMAND_HANDLERS] || [];
        
        for (const cmd of commands) {
            const name = cmd.options.name.toLowerCase();
            this.handlers.set(name, (this.plugin as any)[cmd.method].bind(this.plugin));
            
            this.definitions.push({
                name: '/' + cmd.options.name,
                description: cmd.options.description || '',
                aliases: cmd.options.aliases || [],
                params: cmd.options.params ? cmd.options.params.map(p => ParamSpec.fromPartial(p)) : []
            });
        }
    }

    public getCommandDefinitions(): CommandSpec[] {
        return this.definitions;
    }

    private handleCommandEvent(data: any, context: EventContext<any>) {
         const cmdName = data.command.toLowerCase();
         const handler = this.handlers.get(cmdName);
         
         if (handler) {
             handler(data.playerUuid, data.args, context);
         }
    }
}
