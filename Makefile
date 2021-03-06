.PHONY: all build clean test

all:: assets/basic.svg assets/c.svg assets/basic.png assets/c.png assets/this.png build test

include makesense.mk

assets:
	mkdir -p assets

assets/basic.svg: assets makesense
	@MAKE -C testdata/basic -Bnd | ./makesense --type gv > assets/basic.svg

assets/c.svg: assets makesense
	make -C testdata/c -Bnd | ./makesense --type gv > assets/c.svg

assets/basic.png: assets makesense
	make -C testdata/basic -Bnd | ./makesense --type dot | dot -Tpng -o assets/basic.png

assets/c.png: assets makesense
	make -C testdata/c -Bnd | ./makesense --type dot | dot -Tpng -o assets/c.png

assets/this.png: assets makesense
	make -Bnd | ./makesense --type dot | dot -Tpng -o assets/this.png

build: makesense

makesense: makesense.go makesense_test.go
	go build ./...

test: makesense_test.go
	go test ./...

clean::
	rm -rf ./assets
	rm -f makesense