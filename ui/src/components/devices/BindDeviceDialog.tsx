import { useState } from "react"
import { toast } from "sonner"
import { pb } from "@/lib/api"
import { Button } from "@/components/ui/button"
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog"
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp"
import { Plus } from "lucide-react"

interface BindDeviceDialogProps {
	agentId: string
	onSuccess?: () => void
}

export function BindDeviceDialog({ agentId, onSuccess }: BindDeviceDialogProps) {
	const [open, setOpen] = useState(false)
	const [code, setCode] = useState("")
	const [loading, setLoading] = useState(false)

	const handleBind = async () => {
		if (code.length !== 6) {
			toast.error("Please enter a valid 6-digit code")
			return
		}

		setLoading(true)
		try {
			await pb.send("/hub/api/device/bind", {
				method: "POST",
				body: {
					code,
					agentId,
				},
			})
			toast.success("Device bound successfully")
			setOpen(false)
			setCode("")
			onSuccess?.()
		} catch (err: any) {
			console.error("Bind error:", err)
			toast.error(err.message || "Failed to bind device")
		} finally {
			setLoading(false)
		}
	}

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button size="sm">
					<Plus className="mr-2 h-4 w-4" />
					Add Device
				</Button>
			</DialogTrigger>
			<DialogContent className="sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle>Bind New Device</DialogTitle>
					<DialogDescription>
						Enter the 6-digit code displayed on your device to link it to this agent.
					</DialogDescription>
				</DialogHeader>
				<div className="flex justify-center py-4">
					<InputOTP maxLength={6} value={code} onChange={(value) => setCode(value)} disabled={loading}>
						<InputOTPGroup>
							<InputOTPSlot index={0} />
							<InputOTPSlot index={1} />
							<InputOTPSlot index={2} />
							<InputOTPSlot index={3} />
							<InputOTPSlot index={4} />
							<InputOTPSlot index={5} />
						</InputOTPGroup>
					</InputOTP>
				</div>
				<DialogFooter>
					<Button onClick={handleBind} disabled={loading || code.length !== 6}>
						{loading ? "Binding..." : "Bind Device"}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	)
}
