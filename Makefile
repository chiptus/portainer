lint-errors:
	cd api && git status --porcelain | GREP \.go | cut -d' ' -f 3 | sed 's#api/#./#' | xargs wrapcheck && cd ..

lint-pr: lint-errors

# Assert wrapcheck is installed.
ifeq (, $(shell which wrapcheck))
	$(error Building this project requires wrapcheck linter. Do 'go install github.com/tomarrell/wrapcheck/v2/cmd/wrapcheck@v2' to install it)
endif
