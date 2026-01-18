import { atom } from "nanostores"
import { pb } from "./api"

/** Store if user is authenticated */
export const $authenticated = atom(pb.authStore.isValid)
