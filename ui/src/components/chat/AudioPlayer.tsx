import { useState, useRef } from "react"
import { pb } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Play, Pause } from "lucide-react"
import type { ChatMessage } from "./types"

interface AudioPlayerProps {
	message: ChatMessage
}

export function AudioPlayer({ message }: AudioPlayerProps) {
	const audioRef = useRef<HTMLAudioElement>(null)
	const [isPlaying, setIsPlaying] = useState(false)

	if (!message.chat_audio) {
		return null
	}

	const audioUrl = pb.files.getURL(message, message.chat_audio)

	const togglePlay = () => {
		if (audioRef.current) {
			if (isPlaying) {
				audioRef.current.pause()
			} else {
				audioRef.current.play()
			}
		}
	}

	return (
		<div className="flex items-center">
			<audio
				ref={audioRef}
				src={audioUrl}
				onPlay={() => setIsPlaying(true)}
				onPause={() => setIsPlaying(false)}
				onEnded={() => setIsPlaying(false)}
				className="hidden"
			/>
			<Button
				onClick={togglePlay}
				size="icon"
				variant="ghost"
				className="h-6 w-6 rounded-full hover:bg-muted-foreground/20"
				title="Play audio"
			>
				{isPlaying ? <Pause className="h-3 w-3" /> : <Play className="h-3 w-3" />}
			</Button>
		</div>
	)
}
