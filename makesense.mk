all:: .makesense/make.svg

.PHONY: .makesense_clean clean

.makesense:
	mkdir -p .makesense

.makesense/make.svg: .makesense makesense
	make -Bnd | ./bin/makesense --type gv > .makesense/make.svg

.makesense_clean:
	rm -rf .makesense

clean:: .makesense_clean