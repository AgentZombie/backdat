package main

import (
	"flag"
	"log"

	"github.com/AgentZombie/gu"

	"backdat"
	"backdat/fs"
)

var (
	fBackupPath = flag.String("backup", "", "path to backup")
	fInit       = flag.Bool("init", false, "initialize store")
	fStorePath  = flag.String("store", "", "path to FS store")
)

func main() {
	flag.Parse()
	if *fStorePath == "" {
		log.Fatal("required: -store")
	}
	if *fInit {
		gu.FatalIfError(fs.Init(*fStorePath), "initializing FS store")
		return
	}
	if *fBackupPath == "" {
		log.Fatal("required: -backup")
	}
	chunks, fps, ss, err := fs.New(*fStorePath)
	gu.FatalIfError(err, "accessing FS store")

	b := backdat.Backup{
		Chunks:    chunks,
		FP:        fps,
		Snapshots: ss,
	}

	gu.FatalIfError(b.Backup(*fBackupPath), "performing backup")
}
