package main

import (
	"flag"
	"fmt"

	"github.com/evilsocket/shellz/core"
	"github.com/evilsocket/shellz/log"
	"github.com/evilsocket/shellz/models"
	"github.com/evilsocket/shellz/queue"
)

var (
	runCommand = "uptime"
	onFilter   = "*"
	onNames    = []string{}
	on         = models.Shells{}
)

func init() {
	flag.StringVar(&runCommand, "run", runCommand, "Command to run on the selected shells.")
	flag.StringVar(&onFilter, "on", onFilter, "Comma separated list of shell names to select or * for all.")
	flag.BoolVar(&log.DebugMessages, "debug", log.DebugMessages, "Enable debug messages.")
	flag.Parse()
}

func main() {
	log.Raw(core.Banner)

	err, idents, shells := models.Load()
	if err != nil {
		log.Fatal("error while loading identities and shells: %s", err)
	} else if len(shells) == 0 {
		log.Fatal("no shells found on the system, start creating json files inside %s", models.Paths["shells"])
	} else {
		log.Debug("loaded %d identities and %d shells", len(idents), len(shells))
	}

	if onFilter == "*" {
		on = shells
	} else {
		for _, name := range core.CommaSplit(onFilter) {
			if shell, found := shells[name]; !found {
				log.Fatal("can't find shell %s", name)
			} else {
				on[name] = shell
			}
		}
	}

	if len(on) == 0 {
		log.Fatal("no shell selected by the filter %s", core.Dim(onFilter))
	}

	wq := queue.New(-1, func(job queue.Job) {
		shell := job.(models.Shell)
		name := shell.Name
		err, session := shell.NewSession()
		if err != nil {
			log.Warning("error while creating session for shell %s: %s", name, err)
			return
		}
		defer session.Close()

		out, err := session.Exec(runCommand)
		if err != nil {
			log.Error("%s (%s %s) > %s (%s)\n\n%s", core.Bold(name), core.Green(shell.Type), core.Dim(fmt.Sprintf("%s:%d", shell.Address, shell.Port)), runCommand, core.Red(err.Error()), out)
		} else {
			log.Info("%s (%s %s) > %s\n\n%s", core.Bold(name), core.Green(shell.Type), core.Dim(fmt.Sprintf("%s:%d", shell.Address, shell.Port)), core.Blue(runCommand), out)
		}
	})

	log.Info("running %s on %d shells ...\n", core.Dim(runCommand), len(on))

	for name, _ := range on {
		wq.Add(on[name])
	}

	wq.WaitDone()
}
