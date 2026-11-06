package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/GAMINpanda/somas/internal/server"
)

func main() {
	agents := flag.Int("agents", 10, "number of agents")
	rounds := flag.Int("rounds", 5, "number of rounds (M)")
	turns := flag.Int("turns", 10, "number of turns per round (N)")
	out := flag.String("out", "runs/latest.json", "where to write the recorded run for the UI")
	serve := flag.Bool("serve", false, "after the run, serve the viewer with this run loaded")
	addr := flag.String("addr", "localhost:8080", "address the viewer is served on with -serve")
	flag.Parse()

	s := server.Initialise(*agents, *rounds, *turns, 10, time.Millisecond) // message bandwidth of 10 and a 1ms turn timeout
	rec := s.EnableRecording(*rounds, *turns)
	s.Start()
	fmt.Printf("SOMAS Platform Finished: %d cycles %d agents still alive\n", s.Iterations, len(s.GetAgentMap()))

	if err := rec.Write(*out); err != nil {
		log.Fatalf("could not write the run to %s: %v", *out, err)
	}
	fmt.Printf("Recorded %d frames to %s\n", len(rec.Run().Frames), *out)

	if *serve {
		runJSON, err := rec.JSON()
		if err != nil {
			log.Fatalf("could not encode the run: %v", err)
		}
		fmt.Printf("Viewer ready at http://%s (press Ctrl+C to stop)\n", *addr)
		log.Fatal(http.ListenAndServe(*addr, newViewerHandler(runJSON)))
	}
}
