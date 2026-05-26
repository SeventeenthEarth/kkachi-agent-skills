.PHONY: test test-prepare test-unit test-int test-e2e

PYTHON ?= python3
PYTHONPATH := src

test-prepare:
	PYTHONPATH=$(PYTHONPATH) $(PYTHON) -m compileall -q src tests

test-unit:
	PYTHONPATH=$(PYTHONPATH) $(PYTHON) -m unittest tests.test_discovery

test-int:
	PYTHONPATH=$(PYTHONPATH) $(PYTHON) -m unittest tests.test_cli_integration

test-e2e:
	PYTHONPATH=$(PYTHONPATH) KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1 $(PYTHON) -m unittest tests.test_e2e_real_repo

test: test-prepare test-unit test-int test-e2e
