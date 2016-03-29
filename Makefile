build:
	go get github.com/mat/besticon/besticon/iconserver
	go get github.com/mat/besticon/besticon

test_all: build test test_bench
	go test -v github.com/mat/besticon/besticon/iconserver

test:
	go test -v github.com/mat/besticon/ico
	go test -v github.com/mat/besticon/besticon
	go test -v github.com/mat/besticon/lettericon
	go test -v github.com/mat/besticon/colorfinder

test_race:
	go test -v -race github.com/mat/besticon/ico
	go test -v -race github.com/mat/besticon/besticon
	go test -v -race github.com/mat/besticon/besticon/iconserver
	go test -v -race github.com/mat/besticon/lettericon
	go test -v -race github.com/mat/besticon/colorfinder

test_bench:
	go test github.com/mat/besticon/lettericon -bench .
	go test github.com/mat/besticon/colorfinder -bench .

update_godeps:
	godep save ./...

install_godeps:
	grep ImportPath Godeps/Godeps.json | cut -d ":" -f 2 | tr -d '"' | tr -d "," | grep -v besticon | xargs -n 1 | xargs go get

deploy:
	git push heroku master
	heroku config:set DEPLOYED_AT=`date +%s`

install:
	go get ./...

run_server:
	go build -o bin/iconserver github.com/mat/besticon/besticon/iconserver
	PORT=3000 DEPLOYED_AT=`date +%s` ./bin/iconserver

install_devtools:
	go get golang.org/x/tools/cmd/...
	go get github.com/golang/lint/golint
	go get github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata/...

style:
	find . -name "*.go" | grep -v Godeps/ | xargs go tool vet -all
	find . -name "*.go" | grep -v Godeps/ | xargs golint

coverage_besticon:
	go test -coverprofile=coverage.out -covermode=count github.com/mat/besticon/besticon && go tool cover -html=coverage.out && unlink coverage.out

coverage_ico:
	go test -coverprofile=coverage.out -covermode=count github.com/mat/besticon/ico && go tool cover -html=coverage.out && unlink coverage.out

coverage_iconserver:
	go test -coverprofile=coverage.out -covermode=count github.com/mat/besticon/besticon/iconserver && go tool cover -html=coverage.out && unlink coverage.out

vendor_dependencies:
	godep save -r ./...
	# Need to go get in order to fill $GOPATH/pkg... to minimize compile times:
	go get ./...

test_websites:
	go get ./...
	cat besticon/testdata/websites.txt | xargs -P 10 -n 1  besticon

minify_css:
	curl -X POST -s --data-urlencode 'input@besticon/iconserver/assets/main.css' http://cssminifier.com/raw > besticon/iconserver/assets/main-min.css

update_assets:
	go-bindata -pkg assets -ignore assets.go -o besticon/iconserver/assets/assets.go besticon/iconserver/assets/

gotags:
	gotags -tag-relative=true -R=true -sort=true -f="tags" -fields=+l .

#
## Building ##
#

clean:
	rm -rf bin/*
	rm -f iconserver*.zip

build_darwin_amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/darwin_amd64/iconserver github.com/mat/besticon/besticon/iconserver

build_linux_amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/linux_amd64/iconserver github.com/mat/besticon/besticon/iconserver

build_windows_amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/windows_amd64/iconserver.exe github.com/mat/besticon/besticon/iconserver

build_all_platforms: build_darwin_amd64 build_linux_amd64 build_windows_amd64
	find bin/ -type file | xargs file

github_package: clean build_all_platforms
	zip -o -j iconserver_darwin-amd64 bin/darwin_amd64/* Readme.markdown LICENSE
	zip -o -j iconserver_linux_amd64 bin/linux_amd64/* Readme.markdown LICENSE
	zip -o -j iconserver_windows_amd64 bin/windows_amd64/* Readme.markdown LICENSE
	file iconserver*.zip
	ls -alht iconserver*.zip

build_docker_image: build_linux_amd64
	docker build -t matthiasluedtke/iconserver .

new_release: bump_version rewrite-version.go git_tag_version

bump_version:
	cat VERSION
	head -n1 VERSION | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g' > NEW_VERSION
	mv NEW_VERSION VERSION
	cat VERSION

rewrite-version.go:
	echo "package besticon\n\n// Version string, same as VERSION, generated my Make\nconst VersionString = \"`cat VERSION`\"" > besticon/version.go

git_tag_version:
	git commit VERSION besticon/version.go -m "Release `cat VERSION`"
	git tag `cat VERSION`
