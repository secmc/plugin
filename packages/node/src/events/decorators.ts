import { EventType } from '../generated/plugin.js';

export const EVENT_HANDLERS = Symbol('EVENT_HANDLERS');

export function On(type: EventType) {
    return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
        if (!target[EVENT_HANDLERS]) {
            target[EVENT_HANDLERS] = [];
        }
        target[EVENT_HANDLERS].push({
            type,
            method: propertyKey
        });
    };
}
