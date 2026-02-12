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