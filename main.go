package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	config "github.com/sssseraphim/gator/internal/config"
	"github.com/sssseraphim/gator/internal/database"
	"log"
	"os"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", conf.Db_url)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	dbQueries := database.New(db)

	programState := &state{
		db:  dbQueries,
		cfg: &conf,
	}
	cmds := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerListFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	if len(os.Args) < 2 {
		log.Fatal("no arguments")
	}
	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]
	err = cmds.run(programState, command{
		name: cmdName,
		args: cmdArgs,
	})
	if err != nil {
		log.Fatal(err)
	}
}
