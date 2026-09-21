const bus = new EventTarget()
export function getGlobalEventBus() { return { subscribe: (name: string, handler: EventListener) => bus.addEventListener(name, handler), unsubscribe: (name: string, handler: EventListener) => bus.removeEventListener(name, handler), dispatch: (name: string, detail?: unknown) => bus.dispatchEvent(new CustomEvent(name, { detail })) } }
