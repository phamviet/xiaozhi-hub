import { LoaderCircle, MailIcon, SendHorizonalIcon } from "lucide-react"
import { useCallback, useState } from "react"
import { pb } from "@/lib/api"
import { cn } from "@/lib/utils"
import { buttonVariants } from "../ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "../ui/dialog"
import { Input } from "../ui/input"
import { Label } from "../ui/label"
import { toast } from "sonner"

const showLoginFaliedToast = () => {
	toast.error("Login attempt failed", {
		description: `Please check your credentials and try again`,
	})
}

export default function ForgotPassword() {
	const [isLoading, setIsLoading] = useState<boolean>(false)
	const [email, setEmail] = useState("")

	const handleSubmit = useCallback(
		async (e: React.FormEvent<HTMLFormElement>) => {
			e.preventDefault()
			setIsLoading(true)
			try {
				// console.log(email)
				await pb.collection("users").requestPasswordReset(email)
				toast.info(`Password reset request received`, {
					description: `Check ${email} for a reset link.`,
				})
			} catch (e) {
				showLoginFaliedToast()
			} finally {
				setIsLoading(false)
				setEmail("")
			}
		},
		[email]
	)

	return (
		<>
			<form onSubmit={handleSubmit}>
				<div className="grid gap-3">
					<div className="grid gap-1 relative">
						<MailIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
						<Label className="sr-only" htmlFor="email">
							Email
						</Label>
						<Input
							value={email}
							onChange={(e) => setEmail(e.target.value)}
							id="email"
							name="email"
							required
							placeholder="name@example.com"
							type="email"
							autoCapitalize="none"
							autoComplete="email"
							autoCorrect="off"
							disabled={isLoading}
							className="ps-9"
						/>
					</div>
					<button className={cn(buttonVariants())} disabled={isLoading}>
						{isLoading ? (
							<LoaderCircle className="me-2 h-4 w-4 animate-spin" />
						) : (
							<SendHorizonalIcon className="me-2 h-4 w-4" />
						)}
						Reset Password
					</button>
				</div>
			</form>
			<Dialog>
				<DialogTrigger asChild>
					<button className="text-sm mx-auto hover:text-brand underline underline-offset-4 opacity-70 hover:opacity-100 transition-opacity">
						Command line instructions
					</button>
				</DialogTrigger>
				<DialogContent className="max-w-[41em]">
					<DialogHeader>
						<DialogTitle>Command line instructions</DialogTitle>
					</DialogHeader>
					<p className="text-primary/70 text-[0.95em] leading-relaxed">
						If you've lost the password to your admin account, you may reset it using the following command.
					</p>
					<p className="text-primary/70 text-[0.95em] leading-relaxed">
						Then log into the backend and reset your user account password in the users table.
					</p>
					<code className="bg-muted rounded-sm py-0.5 px-2.5 me-auto text-sm">
						./beszel superuser upsert user@example.com password
					</code>
					<code className="bg-muted rounded-sm py-0.5 px-2.5 me-auto text-sm">
						docker exec beszel /beszel superuser upsert name@example.com password
					</code>
				</DialogContent>
			</Dialog>
		</>
	)
}
