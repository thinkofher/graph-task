package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
	tasks "github.com/thinkofher/graph-task"
	"github.com/thinkofher/graph-task/storage"
)

func run() error {
	ctx := context.Background()

	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		return err
	}
	defer conn.Close()

	s := storage.New(conn)
	service := &tasks.Service{
		Lister:  s,
		Storage: s,
	}

	if len(os.Args) > 2 && os.Args[1] == "id" {
		t, err := service.TaskWithID(ctx, os.Args[2])
		if err != nil {
			return err
		}
		fmt.Printf("%q\n", t)
	} else if len(os.Args) > 1 && os.Args[1] == "all" {
		allTasks, err := service.AllTasks(ctx)
		if err != nil {
			return err
		}
		for i, t := range allTasks {
			fmt.Printf("%d : %q\n", i, t)
		}
	} else {
		id, err := service.NewTask(ctx, tasks.TaskBuilder{
			Author:   "Beniamin",
			Comment:  "Very hard task",
			Deadline: time.Now().Add(24 * time.Hour),
		})
		if err != nil {
			return err
		}

		if err := service.DoneTask(ctx, id, tasks.Report{
			By: "Mariusz",
			At: time.Now().Add(time.Second * 2),
		}); err != nil {
			return err
		}

	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("run: %s\n", err)
	}
}
