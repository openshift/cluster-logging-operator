# Source files
ADOC=$(shell find * -name '*.adoc')
DOT=$(shell find * -name '*.dot')

# Target files
OUT=html
HTML=$(patsubst %.adoc, $(OUT)/%.html, $(ADOC))
SVG=$(patsubst %.dot, $(OUT)/%.svg, $(DOT))

all: $(HTML) $(SVG)

$(OUT):
	mkdir $(OUT)

browse: all
	xdg-open $(OUT)/index.html; sleep 1 # xdg-open bug - needs sleep

$(OUT)/%.html: %.adoc $(OUT)
	asciidoctor -v -r asciidoctor-diagram $< -D $(OUT)

$(OUT)/%.svg: %.dot $(OUT)
	dot -Tsvg $< -o $@

clean:
	rm -rf $(OUT)
