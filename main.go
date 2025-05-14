package main

import (
	"context"
	"os"

	"github.com/alnah/fla/cli"
	"github.com/alnah/fla/config"
	"github.com/alnah/fla/logger"
)

func main() {
	/********* Setup logger *********/
	log := logger.New()

	/********* Setup config *********/
	path, err := config.Path()
	if err != nil {
		log.Error("retrieving config path", "error", err.Error())
		return
	}
	cfg, err := config.Load(log, path.DirPath, path.FileName)
	if err != nil {
		log.Error("loading config", "error", err.Error())
		return
	}
	if err := cfg.EnsureIODirs(); err != nil {
		log.Error("ensuring application directories", "error", err.Error())
		return
	}

	cmd := cli.NewFlaCmd()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Error("root command", "error", err.Error())
	}
}
