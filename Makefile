all: goscript.$(GOARCH)

clean:
	rm -f *.8 *.386

goscript.386: goscript.8
	8l -o goscript.386 goscript.8

goscript.8: goscript.go
	8g goscript.go

fixindent:
	gofmt -spaces=true -tabwidth=4 -tabindent=false < goscript.go > goscript.go.temp
	cat goscript.go.temp > goscript.go
	rm goscript.go.temp
