import type { RecordModel } from "pocketbase"

export interface ChatMessage extends RecordModel {
	agent: string
	device: string
	mac_address: string
	conversation_id: string
	content: string
	chat_type: "1" | "2" // 1: User, 2: Assistant
	chat_audio: string // filename
}

export interface Conversation {
	id: string
	messages: ChatMessage[]
	lastMessageTime: string
}
