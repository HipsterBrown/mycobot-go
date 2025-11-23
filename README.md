# mycobot-go

Go library for controlling Elephant Robotics myCobot series robotic arms.

## Supported Models

- MyCobot 280
- MyCobot 320
- MechArm 270
- MyPalletizer 260

## Features

- Thread-safe concurrent access
- Context-based timeout/cancellation
- Strongly-typed API
- Interface-based gripper support
- Exposed protocol layer for advanced usage

## Installation

```bash
go get github.com/yourusername/mycobot-go
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/yourusername/mycobot-go"
    "github.com/yourusername/mycobot-go/types"
)

func main() {
    robot := mycobot.NewMyCobot280("/dev/ttyUSB0")
    ctx := context.Background()

    if err := robot.Open(ctx); err != nil {
        log.Fatal(err)
    }
    defer robot.Close()

    robot.PowerOn(ctx)
    robot.SendAngles(ctx, types.Angles{0, 0, 0, 0, 0, 0}, types.SpeedMedium)
}
```

## Documentation

See [docs/plans/2025-11-23-mycobot-go-port-design.md](docs/plans/2025-11-23-mycobot-go-port-design.md) for architecture details.

## License

MIT
