await Bun.build({
  entrypoints: ["./client/player.ts", "./client/two.ts"],
  outdir: "./client",
});
