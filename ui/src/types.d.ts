import type { RecordModel } from "pocketbase"

// global window properties
declare global {
	var APP: {
		BASE_PATH: string
		HUB_URL: string
	}
}

export interface TranscribeProfile extends RecordModel {
	name: string
	provider: "MODAL" | "RUNPOD"
}

type JobStatus = "PENDING" | "QUEUED" | "PROCESSING" | "COMPLETED" | "FAILED"

interface TranscribeJob extends RecordModel {
	id: string
	name: string
	transcript: TranscriptResult
	status: JobStatus
	created: string
	updated: string
}

type TranscriptResult = {
	text: string
	model_used: string
	// file size in bytes
	size: number
	// duration in seconds
	duration: number
	segments: [
		{
			start: number
			end: number
			text: string
		},
	]
	word_segments: [
		{
			start: number
			end: number
			score: number
			word: string
		},
	]
	classifications: [
		{
			label: string
		},
	]
}
