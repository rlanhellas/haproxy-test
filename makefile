.DEFAULT_GOAL := help
SHELL := /bin/bash

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  run: Run the program"
	@echo "  switch: Switch the priority between ServerA and ServerB"

.PHONY: run
run:
	@echo "Running the program..."
	docker compose up -d --build

.PHONY: switch
switch:
	@echo "Switching priority between ServerA and ServerB ..."
	@./switch.sh
