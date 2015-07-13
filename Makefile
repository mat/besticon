
test_all: test
	go test -v github.com/mat/besticon/besticon/iconserver
	go get ./...

test:
	go test -v github.com/mat/besticon/ico
	go test -v github.com/mat/besticon/besticon

test_race:
	go test -v -race github.com/mat/besticon/ico
	go test -v -race github.com/mat/besticon/besticon
	go test -v -race github.com/mat/besticon/besticon/iconserver

update_godeps: 
	godep save ./...

install_godeps:
	grep ImportPath Godeps/Godeps.json | cut -d ":" -f 2 | tr -d '"' | tr -d "," | grep -v besticon | xargs -n 1 | xargs go get

deploy:
	git push heroku master
	heroku config:set GIT_REVISION=`git describe --always` DEPLOYED_AT=`date +%s` ICONS_ANALYTICS_ID=$(ICONS_ANALYTICS_ID)

install:
	go get ./...

run_server:
	ICONS_ANALYTICS_ID=$(ICONS_ANALYTICS_ID) DEPLOYED_AT=`date +%s` GIT_REVISION=`git describe --always` go run besticon/iconserver/server.go -port=3000

install_devtools:
	go get golang.org/x/tools/cmd/...
	go get github.com/golang/lint/golint
	go get github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata/...

check:
	find . -name "*.go" | grep -v Godeps/ | xargs go tool vet -all
	find . -name "*.go" | grep -v Godeps/ | xargs golint

coverage_besticon:
	go test -coverprofile=coverage.out -covermode=count github.com/mat/besticon/besticon && go tool cover -html=coverage.out && unlink coverage.out

coverage_ico:
	go test -coverprofile=coverage.out -covermode=count github.com/mat/besticon/ico && go tool cover -html=coverage.out && unlink coverage.out

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

build_linux_amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/linux_amd64/iconserver besticon/iconserver/server.go

build_docker_image: build_linux_amd64
	docker build -t matthiasluedtke/iconserver .

