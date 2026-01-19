# Generate schema snapshot
snapshot: go run . migrate collections
build-ui:
	bun install --cwd ./ui
	bun run --cwd ./ui build

dev-ui:
	bun install --cwd ./ui
	bun run --cwd ./ui dev --host 0.0.0.0

dev-api: export ENV=dev
dev-api: export APP_URL=http://127.0.0.1:8090
dev-api:
	mkdir -p ./ui/dist && touch ./ui/dist/index.html
	air

# make -j dev
dev-all: dev-ui dev-api

dev:
	make -j dev-all

clean:
	rm -rf ui/dist

docker: clean
	docker build -t xiaozhi-hub .