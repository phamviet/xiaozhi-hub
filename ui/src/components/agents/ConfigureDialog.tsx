import type { AIAgent } from "@/components/agents/types.ts"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog.tsx"
import { Button } from "@/components/ui/button.tsx"
import { Settings } from "lucide-react"
import { useState } from "react"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { pb } from "@/lib/api"
import { toast } from "sonner"
import { Spinner } from "@/components/ui/spinner"

function ConfigureDialog({ agent }: { agent: AIAgent }) {
	const [open, setOpen] = useState(false)
	const [rolePrompt, setRolePrompt] = useState(agent.role_prompt)
	const [loading, setLoading] = useState(false)

	const handleSave = async () => {
		setLoading(true)
		try {
			await pb.collection("ai_agent").update(agent.id, {
				role_prompt: rolePrompt,
			})
			toast.success("Character updated successfully")
			setOpen(false)
			// Optionally trigger a refresh if needed, but realtime updates might handle it
		} catch (error) {
			console.error(error)
			toast.error("Failed to update character")
		} finally {
			setLoading(false)
		}
	}

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button variant="outline" size="sm">
					<Settings className="h-4 w-4" />
					Configure Role
				</Button>
			</DialogTrigger>
			<DialogContent className="max-w-2xl">
				<DialogHeader>
					<DialogTitle>Configure {agent.agent_name}</DialogTitle>
				</DialogHeader>
				<div className="grid gap-4 py-4">
					<div className="grid gap-2">
						<Label htmlFor="role_prompt">Character</Label>
						<Textarea
							id="role_prompt"
							value={rolePrompt}
							onChange={(e) => setRolePrompt(e.target.value)}
							placeholder="Enter the system prompt for this agent..."
							className="min-h-100"
						/>
					</div>
					<div className="flex justify-end gap-2">
						<Button variant="outline" onClick={() => setOpen(false)} disabled={loading}>
							Cancel
						</Button>
						<Button onClick={handleSave} disabled={loading}>
							{loading && <Spinner className="mr-2 h-4 w-4 animate-spin" />}
							Save Changes
						</Button>
					</div>
				</div>
			</DialogContent>
		</Dialog>
	)
}

export default ConfigureDialog
