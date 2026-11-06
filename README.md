# SOMAS coursework

This is the coursework Go module `github.com/GAMINpanda/somas`.
It uses [basePlatformSOMAS](https://github.com/MattSScott/basePlatformSOMAS)
as a dependency instead of keeping a copy of the library's source here.

## Getting started

Install Go 1.23 or later, then run from this directory:

```sh
go mod download
go run .          # run one game and record it to runs/latest.json
go test ./...
```

## Watching a game

The quickest way to see a game is:

```sh
go run . -serve
```

This runs a game, then serves the viewer with that game loaded at
<http://localhost:8080>. Open that address in a browser. Press Ctrl+C to stop.
Each run is random, so stop and run the command again to watch a new game.

In the viewer:

- The grid is the vegetable patch. Tiles go from green (full) to bare soil (eaten).
- Each AmoeBoid stays on its own cell. The green bar under it is its body energy
  (from vegetables) and the yellow bar is its brain energy (from sunlight).
- The agent eating this turn has a gold ring. Dead agents are grey.
- Clouds drift across the patch once per round. Agents under a cloud get less
  sunlight. Hover over a cell to see its vegetables, cloud cover and agent.
- The side panel shows the patch total (CPR), the sunlight kitty (CTK) against
  its threshold (green when the patch regrew), how many agents are alive,
  and a log of everything that has happened so far.

Controls: Play/Pause (or the space bar), step back and forward (or the arrow
keys), the slider to jump anywhere, and a speed menu. Adding `#frame=12` to the
address opens the viewer at frame 12, which is handy for showing a particular
moment.

### Options

| Flag | Default | Meaning |
|---|---|---|
| `-agents` | 10 | number of agents |
| `-rounds` | 5 | number of rounds (M) |
| `-turns` | 10 | turns per round (N) |
| `-out` | `runs/latest.json` | where the recorded game is saved |
| `-serve` | off | serve the viewer after the game |
| `-addr` | `localhost:8080` | address for the viewer (change it if 8080 is in use) |

For example: `go run . -serve -agents 20 -rounds 10 -turns 8`.

### Viewing a saved game

The viewer is the single file `ui/index.html`. You can open it in a browser
without the server, click **Open run…** and choose a saved `.json` file
(or drag the file onto the page). Saved games are in `runs/`, which Git ignores.

The viewer is built into the program, so after editing `ui/index.html`
run `go run . -serve` again to see the change.

## Code layout

| Path | What it holds |
|---|---|
| `main.go`, `serve.go` | command-line flags, recording, and the viewer web server |
| `internal/server` | the game: rounds, turns, sunlight, appropriation, provision, regrowth, deaths |
| `internal/common/objects` | the AmoeBoid agent (random strategy for now) and its messages |
| `internal/patch` | the vegetable patch as a grid of tiles, kept equal to the CPR total |
| `internal/weather` | clouds that drift over the grid and shade agents from the sun |
| `internal/recorder` | records each step of a game as JSON for the viewer |
| `ui/index.html` | the viewer |

## Using the platform

Import the exported packages in your coursework code:

```go
import (
    "github.com/MattSScott/basePlatformSOMAS/v2/pkg/agent"
    "github.com/MattSScott/basePlatformSOMAS/v2/pkg/message"
    "github.com/MattSScott/basePlatformSOMAS/v2/pkg/server"
)
```

Go downloads the library into its module cache outside this repository.
`go.mod` pins the dependency version, and `go.sum` records its checksums.
Commit both files so teammates use the same dependencies. You do not need to
run `go mod init` again.

To update the platform deliberately:

```sh
go get github.com/MattSScott/basePlatformSOMAS/v2@latest
go mod tidy
go test ./...
```

See the [platform manual](https://github.com/MattSScott/basePlatformSOMAS/blob/main/basePlatformSOMASv2.0.pdf)
for its extension points and simulation lifecycle.

The existing `LICENSE` retains the original platform's MIT license notice
from the initial import. The platform's original history remains in Git;
its source files are now supplied through the dependency.
