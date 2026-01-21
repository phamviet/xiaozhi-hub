import { useState, useEffect } from "react"
import { pb } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { MessageSquare, User, Bot } from "lucide-react"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"
import type { AIAgent } from "@/components/agents/types"
import { AudioPlayer } from "./AudioPlayer"
import type { AIAgentChat, ChatMessage } from "@/components/chat/types.ts"

interface Conversation extends AIAgentChat {
	messages: ChatMessage[]
}

export interface ExpandedChatMessage extends ChatMessage {
	expand?: {
		chat: Conversation
	}
}

interface ChatHistoryDialogProps {
	agent: AIAgent
}

export function ChatHistoryDialog({ agent }: ChatHistoryDialogProps) {
	const [open, setOpen] = useState(false)
	const [conversations, setConversations] = useState<Conversation[]>([])
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
			// Fetch ai_agent_chat records for the current agent
			// Expand 'ai_agent_chat_history(chat)' to get all messages for each chat
			// Sort by created date descending for the chat sessions
			const chatRecords = await pb.collection("ai_agent_chat").getList<AIAgentChat>(1, 500, {
				filter: `agent = "${agent.id}"`,
				sort: "-created",
				expand: "ai_agent_chat_history(chat)", // This expands the related messages
			})

			const fetchedConversations: Conversation[] = chatRecords.items.map((chat) => {
				const messages = (chat.expand?.["ai_agent_chat_history(chat)"] || []) as ExpandedChatMessage[]
				// Sort messages within a conversation by created time ascending
				messages.sort((a, b) => new Date(a.created).getTime() - new Date(b.created).getTime())
				return {
					...chat,
					messages: messages,
				}
			})

			setConversations(fetchedConversations)

			// Select the first conversation by default if none is selected
			if (!selectedConversationId && fetchedConversations.length > 0) {
				setSelectedConversationId(fetchedConversations[0].id)
			}
		} catch (err) {
			console.error("Error fetching chat history:", err)
		} finally {
			setLoading(false)
		}
	}

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
						{loading && conversations.length === 0 ? (
							<div className="flex justify-center p-4">
								<Spinner className="size-6 text-muted-foreground" />
							</div>
						) : conversations.length === 0 ? (
							<div className="p-4 text-center text-muted-foreground text-sm">No conversations found.</div>
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
											{new Date(conv.ended || conv.created).toLocaleString()}
										</span>
										<span className="text-xs text-muted-foreground truncate w-full mt-1">
											{conv.summary || conv.messages[0]?.content || "No summary"}
										</span>
										<span className="text-[10px] text-muted-foreground mt-1">{conv.messages.length} messages</span>
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
										className={cn("flex gap-3 max-w-[80%]", msg.chat_type === "1" ? "ml-auto flex-row-reverse" : "")}
									>
										<div
											className={cn(
												"flex-shrink-0 h-8 w-8 rounded-full flex items-center justify-center",
												msg.chat_type === "1" ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
											)}
										>
											{msg.chat_type === "1" ? <User className="h-4 w-4" /> : <Bot className="h-4 w-4" />}
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
