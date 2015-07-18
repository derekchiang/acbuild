.PHONY: all remove-temp-files

all: remove-temp-files
	godep go install ./acb

remove-temp-files:
	find . -name flymake_* -delete
