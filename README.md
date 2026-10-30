# SOMAS coursework

This is the coursework Go module `github.com/GAMINpanda/somas`.
It uses [basePlatformSOMAS](https://github.com/MattSScott/basePlatformSOMAS)
as a dependency instead of keeping a copy of the library's source here.

## Getting started

Install Go 1.23 or later, then run from this directory:

```sh
go mod download
go run .
go test ./...
```

The starter in `main.go` creates a platform server and registers one agent.
It does not start a simulation yet. Add coursework behaviour to your own agents,
messages and server, then configure the platform's game runner to run turns.

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
