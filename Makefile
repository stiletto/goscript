all: goscript.$(GOARCH)

clean:
	rm -f *.8 *.386

goscript.386: goscript.8
	8l -o goscript.386 goscript.8

goscript.8: goscript.go
	8g goscript.go

