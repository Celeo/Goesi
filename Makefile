all: run

BUILD_COMMAND = go build
RUN_COMMAND = ./goesi
CLEAN_COMMAND = go clean

build:
	@$(BUILD_COMMAND)

run: build
	@$(RUN_COMMAND)

clean:
	@$(CLEAN_COMMAND)

prod-run: build
	@GIN_MODE=release $(RUN_COMMAND)
