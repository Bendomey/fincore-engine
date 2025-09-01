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