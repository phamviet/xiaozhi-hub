import { useState, useEffect, useMemo } from "react"
import { pb } from "@/lib/api"
import { Button } from "@/components/ui/button"
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog"
import { MessageSquare, User, Bot } from "lucide-react"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"
import type { AIAgent } from "@/components/agents/types"
import type { ChatMessage, Conversation } from "./types"
import { AudioPlayer } from "./AudioPlayer"

interface ChatHistoryDialogProps {
	agent: AIAgent
}

export function ChatHistoryDialog({ agent }: ChatHistoryDialogProps) {
	const [open, setOpen] = useState(false)
	const [messages, setMessages] = useState<ChatMessage[]>([])
	const [loading, setLoading] = useState(false)
	const [selectedConversationId, setSelectedConversationId] = useState<string | null>(null)

	useEffect(() => {
		if (open) {
			fetchHistory()
		}
	}, [open])

	const fetchHistory = async () => {
		setLoading(true)
		try {
			// Fetch all history for this agent
			// Note: In a real app with lots of history, we might want to paginate or fetch conversations first.
			// But for now, we'll fetch recent history (e.g., last 500 messages)
			const records = await pb.collection("ai_agent_chat_history").getList<ChatMessage>(1, 500, {
				filter: `agent = "${agent.id}"`,
				sort: "-created",
			})
			setMessages(records.items)
		} catch (err) {
			console.error("Error fetching chat history:", err)
		} finally {
			setLoading(false)
		}
	}

	const conversations = useMemo(() => {
		const groups = new Map<string, ChatMessage[]>()

		messages.forEach((msg) => {
			const convId = msg.conversation_id || "unknown"
			if (!groups.has(convId)) {
				groups.set(convId, [])
			}
			groups.get(convId)?.push(msg)
		})

		const convs: Conversation[] = []
		groups.forEach((msgs, id) => {
			// Sort messages by created time ascending
			msgs.sort((a, b) => new Date(a.created).getTime() - new Date(b.created).getTime())

			convs.push({
				id,
				messages: msgs,
				lastMessageTime: msgs[msgs.length - 1].created,
			})
		})

		// Sort conversations by last message time descending
		convs.sort((a, b) => new Date(b.lastMessageTime).getTime() - new Date(a.lastMessageTime).getTime())

		return convs
	}, [messages])

	// Select first conversation by default if none selected
	useEffect(() => {
		if (!selectedConversationId && conversations.length > 0) {
			setSelectedConversationId(conversations[0].id)
		}
	}, [conversations, selectedConversationId])

	const selectedConversation = conversations.find((c) => c.id === selectedConversationId)

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button variant="outline" size="sm">
					<MessageSquare className="h-4 w-4" />
					Chat History
				</Button>
			</DialogTrigger>
			<DialogContent className="max-w-4xl h-[80vh] flex flex-col p-0 gap-0">
				<DialogHeader className="p-6 pb-2">
					<DialogTitle>Chat History - {agent.agent_name}</DialogTitle>
				</DialogHeader>

				<div className="flex flex-1 overflow-hidden border-t mt-2">
					{/* Left Sidebar: Conversation List */}
					<div className="w-1/3 border-r overflow-y-auto bg-muted/10">
						{loading && messages.length === 0 ? (
							<div className="flex justify-center p-4">
								<Spinner className="size-6 text-muted-foreground" />
							</div>
						) : conversations.length === 0 ? (
							<div className="p-4 text-center text-muted-foreground text-sm">
								No conversations found.
							</div>
						) : (
							<div className="flex flex-col">
								{conversations.map((conv) => (
									<button
										key={conv.id}
										onClick={() => setSelectedConversationId(conv.id)}
										className={cn(
											"flex flex-col items-start p-4 border-b text-left hover:bg-muted/50 transition-colors",
											selectedConversationId === conv.id && "bg-muted"
										)}
									>
										<span className="font-medium text-sm truncate w-full">
											{new Date(conv.lastMessageTime).toLocaleString()}
										</span>
										<span className="text-xs text-muted-foreground truncate w-full mt-1">
											{conv.messages[conv.messages.length - 1].content}
										</span>
										<span className="text-[10px] text-muted-foreground mt-1">
											{conv.messages.length} messages
										</span>
									</button>
								))}
							</div>
						)}
					</div>

					{/* Right Content: Chat Messages */}
					<div className="flex-1 flex flex-col overflow-hidden bg-background">
						{selectedConversation ? (
							<div className="flex-1 overflow-y-auto p-4 space-y-4">
								{selectedConversation.messages.map((msg) => (
									<div
										key={msg.id}
										className={cn(
											"flex gap-3 max-w-[80%]",
											msg.chat_type === "1" ? "ml-auto flex-row-reverse" : ""
										)}
									>
										<div
											className={cn(
												"flex-shrink-0 h-8 w-8 rounded-full flex items-center justify-center",
												msg.chat_type === "1"
													? "bg-primary text-primary-foreground"
													: "bg-muted text-muted-foreground"
											)}
										>
											{msg.chat_type === "1" ? (
												<User className="h-4 w-4" />
											) : (
												<Bot className="h-4 w-4" />
											)}
										</div>
										<div
											className={cn(
												"flex flex-col gap-1 p-3 rounded-lg text-sm",
												msg.chat_type === "1"
													? "bg-primary text-primary-foreground rounded-tr-none"
													: "bg-muted text-foreground rounded-tl-none"
											)}
										>
											<div className="flex items-center gap-2">
												<p>{msg.content}</p>
												{msg.chat_audio && <AudioPlayer message={msg} />}
											</div>
											<span className="text-[10px] opacity-70 self-end">
												{new Date(msg.created).toLocaleTimeString()}
											</span>
										</div>
									</div>
								))}
							</div>
						) : (
							<div className="flex-1 flex items-center justify-center text-muted-foreground">
								Select a conversation to view details
							</div>
						)}
					</div>
				</div>
			</DialogContent>
		</Dialog>
	)
}
