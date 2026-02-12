install-tools:
	go install github.com/cespare/reflex@v0.3.1
	go install github.com/swaggo/swag/cmd/swag@v1.16.6
	go install mvdan.cc/gofumpt@v0.9.2
	go install github.com/segmentio/golines@v0.13.0

run:
	scripts/run.sh

run-dev:
	scripts/run-dev.sh

build-server:
	scripts/build.sh

setup-db:
	go run init/main.go init/setup.go -init true

update-db:
	go run init/main.go init/setup.go -init false

deploy-staging:
	fly deploy --config fly.staging.toml --remote-only

lint:
	~/go/bin/gofumpt -l -d .
	~/go/bin/golines -m 120 -d .

lint-fix:
	~/go/bin/gofumpt -l -w .
	~/go/bin/golines -m 120 -w .
	~/go/bin/swag fmt