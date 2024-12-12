# jamfu

To install dependencies:

```bash
bun install

go install github.com/a-h/templ/cmd/templ@latest
go install github.com/bokwoon95/wgo@latest

# install ffmpeg
```

```
wgo -file=.go -file=.templ -xfile=_templ.go templ generate :: go run main.go

wgo -file=.ts bun build.ts
```

To run:

```bash
bun run index.ts
```

This project was created using `bun init` in bun v1.1.31. [Bun](https://bun.sh) is a fast all-in-one JavaScript runtime.
