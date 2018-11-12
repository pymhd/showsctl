package main

import (
	"flag"
	"fmt"
	"strconv"

	log "github.com/pymhd/go-logging"
	"github.com/pymhd/myshows"
)

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func genNotificationCaption(o myshows.EpisodeDesc) string {
	return fmt.Sprintf("New episode of %s was released\nSeason: %d, Episode: %d", o.Show.Title, o.Episode.SeasonNum, o.Episode.EpisodeNum)
}

func parseIds() []int {
	ids := flag.Args()
	ret := make([]int, len(ids))

	for n, id := range ids {
		i, err := strconv.Atoi(id)
		must(err)
		ret[n] = i
	}
	return ret
}
