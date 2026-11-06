package main

import (
	"embed"
	"io/fs"
	"net/http"
)

// The viewer is built into the program, so -serve works from any folder without copying the ui files around.
// Rebuild (go run .) after editing ui/index.html to see the change.
//
//go:embed ui
var uiFiles embed.FS

// newViewerHandler serves the viewer at / and the recorded run at /run.json
func newViewerHandler(runJSON []byte) http.Handler {
	ui, err := fs.Sub(uiFiles, "ui")
	if err != nil {
		panic(err) // only possible if the embed line above is wrong
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServerFS(ui))
	mux.HandleFunc("/run.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(runJSON)
	})
	return mux
}
