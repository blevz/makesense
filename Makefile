all: assets/basic.svg assets/c.svg assets/basic.png assets/c.png assets/this.png assets/this.svg build test

assets: build
	mkdir -p assets

assets/basic.svg: assets makesense
	make -C testdata/basic -Bnd | ./makesense --type gv > assets/basic.svg

assets/c.svg: assets makesense
	make -C testdata/c -Bnd | ./makesense --type gv > assets/c.svg

assets/this.svg: assets makesense
	make -Bnd | ./makesense --type gv > assets/this.svg

assets/basic.png: assets makesense
	make -C testdata/basic -Bnd | ./makesense --type dot | dot -Tpng -o assets/basic.png

assets/c.png: assets makesense
	make -C testdata/c -Bnd | ./makesense --type dot | dot -Tpng -o assets/c.png

assets/this.png: assets makesense
	make -Bnd | ./makesense --type dot | dot -Tpng -o assets/this.png

build:

makesense: build
	go build ./...

test:
	go test ./...

clean:
	rm -rf ./assets
	rm makesense