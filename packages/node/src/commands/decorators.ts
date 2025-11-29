import { ParamType } from '../generated/command.js';

export const COMMAND_HANDLERS = Symbol('COMMAND_HANDLERS');

export interface CommandParamSpec {
    name: string;
    description?: string;
    type: ParamType;
    optional: boolean;
    suffix?: string;
    enumValues?: string[];
}

export interface CommandOptions {
    name: string;
    description?: string;
    aliases?: string[];
    params?: CommandParamSpec[];
}

export function RegisterCommand(options: CommandOptions) {
    return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
        if (!target[COMMAND_HANDLERS]) {
            target[COMMAND_HANDLERS] = [];
        }
        target[COMMAND_HANDLERS].push({
            options,
            method: propertyKey
        });
    };
}
