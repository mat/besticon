build:
	go build ./...

format:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

test_all: build test test_bench
	go test -v github.com/mat/besticon/v3/besticon/iconserver

test:
	go test -v github.com/mat/besticon/v3/ico
	go test -v github.com/mat/besticon/v3/besticon
	go test -v github.com/mat/besticon/v3/besticon/iconserver
	go test -v github.com/mat/besticon/v3/lettericon
	go test -v github.com/mat/besticon/v3/colorfinder

test_race:
	go test -v -race github.com/mat/besticon/v3/ico
	go test -v -race github.com/mat/besticon/v3/besticon
	go test -v -race github.com/mat/besticon/v3/besticon/iconserver
	go test -v -race github.com/mat/besticon/v3/lettericon
	go test -v -race github.com/mat/besticon/v3/colorfinder

test_bench:
	go test github.com/mat/besticon/v3/lettericon -bench .
	go test github.com/mat/besticon/v3/colorfinder -bench .

deploy:
	git push heroku master
	heroku config:set DEPLOYED_AT=`date +%s`

install:
	go install ./...

RUN_SERVER_PORT ?= 3000
RUN_SERVER_SERVE_ASSETS_FROM_DISK ?= true
DEMO_SITES ?= bing.com,github.com,reddit.com

run_server:
	go build -tags netgo -ldflags '-s -w' -o bin/iconserver github.com/mat/besticon/v3/besticon/iconserver
	PORT=$(RUN_SERVER_PORT) DEPLOYED_AT=`date +%s` HOST_ONLY_DOMAINS=* DEMO_SITES="$(DEMO_SITES)" SERVE_ASSETS_FROM_DISK=$(RUN_SERVER_SERVE_ASSETS_FROM_DISK) ./bin/iconserver

coverage_besticon:
	go test -coverprofile=coverage.out -covermode=count github.com/mat/besticon/v3/besticon && go tool cover -html=coverage.out && unlink coverage.out

coverage_ico:
	go test -coverprofile=coverage.out -covermode=count github.com/mat/besticon/v3/ico && go tool cover -html=coverage.out && unlink coverage.out

coverage_iconserver:
	go test -coverprofile=coverage.out -covermode=count github.com/mat/besticon/v3/besticon/iconserver && go tool cover -html=coverage.out && unlink coverage.out

test_websites:
	go mod download
	cat besticon/testdata/websites.txt | xargs -P 10 -n 1  besticon

minify_css:
	curl -X POST -s --data-urlencode 'input@besticon/iconserver/assets/main.css' http://cssminifier.com/raw > besticon/iconserver/assets/main-min.css

gotags:
	gotags -tag-relative=true -R=true -sort=true -f="tags" -fields=+l .

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint is required for 'make lint' but is not installed."; \
		echo "Install it first: https://golangci-lint.run"; \
		echo "For Homebrew: brew install golangci-lint"; \
		exit 1; \
	}
	golangci-lint run ./...

#
## Building ##
#

clean:
	rm -rf bin/*
	rm -f iconserver*.zip

build_darwin_amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/darwin_amd64/iconserver -ldflags "-s -w -X github.com/mat/besticon/v3/besticon.BuildDate=`date +'%Y-%m-%d'`" github.com/mat/besticon/v3/besticon/iconserver

build_linux_amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/linux_amd64/iconserver -ldflags "-s -w -X github.com/mat/besticon/v3/besticon.BuildDate=`date +'%Y-%m-%d'`" github.com/mat/besticon/v3/besticon/iconserver

build_linux_arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/linux_arm64/iconserver -ldflags "-s -w -X github.com/mat/besticon/v3/besticon.BuildDate=`date +'%Y-%m-%d'`" github.com/mat/besticon/v3/besticon/iconserver

build_windows_amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/windows_amd64/iconserver.exe -ldflags "-s -w -X github.com/mat/besticon/v3/besticon.BuildDate=`date +'%Y-%m-%d'`" github.com/mat/besticon/v3/besticon/iconserver

build_all_platforms: build_darwin_amd64 build_linux_amd64 build_linux_arm64 build_windows_amd64
	find bin/ -type file | xargs file

## Docker ##
docker_build_image:
	docker build --platform=linux/amd64 --build-arg TARGETARCH=amd64 -t matthiasluedtke/iconserver:latest -t matthiasluedtke/iconserver:`cat VERSION` .

docker_run:
	docker run -p 3000:8080 --env-file docker_run.env matthiasluedtke/iconserver:latest

docker_push_images_all: docker_push_image_latest docker_push_image_version

docker_push_image_latest:
	docker push matthiasluedtke/iconserver:latest

docker_push_image_version:
	docker push matthiasluedtke/iconserver:`cat VERSION`

docker_release: docker_build_image docker_push_images_all

## New GitHub Release ##
github_new_release: new_release_tag github_package
	gh release create $(shell cat VERSION)
	gh release upload $(shell cat VERSION) iconserver_*.zip

github_package: clean build_all_platforms
	zip -o -j iconserver_darwin-amd64 bin/darwin_amd64/* Readme.markdown LICENSE NOTICES
	zip -o -j iconserver_linux_amd64 bin/linux_amd64/* Readme.markdown LICENSE NOTICES
	zip -o -j iconserver_linux_arm64 bin/linux_arm64/* Readme.markdown LICENSE NOTICES
	zip -o -j iconserver_windows_amd64 bin/windows_amd64/* Readme.markdown LICENSE NOTICES
	file iconserver*.zip
	ls -alht iconserver*.zip

new_release_tag: update_notices_file bump_version rewrite-version.go git_tag_version

bump_version:
	vi VERSION

rewrite-version.go:
	echo "package besticon\n\n// Version string, same as VERSION, generated my Make\nconst VersionString = \"`cat VERSION`\"" > besticon/version.go

git_tag_version:
	git commit VERSION besticon/version.go -m "Release `cat VERSION`"
	git tag `cat VERSION`
	git push --tags
	git push

update_notices_file:
	licensed cache
	licensed notice
	cp .licenses/NOTICE NOTICES
	cat notices-more/* >> NOTICES
	git commit NOTICES -m "Update NOTICES" || echo "No change to NOTICES to commit"
