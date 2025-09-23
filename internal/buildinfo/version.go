package buildinfo

// Version is set at build time via -ldflags "-X github.com/MohammedMogeab/largo-installer/internal/buildinfo.Version=..."
// Defaults to "dev" when not provided.
var Version = "dev"

