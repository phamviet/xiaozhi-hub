import PocketBase from "pocketbase"
import { basePath } from "@/components/router"

/** PocketBase JS Client */
export const pb = new PocketBase(basePath)

export const isAdmin = () => pb.authStore.record?.role === "admin"
export const isReadOnlyUser = () => pb.authStore.record?.role === "readonly"

export const verifyAuth = () => {
	pb.collection("users")
		.authRefresh()
		.catch(() => {
			logOut()
		})
}

/** Logs the user out by clearing the auth store and unsubscribing from realtime updates. */
export function logOut() {
	sessionStorage.setItem("lo", "t") // prevent auto login on logout
	pb.authStore.clear()
	pb.realtime.unsubscribe()
}
