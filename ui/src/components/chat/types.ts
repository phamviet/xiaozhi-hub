import type { RecordModel } from "pocketbase"

export interface ChatMessage extends RecordModel {
	id: string
	chat: string // relation ID to ai_agent_chat
	content: string
	chat_type: "1" | "2" // "1" for User, "2" for Assistant
	chat_audio?: string // file name
	created: string
	updated: string
}

export interface AIAgentChat extends RecordModel {
	id: string // UUID
	agent: string // relation ID to ai_agent
	summary?: string
	ended?: string // date string
	created: string
	updated: string
}
