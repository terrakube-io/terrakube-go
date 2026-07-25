import { defineConfig } from "sourcey";

// API reference for the Terrakube Go client library.
//
// Regenerate the godoc snapshot (requires the Go toolchain):
//   npx sourcey godoc -m . -o godoc.json
// Build the static site (no Go required — reads the committed snapshot):
//   npx sourcey build
//
// mode:"snapshot" pins to the committed godoc.json so CI and contributors
// without the Go toolchain can still build the docs (a plain `go list` run
// in the docs dir would fail when it is not the module dir).
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
