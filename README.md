# Minefetch

Minefetch is a neofetch-like tool for fetching Minecraft server information.

![Output of `minefetch hypixel.net`](hypixel.png)

## Install

Minefetch is a single binary with no third-party dependencies,
so installation is as simple as downloading the binary for your platform
and placing it in a convenient location.

To update, simply repeat the installation process.

### Script

For convenience, you can use the provided scripts to download and install the latest version.

#### Unix-like (macOS and Linux)

```sh
curl -fsSL bhv.sh/minefetch.sh | sh
```

#### Windows

```ps1
iwr bhv.sh/minefetch.ps1 | iex
```

### Go

If you have Go installed, you can use the `go install` command.

```sh
go install bhv.sh/minefetch@latest
```

### Manual

To install an older version, or if the above methods do not work,
you can manually download a prebuilt binary from the
[GitHub releases](https://github.com/bhavjitChauhan/minefetch/releases) page.

## Usage

Run against a Java Edition server:

```sh
minefetch hypixel.net
```

Run against a Bedrock Edition server:

```sh
minefetch --bedrock play.lbsg.net
```

View all available options:

```sh
minefetch --help
```

## Features

- [x] Java Edition
- [x] Bedrock Edition (`--bedrock`)
- [x] Server icon
- [x] RGB text
- [x] Crossplay
- [x] Cracked servers (`--cracked`)
- [x] Mojang's blocked server list (`--blocked`)
- [x] Query (`--query`)
- [x] RCON (`--rcon`)
- [x] Chat report prevention
- [x] SRV lookup
- [x] Raw output (`--output raw`)
- [ ] MOTD sprites
- [ ] Legacy status
- [ ] Newer Forge servers

Contributions are welcome.

## Structure

Minefetch has no third-party dependencies.
All libraries are implemented in the (internal)[internal] directory.

```
.                      Main package
└── internal
    ├── mc             Subset of the Java Edition protocol
    ├── mcpe           Subset of the Raknet protocol as used by Bedrock Edition
    ├── term           Terminal syscalls and ANSI/xterm escape codes
    ├── emoji          Emoji detection and manipulation
    ├── flag           CLI flag parsing
    └── image
        ├── sixel      Sixel encoding
        ├── scale      Lanczos image scaling
        ├── quant      Median-cut image quantization
        ├── pngconfig  PNG header decoding
        └── print      Terminal image rendering via Unicode
```

These packages are not intended for external use, and may break at any time.

## Related

- [neofetch](https://github.com/dylanaraps/neofetch)
- [mcstatus.io](https://mcstatus.io)
- [mcsrvstat.us](https://mcsrvstat.us)
