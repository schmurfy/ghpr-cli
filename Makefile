

VERSION := $(shell cat version)
VERSION_TAG := v$(VERSION)


build:
	
	go build -ldflags="-X 'main.Version=v$(VERSION)'" -o ghpr .

release: build
	git tag v$(VERSION)
	git push --tags
	github-release release -u schmurfy -r ghpr-cli -t $(VERSION_TAG)
	github-release upload -u schmurfy -r ghpr-cli -t $(VERSION_TAG) -f ghpr --name ghpr
	
