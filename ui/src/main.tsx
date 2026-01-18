import { lazy, Suspense } from "react"
import { createRoot } from "react-dom/client"
import "./index.css"
import { ThemeProvider } from "@/components/theme-provider.tsx"

import { useStore } from "@nanostores/react"
import { Toaster } from "@/components/ui/sonner"
import { $router } from "@/components/router.tsx"
import { $authenticated } from "@/lib/stores.ts"
const Home = lazy(() => import("@/routes/home.tsx"))
const LoginPage = lazy(() => import("@/routes/login.tsx"))

function App() {
	const page = useStore($router)
	if (!page) {
		return <h1 className="text-3xl text-center my-14">404</h1>
	} else if (page.route === "home") {
		return <Home />
	}
}

const Layout = () => {
	const authenticated = useStore($authenticated)
	if (!authenticated) {
		return (
			<Suspense>
				<LoginPage />
			</Suspense>
		)
	}

	return (
		<div style={{ "--container": "1500px" } as React.CSSProperties}>
			<div className="container relative">
				<App />
			</div>
		</div>
	)
}

createRoot(document.getElementById("app")!).render(
	<ThemeProvider defaultTheme="dark">
		<Layout />
		<Toaster position="top-center" />
	</ThemeProvider>
)
