import { useStore } from "@nanostores/react"
import type { AuthMethodsList } from "pocketbase"
import { useEffect, useMemo, useState } from "react"
import { UserAuthForm } from "@/components/login/auth-form"
import { pb } from "@/lib/api"
import { ModeToggle } from "@/components/mode-toggle"
import { $router } from "@/components/router"
import { useTheme } from "@/components/theme-provider"
import ForgotPassword from "@/components/login/forgot-pass-form"
import { OtpRequestForm } from "@/components/login/otp-forms"

export default function () {
	const page = useStore($router)
	const [isFirstRun, setFirstRun] = useState(false)
	const [authMethods, setAuthMethods] = useState<AuthMethodsList>()
	const { theme } = useTheme()

	useEffect(() => {
		document.title = `Login`

		pb.send("/api/first-run", {}).then(({ firstRun }) => {
			setFirstRun(firstRun)
		})
	}, [])

	useEffect(() => {
		pb.collection("users")
			.listAuthMethods()
			.then((methods) => {
				setAuthMethods(methods)
			})
	}, [])

	const subtitle = useMemo(() => {
		if (isFirstRun) {
			return `Please create an admin account`
		} else if (page?.route === "forgot_password") {
			return `Enter email address to reset password`
		} else if (page?.route === "request_otp") {
			return `Request a one-time password`
		} else {
			return `Please sign in to your account`
		}
	}, [isFirstRun, page])

	if (!authMethods) {
		return null
	}

	return (
		<div className="min-h-svh grid items-center py-12">
			<div
				className="grid gap-5 w-full px-4 mx-auto"
				// @ts-expect-error
				style={{ maxWidth: "21.5em", "--border": theme == "light" ? "hsl(30, 8%, 70%)" : "hsl(220, 3%, 25%)" }}
			>
				<div className="absolute top-3 right-3">
					<ModeToggle />
				</div>
				<div className="text-center">
					<h1 className="mb-3">
						<span>Pocketbase</span>
					</h1>
					<p className="text-sm text-muted-foreground">{subtitle}</p>
				</div>
				{page?.route === "forgot_password" ? (
					<ForgotPassword />
				) : page?.route === "request_otp" ? (
					<OtpRequestForm />
				) : (
					<UserAuthForm isFirstRun={isFirstRun} authMethods={authMethods} />
				)}
			</div>
		</div>
	)
}
