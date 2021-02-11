LIBRARY_ROOT := reddit_edgecontext/
PYTHON_ROOTS := $(LIBRARY_ROOT) tests/
PYTHON_FILES = $(shell find $(PYTHON_ROOTS) setup.py -name '*.py')
REORDER_PYTHON_IMPORTS := reorder-python-imports --py3-plus --separate-from-import --separate-relative
THRIFT := thrift


thrift: $(LIBRARY_ROOT)/thrift/__init__.py


.PHONY: fmt
fmt:
	$(REORDER_PYTHON_IMPORTS) --exit-zero-even-if-changed $(PYTHON_FILES)
	black $(PYTHON_FILES)


.PHONY: lint
lint:
	$(REORDER_PYTHON_IMPORTS) --diff-only $(PYTHON_FILES)
	black --diff --check $(PYTHON_FILES)
	flake8 $(PYTHON_ROOTS)
	mypy $(LIBRARY_ROOT)


.PHONY: test
test:
	python -m pytest -v tests/


.PHONY: docs
docs:
	sphinx-build -M html docs/ build/


$(LIBRARY_ROOT)/thrift/__init__.py: ../../edgecontext.thrift
	mkdir -p build/thrift/$<
	$(THRIFT) -strict -gen py:slots -out build/thrift/ $<
	cp -r build/thrift/reddit_edgecontext/thrift/ reddit_edgecontext/
