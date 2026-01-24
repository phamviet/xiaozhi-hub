import { useEffect, useState } from "react"
import { pb } from "@/lib/api"
import type { AIAgent } from "./types"
import { AgentCard } from "./AgentCard"
import { Spinner } from "@/components/ui/spinner"

export function AgentList() {
	const [agents, setAgents] = useState<AIAgent[]>([])
	const [loading, setLoading] = useState(true)
	const [error, setError] = useState<string | null>(null)

	useEffect(() => {
		const fetchAgents = async () => {
			try {
				const records = await pb.collection("ai_agent").getList<AIAgent>(1, 50, {
					sort: "-created",
				})
				setAgents(records.items)
			} catch (err: any) {
				console.error("Error fetching agents:", err)
				setError("Failed to load agents.")
			} finally {
				setLoading(false)
			}
		}

		fetchAgents()

		// Subscribe to realtime updates
		pb.collection("ai_agent").subscribe<AIAgent>("*", (e) => {
			if (e.action === "create") {
				setAgents((prev) => [e.record, ...prev])
			} else if (e.action === "update") {
				setAgents((prev) => prev.map((agent) => (agent.id === e.record.id ? e.record : agent)))
			} else if (e.action === "delete") {
				setAgents((prev) => prev.filter((agent) => agent.id !== e.record.id))
			}
		})

		return () => {
			pb.collection("ai_agent").unsubscribe("*")
		}
	}, [])

	if (loading) {
		return (
			<div className="flex justify-center p-8">
				<Spinner className="size-8 text-muted-foreground" />
			</div>
		)
	}

	if (error) {
		return <div className="text-destructive p-4">{error}</div>
	}

	if (agents.length === 0) {
		return <div className="text-center p-8 text-muted-foreground">No agents found. Create one to get started.</div>
	}

	return (
		<div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
			{agents.map((agent) => (
				<AgentCard key={agent.id} agent={agent} />
			))}
		</div>
	)
}
