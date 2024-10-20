dev::
	wgo -file=.go -file=.templ -xfile=_templ.go templ generate :: go run server.go

client::
	wgo -file=.ts bun build.ts
