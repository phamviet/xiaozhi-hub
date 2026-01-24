import type { RecordModel } from "pocketbase"

export interface AIAgent extends RecordModel {
	user: string
	agent_name: string
	role_prompt: string
	lang_code: string
	asr_model_id: string
	vad_model_id: string
	llm_model_id: string
	tts_model_id: string
	tts_voice_id: string
	mem_model_id: string
	intent_model_id: string
	chat_history_enabled: boolean
}

export interface ModelConfig extends RecordModel {
	model_name: string
	model_type: string
	is_default: boolean
	is_enabled: boolean
	config_json: Record<string, any>
	provider_id: string
}

export interface AIDevice extends RecordModel {
	mac_address: string
	user: string
	agent: string
	board: string
	last_connected: string
}
