import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs))
}

/** Adds event listener to node and returns function that removes the listener */
export function listen<T extends Event = Event>(node: Node, event: string, handler: (event: T) => void) {
	node.addEventListener(event, handler as EventListener)
	return () => node.removeEventListener(event, handler as EventListener)
}

// biome-ignore lint/suspicious/noExplicitAny: any is used to allow any function to be passed in
export function debounce<T extends (...args: any[]) => any>(func: T, wait: number): (...args: Parameters<T>) => void {
	let timeout: ReturnType<typeof setTimeout>
	return (...args: Parameters<T>) => {
		clearTimeout(timeout)
		timeout = setTimeout(() => func(...args), wait)
	}
}

// Cache for runOnce
// biome-ignore lint/complexity/noBannedTypes: Function is used to allow any function to be passed in
const runOnceCache = new WeakMap<Function, { done: boolean; result: unknown }>()
/** Run a function only once */
// biome-ignore lint/suspicious/noExplicitAny: any is used to allow any function to be passed in
export function runOnce<T extends (...args: any[]) => any>(fn: T): T {
	return ((...args: Parameters<T>) => {
		let state = runOnceCache.get(fn)
		if (!state) {
			state = { done: false, result: undefined }
			runOnceCache.set(fn, state)
		}
		if (!state.done) {
			state.result = fn(...args)
			state.done = true
		}
		return state.result
	}) as T
}
