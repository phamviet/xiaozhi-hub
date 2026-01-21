import { memo, useMemo } from "react"
import { AgentList } from "@/components/agents/AgentList"

export default memo(() => {
	document.title = "Home page"

	return useMemo(
		() => (
			<>
				<div className="grid gap-4">
					<h1 className="text-2xl font-bold tracking-tight">My Agents</h1>
					<AgentList />
				</div>
			</>
		),
		[]
	)
})
