A spiritual successor to [makefile2graph](https://github.com/lindenb/makefile2graph/) rewritten in go with added functionality.

Generate visualizations of makefile DAGs:

Graphviz to svg:

`make -C testdata/basic -Bnd | ./makesense --type gv > assets/basic.svg`

![assets/basic.svg](assets/basic.svg)

`make -C testdata/c -Bnd | ./makesense --type gv > assets/c.svg`

![assets/c.svg](assets/c.svg)

This project's makefile:

![assets/this.svg](assets/this.svg)

Dot:

`make -C testdata/basic -Bnd | ./makesense --type dot | dot -Tpng -o assets/basic.png`

![assets/basic.png](assets/basic.png)

`make -C testdata/c -Bnd | ./makesense --type dot | dot -Tpng -o assets/c.png`

![assets/c.png](assets/c.png)

This project's makefile:

![assets/this.png](assets/this.png)