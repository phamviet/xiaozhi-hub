import { memo, useMemo } from "react"

export default memo(() => {
	document.title = "Home page"

	return useMemo(
		() => (
			<>
				<div className="grid gap-4">
					Welcome!
				</div>
			</>
		),
		[]
	)
})
