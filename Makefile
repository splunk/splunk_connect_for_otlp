GOCMD?=go

FIND_MOD_ARGS=-type f -name "go.mod"  -not -path "./packaging/technical-addon/*"
TO_MOD_DIR=dirname {} \; | sort | egrep  '^./'

ALL_MODS := $(shell find . $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR)) $(PWD)

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

.PHONY := build
build: ta/otlpinput/linux_x86_64/bin/otlpinput ta/otlpinput/windows_x86_64/bin/otlpinput

.PHONY := splunk
splunk: otlpinput.tgz
	docker run --rm -it -v $(PWD)/otlpinput.tgz:/tmp/otlpinput.tgz \
		-e "SPLUNK_PASSWORD=changeme" \
		-e "SPLUNK_APPS_URL=file:///tmp/otlpinput.tgz" \
		-e "SPLUNK_START_ARGS=--accept-license" \
		-e "SPLUNK_HEC_TOKEN=000000-0000-00000-0000000000" \
		-p 4317:4317 \
		-p 4318:4318 \
		-p 8000:8000 \
		splunk/splunk:9.3

# Define a delegation target for each module
.PHONY: $(ALL_MODS)
$(ALL_MODS):
	@echo "Running target '$(TARGET)' in module '$@'"
	$(MAKE) --no-print-directory -C $@ $(TARGET)

# Triggers each module's delegation target
.PHONY: for-all-target
for-all-target: $(ALL_MODS)

.PHONY: test-all
test-all:
	$(MAKE) for-all-target TARGET="test"
	$(MAKE) test

.PHONY: gotidy
gotidy:
	@for mod in $$(find . -name go.mod | xargs dirname); do \
		echo "Tidying $$mod"; \
		(cd $$mod && rm -rf go.sum && $(GOCMD) mod tidy -compat=1.24.0 && $(GOCMD) get toolchain@none) || exit $?; \
	done
