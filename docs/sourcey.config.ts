import { defineConfig } from "sourcey";

// API reference for the Terrakube Go client library.
//
// Build locally (from this directory):
//   npm install --no-save sourcey
//   ./node_modules/.bin/sourcey godoc -m .. -o godoc.json   # needs the Go toolchain
//   ./node_modules/.bin/sourcey build
//
// godoc.json is a generated artifact and is NOT committed — the Docs workflow
// regenerates it from the checkout on every build, so the reference can never
// drift from the source it documents.
//
// mode:"snapshot" reads that regenerated file rather than shelling out to
// `go list` from this directory (which is not the module dir, so it would fail).
export default defineConfig({
  prettyUrls: "strip",
  baseUrl: "/terrakube-go",
  name: "Terrakube Go Client",
  // Source links from generated API symbols back to the Go source on main.
  repo: "https://github.com/terrakube-io/terrakube-go",
  editBranch: "main",
  editBasePath: "",
  theme: {
    preset: "default",
    colors: { primary: "#7b3fe4" },
  },
  navigation: {
    tabs: [
      {
        tab: "Guide",
        slug: "",
        groups: [{ group: "Start", pages: ["introduction"] }],
      },
      {
        tab: "API Reference",
        slug: "api",
        godoc: {
          module: "..",
          packages: ["./..."],
          snapshot: "godoc.json",
          mode: "snapshot",
          includeTests: true,
          sourceBasePath: "",
        },
      },
    ],
  },
});
