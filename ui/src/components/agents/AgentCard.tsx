import { useState, useEffect } from "react"
import { toast } from "sonner"
import { pb } from "@/lib/api"
import type { AIAgent, AIDevice } from "./types"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Settings, Smartphone } from "lucide-react"
import { BindDeviceDialog } from "@/components/devices/BindDeviceDialog"
import { ChatHistoryDialog } from "@/components/chat/ChatHistoryDialog"
import { Spinner } from "@/components/ui/spinner"

interface AgentCardProps {
	agent: AIAgent
}

export function AgentCard({ agent }: AgentCardProps) {
	return (
		<div className="rounded-xl border bg-card text-card-foreground shadow">
			<div className="flex flex-col space-y-1.5 p-6">
				<h3 className="font-semibold leading-none tracking-tight">{agent.agent_name}</h3>
				<p className="text-sm text-muted-foreground">Language: {agent.lang_code}</p>
			</div>
			<div className="p-6 pt-0">
				<p className="text-sm text-muted-foreground line-clamp-3">
					{agent.role_prompt || "No role prompt configured."}
				</p>
			</div>
			<div className="flex flex-wrap items-center p-6 pt-0 gap-2">
				<ConfigureModal agent={agent} />
				<DeviceListModal agent={agent} />
				<ChatHistoryDialog agent={agent} />
			</div>
		</div>
	)
}

function ConfigureModal({ agent }: { agent: AIAgent }) {
	return (
		<Dialog>
			<DialogTrigger asChild>
				<Button variant="outline" size="sm">
					<Settings className="h-4 w-4" />
					Configure Role
				</Button>
			</DialogTrigger>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Configure {agent.agent_name}</DialogTitle>
				</DialogHeader>
				<div className="py-4">
					<p className="text-muted-foreground">Configuration options for models will appear here.</p>
				</div>
			</DialogContent>
		</Dialog>
	)
}

function DeviceListModal({ agent }: { agent: AIAgent }) {
	const [devices, setDevices] = useState<AIDevice[]>([])
	const [loading, setLoading] = useState(false)
	const [open, setOpen] = useState(false)
	const [deviceCount, setDeviceCount] = useState<number | null>(null)

	useEffect(() => {
		// Fetch count on mount
		pb.collection("ai_device")
			.getList(1, 1, {
				filter: `agent = "${agent.id}"`,
				fields: "id",
			})
			.then((res) => {
				setDeviceCount(res.totalItems)
			})
			.catch((err) => {
				console.error("Error fetching device count:", err)
			})
	}, [agent.id])

	const fetchDevices = async () => {
		setLoading(true)
		try {
			const records = await pb.collection("ai_device").getList<AIDevice>(1, 50, {
				filter: `agent = "${agent.id}"`,
				sort: "-created",
			})
			setDevices(records.items)
			setDeviceCount(records.totalItems)
		} catch (err: any) {
			console.error("Error fetching devices:", err)
			toast.error("Failed to load devices")
		} finally {
			setLoading(false)
		}
	}

	useEffect(() => {
		if (open) {
			fetchDevices()
		}
	}, [open])

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button variant="outline" size="sm">
					<Smartphone className="h-4 w-4" />
					{deviceCount !== null ? `${deviceCount} Devices` : "Devices"}
				</Button>
			</DialogTrigger>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Devices linked to {agent.agent_name}</DialogTitle>
				</DialogHeader>
				<div className="flex justify-end mb-2">
					<BindDeviceDialog agentId={agent.id} onSuccess={fetchDevices} />
				</div>
				<div className="py-4">
					{loading ? (
						<div className="flex justify-center">
							<Spinner className="size-6 text-muted-foreground" />
						</div>
					) : devices.length === 0 ? (
						<p className="text-muted-foreground text-center">No devices linked.</p>
					) : (
						<div className="space-y-2">
							{devices.map((dev) => (
								<div key={dev.id} className="flex justify-between items-center border p-2 rounded">
									<div className="flex flex-col">
										<span className="font-medium">{dev.mac_address}</span>
										<span className="text-xs text-muted-foreground">{dev.board}</span>
									</div>
									<div className="text-xs text-muted-foreground">{new Date(dev.created).toLocaleDateString()}</div>
								</div>
							))}
						</div>
					)}
				</div>
			</DialogContent>
		</Dialog>
	)
}
