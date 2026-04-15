# mycobot-go

Go library for controlling Elephant Robotics myCobot series robotic arms.

## Supported Models

- MechArm 270

## Features

- Thread-safe concurrent access
- Context-based timeout/cancellation
- Strongly-typed API
- Interface-based gripper support
- Exposed protocol layer for advanced usage

## Installation

```bash
go get github.com/hipsterbrown/mycobot-go
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/hipsterbrown/mycobot-go"
    "github.com/hipsterbrown/mycobot-go/types"
)

func main() {
    arm := mycobot.NewMechArm270("/dev/ttyUSB0")
    ctx := context.Background()

    if err := arm.Open(ctx); err != nil {
        log.Fatal(err)
    }
    defer arm.Close()

    arm.PowerOn(ctx)
    arm.SendAngles(ctx, types.Angles{0, 0, 0, 0, 0, 0}, types.SpeedMedium)
}
```

## Documentation

See [docs/plans/2025-11-23-mycobot-go-port-design.md](docs/plans/2025-11-23-mycobot-go-port-design.md) for architecture details.

## License

MIT
