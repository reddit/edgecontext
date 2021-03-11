THRIFT_CMD=thrift

%:
	$(MAKE) -C lib/py/ $@ THRIFT=$(THRIFT_CMD)
	$(MAKE) -C lib/go/ $@ THRIFT_CMD=$(THRIFT_CMD)
