package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/belak/go-gitdir"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config stores all the server-level settings. These cannot be changed at
// runtime. They are only used by the binary and are passed to the proper
// places.
type Config struct {
	BindAddr  string
	BasePath  string
	LogFormat string
	LogDebug  bool
}

// DefaultConfig is used as the base config.
var DefaultConfig = Config{
	BindAddr:  ":2222",
	BasePath:  "./tmp",
	LogFormat: "json",
}

// NewEnvConfig returns a new Config based on environment variables.
func NewEnvConfig() (gitdir.ServerConfig, error) {
	var err error

	c := DefaultConfig

	ret := gitdir.ServerConfig{}

	if rawDebug, ok := os.LookupEnv("GITDIR_DEBUG"); ok {
		c.LogDebug, err = strconv.ParseBool(rawDebug)
		if err != nil {
			return ret, errors.Wrap(err, "GITDIR_DEBUG")
		}
	}

	if logFormat, ok := os.LookupEnv("GITDIR_LOG_FORMAT"); ok {
		if logFormat != "console" && logFormat != "json" {
			return ret, errors.New("GITDIR_LOG_FORMAT: must be console or json")
		}

		c.LogFormat = logFormat
	}

	// Set up the logger - anything other than console defaults to json.
	if c.LogFormat == "console" {
		log.Logger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	}

	if c.LogDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if bindAddr, ok := os.LookupEnv("GITDIR_BIND_ADDR"); ok {
		c.BindAddr = bindAddr
	}

	var ok bool

	if c.BasePath, ok = os.LookupEnv("GITDIR_BASE_DIR"); !ok {
		return ret, errors.New("GITDIR_BASE_DIR: not set")
	}

	if c.BasePath, err = filepath.Abs(c.BasePath); err != nil {
		return ret, errors.Wrap(err, "GITDIR_BASE_DIR")
	}

	info, err := os.Stat(c.BasePath)
	if err != nil {
		return ret, errors.Wrap(err, "GITDIR_BASE_DIR")
	}

	if !info.IsDir() {
		return ret, errors.New("GITDIR_BASE_DIR: not a directory")
	}

	ret.Addr = c.BindAddr
	ret.BaseDir = c.BasePath

	return ret, nil
}