
.PHONY := build
build: ta/otlpstdout/bin/otlpstdout

tgz:
	tar --format ustar -C ta -czvf ta.tgz otlpstdout

ta/otlpstdout/bin/otlpstdout: main.go
	GOOS=linux GOARCH=amd64 go build -trimpath -o ./ta/otlpstdout/bin/otlpstdout .

splunk:
	docker run --rm -it -v $(PWD)/ta.tgz:/tmp/ta.tgz \
		-e "SPLUNK_PASSWORD=changeme" \
		-e "SPLUNK_APPS_URL=file:///tmp/ta.tgz" \
		-e "SPLUNK_START_ARGS=--accept-license" \
		-p 4317:4317 \
		-p 4318:4318 \
		-p 8000:8000 \
		splunk/splunk:9.3