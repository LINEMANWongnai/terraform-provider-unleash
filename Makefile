# This file is managed by makew
# DO NOT EDIT! IT WILL BE REPLACED

-include makefile.mk
-include makefiles/*

MAKEW ?= $(shell which makew)
makefile.mk: $(MAKEW) $(wildcard makew.yaml)
	@$(MAKEW)

