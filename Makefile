
.PHONY := build
build: ta/otlpinput/linux_x86_64/bin/otlpinput ta/otlpinput/windows_x86_64/bin/otlpinput

tgz: build
	tar --format ustar -C ta -czvf otlpinput.tgz otlpinput

ta/otlpinput/linux_x86_64/bin/otlpinput: $(shell find  **/*.go -type f)
	GOOS=linux GOARCH=amd64 go build -C cmd/otlpinput -trimpath -o ../../ta/otlpinput/linux_x86_64/bin/otlpinput .

ta/otlpinput/windows_x86_64/bin/otlpinput: $(shell find  **/*.go -type f)
	GOOS=windows GOARCH=amd64 go build -C cmd/otlpinput -trimpath -o ../../ta/otlpinput/windows_x86_64/bin/otlpinput .

splunk:
	docker run --rm -it -v $(PWD)/ta.tgz:/tmp/ta.tgz \
		-e "SPLUNK_PASSWORD=changeme" \
		-e "SPLUNK_APPS_URL=file:///tmp/ta.tgz" \
		-e "SPLUNK_START_ARGS=--accept-license" \
		-p 4317:4317 \
		-p 4318:4318 \
		-p 8000:8000 \
		splunk/splunk:9.3