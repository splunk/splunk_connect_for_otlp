
.PHONY := build
build: ta/otlpinput/linux_x86_64/bin/otlpinput ta/otlpinput/windows_x86_64/bin/otlpinput

.PHONY := tgz
tgz: build
	tar --format ustar -C ta -czvf otlpinput.tgz otlpinput

otlpinput.tgz: tgz

ta/otlpinput/linux_x86_64/bin/otlpinput: $(shell find  **/*.go -type f)
	mkdir -p ../../ta/otlpinput/linux_x86_64/bin
	GOOS=linux GOARCH=amd64 go build -C cmd/otlpinput -trimpath -o ../../ta/otlpinput/linux_x86_64/bin/otlpinput .

ta/otlpinput/windows_x86_64/bin/otlpinput: $(shell find  **/*.go -type f)
	mkdir -p ../../ta/otlpinput/windows_x86_64/bin
	GOOS=windows GOARCH=amd64 go build -C cmd/otlpinput -trimpath -o ../../ta/otlpinput/windows_x86_64/bin/otlpinput .

.PHONY := splunk
splunk: otlpinput.tgz
	docker run --rm -it -v $(PWD)/otlpinput.tgz:/tmp/otlpinput.tgz \
		-e "SPLUNK_PASSWORD=changeme" \
		-e "SPLUNK_APPS_URL=file:///tmp/otlpinput.tgz" \
		-e "SPLUNK_START_ARGS=--accept-license" \
		-p 4317:4317 \
		-p 4318:4318 \
		-p 8000:8000 \
		splunk/splunk:9.3

.PHONY := test
test:
	go test -v ./...