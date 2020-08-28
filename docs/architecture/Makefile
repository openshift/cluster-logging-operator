# Makefile to generate diagrams and HTML previews.
# Diagrams are checked in, HTML previews are not. GitHub Pages renders asciidoc directly.

# Diagram source files
DOT=$(wildcard images/*.dot)
# Generate as SVG
SVG=$(patsubst %.dot, %.svg, $(DOT))
# Optional link checker
LINK_CHECK=$(shell which asciidoc-link-check)

all: $(SVG)
	$(LINK_CHECK) -q index.html

link_check: $(LINK_CHECK)
	echo "can't find $(LINK_CHECK), try `make install`"

%.html: %.adoc
	asciidoctor $<

%.svg: %.dot
	dot -Tsvg $< -o $@

# Create a local HTML file for preview, not checked in.
browse: index.html $(SVG)
	firefox $<

tools:
	npm install -g asciidoctor asciidoc-link-check

clean:
	rm -rf $(SVG) index.html


