package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type RotateLogger struct {
	Dir    string
	Prefix string
	Format string
	Handle *os.File
	FN     string
	Enable bool
	Target *log.Logger
}

func (rl *RotateLogger) rotate() {
	curFN := fmt.Sprintf("./%s/%s%s.log", rl.Dir, rl.Prefix, time.Now().Format(rl.Format))
	if curFN != rl.FN { /* need rotate */
		if f, err := os.OpenFile(curFN, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644); err != nil {
			log.Println(err)
		} else {
			if rl.Target == nil {
				log.SetOutput(f)
			} else {
				(*rl.Target).SetOutput(f)
			}
			if rl.Handle != nil {
				rl.Handle.Close()
			}
			rl.FN = curFN
			rl.Handle = f
		}
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func MakeRotateLogger(dir, prefix, format string, target *log.Logger) (rl RotateLogger) {
	rl.Dir = dir
	rl.Prefix = prefix
	rl.Format = format
	rl.Target = target
	rl.Enable = true

	if rl.Format == "" {
		rl.Format = "2010-10-16"
	}

	os.Mkdir(dir, 0755)
	rl.rotate()

	ticker := time.NewTicker(time.Second)
	go func() {
		for _ = range ticker.C {
			rl.rotate()
			if !rl.Enable {
				break
			}
		}
	}()
	return rl
}
