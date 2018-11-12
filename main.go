package main

import (
	"flag"
	"fmt"
	"time"
	log "github.com/pymhd/go-logging"
	"github.com/pymhd/myshows"
	"sort"
	bot "tlgrm-bot"
)

var (
	//sm StorManager initialezed in stor.go - json cache
	cfg      *Config
	top      int
	info     int
	watch    bool
	unwatch  bool
	list     bool
	skipflag bool
	vflag    bool
	search   string
	conffile string
	
)

func init() {
	//flag stuff
	flag.StringVar(&conffile, "config", "/etc/myshows/config.yaml", "Specify conffile, default: /etc/myshows/config.yaml")
	flag.StringVar(&search, "search", "", "String. Name to search")
	flag.BoolVar(&watch, "watch", false, "Add show ids to watchlist")
	flag.BoolVar(&unwatch, "unwatch", false, "Del show ids from watchlist")
	flag.BoolVar(&list, "list", false, "List all my shows on myshows.com with 'watching' status")
	flag.BoolVar(&skipflag, "skip", false, "Do not actually send (for instance if empty cache)")
	flag.BoolVar(&vflag, "v", false, "Enable logging")
	flag.IntVar(&top, "top", 0, "Number of episodes to show")
	flag.IntVar(&info, "info", 0, "Search info about show")
	flag.Parse()
	
	if vflag {
		log.EnableDebug()
	}
}

func main() {
	cfg = LoadCfg(conffile)
	defer sm.Save()

	switch {
	case watch:
		MarkEpisodesAsWatched()
	case unwatch:
		ids := parseIds()
		sm.DelFromWatchlist(ids...)
	case list:
		ListAllShows()
	case len(search) > 0:
		SearchShow(search)
	case top > 0:
		ShowTopEpisodes(top)
	case info > 0:
		GetShowInfo(info)
	default:
		//if no flags (main purpose of script)
		NotifyUsers()
	}

}

func MarkEpisodesAsWatched() {
	ids := parseIds()
	log.Debugf("Next ids will be marked as watched: %v\n", ids)
	
	//add as watched in local stor
	sm.AddToWatchlist(ids...)
	
	for _, id := range ids {
		err := myshows.SetShowAsWatching(sm.Token, id)
		if err != nil {
			renewToken()
			err = myshows.SetShowAsWatching(sm.Token, id)
			must(err)
		}
	}
}

func NotifyUsers() {
	defer func(now time.Time) {
		log.Debugf("Whole process of user notification took: %v\n", time.Since(now))
	}(time.Now())
	
	log.Debugln("Starting notification procedure")
	//get guarantee list of episodes
	u := getUnwatchedEpisodes()
	log.Debugf("Found %d episodes to proceed\n", len(u))
	
	var proceeded int
	
	for _, obj := range u {
		if sm.IsMonitored(obj.Show.Id) {
			if !sm.IsSent(obj.Episode.Id) {
				proceeded++
				
				cap := genNotificationCaption(obj)
				log.Debugf("Found unhandled episode. Cap:\n%s\n", cap)
				if skipflag {
					//just skip sending notification
					sm.MarkAsSent(obj.Episode.Id)
					log.Debugf("Marked episode %d as sent\n", obj.Episode.Id)
					log.Warningln(cap)
					continue
				}

				for _, chatId := range cfg.Telegram.Watchers {
					var err error
					switch {
					case len(obj.Episode.Image) > 0: //we have special image for episode
						log.Debugln("Episode contains it's own image url")
						_, _, err = bot.SendPhoto(cfg.Telegram.Key, chatId, obj.Episode.Image, cap, 0)
					case len(obj.Show.Image) > 0: //at least we have image for show
						log.Debugln("Using show's image url")
						_, _, err = bot.SendPhoto(cfg.Telegram.Key, chatId, obj.Show.Image, cap, 0)
					default: // well that's just plain text notification
						log.Warningln("Episode does not have any image url. Neither episode nor show")
						_, err = bot.SendTextMessage(cfg.Telegram.Key, chatId, cap, 0)
					}

					// no need to continue if problems with notifications
					// need to handle them first
					must(err)
					log.Debugf("Successfully sent notification to user with chat id: %d\n", chatId)
				}
				sm.MarkAsSent(obj.Episode.Id)
				log.Debugf("Marked episode %d as sent\n", obj.Episode.Id)
			}
		}
	}
	if proceeded == 0 {
		log.Debugln("There was no any new episodes to proceed (all sent).")
	}
}

func ListAllShows() {
	defer func(now time.Time) {
                log.Debugf("List all my shows took: %v\n", time.Since(now))
        }(time.Now())

	shows := getShowList()

	order := make([]int, len(shows))
	showsMap := make(map[int]myshows.Show, 0)

	for n, s := range shows {
		showsMap[s.Show.Id] = s.Show
		order[n] = s.Show.Id
	}

	sort.Ints(order)

	var prefix string
	for _, id := range order {
		//mark watched episodes
		if sm.IsMonitored(id) {
			prefix = "*"
		} else {
			prefix = " "
		}
		
		fmt.Printf("%d %s%s\n", id, prefix, showsMap[id].TitleOriginal)
	}

}

func SearchShow(s string) {
        defer func(now time.Time) {
                log.Debugf("Searching shows took: %v\n", time.Since(now))
        }(time.Now())

	shows, err := myshows.SearchShow(s)
	must(err)
	
	order := make([]int, len(shows))
	showsMap := make(map[int]myshows.Show, 0)
	
	for n, show := range shows {
		showsMap[show.Id] = show
		order[n] = show.Id
	}
	
	sort.Ints(order)
	
	for _, id := range order {
		fmt.Printf("%-7d %-33s %.2f\n", showsMap[id].Id, showsMap[id].TitleOriginal, showsMap[id].Rating)
	}
}

func ShowTopEpisodes(t int) {
        defer func(now time.Time) {
                log.Debugf("Get top episodes took: %v\n", time.Since(now))
        }(time.Now())

	shows, err := myshows.GetTopShows(t)
	must(err)
	
	for n, s := range shows {
		fmt.Printf("%3d %-33s %.2f   (id: %d)\n", n + 1, s.Show.TitleOriginal, s.Show.Rating, s.Show.Id)
	}
}

func GetShowInfo(id int) {
	defer func(now time.Time) {
                log.Debugf("Get show info took: %v\n", time.Since(now))
        }(time.Now())
        
        show, err := myshows.GetShowById(id)
        must(err)
        
        desc := getImdbDesc(show.Imdb)
        fmt.Printf("Name:     %s\nYear:     %d\nSeasons:  %d\nGenres:   ", show.TitleOriginal, show.Year, show.Seasons)
        for _, genreId := range show.Genres {
        	fmt.Printf("%s, ", Genres[genreId])
        }
        fmt.Printf("\n")
        fmt.Printf("Rating:   %.2f\nStatus:   %s\n", show.Rating, show.Status)
	fmt.Printf("Coutry:   %s\n", show.Country)        
	fmt.Printf("Desc: %s\n", desc)
}


func getShowList() []myshows.ShowDesc {
	ret, err := myshows.GetShowList(sm.Token)
	if err != nil {
		renewToken()
		ret, err = myshows.GetShowList(sm.Token)
	}
	if err != nil {
		log.Fatalln(err)
	}
	return ret
}

func getUnwatchedEpisodes() []myshows.EpisodeDesc {
	ret, err := myshows.GetNextEpisodes(sm.Token)
	if err != nil {
		log.Debugf("Failed to get episode list from 1st attempt (%s). Going to renew token\n", err)
		renewToken()
		log.Debugln("Attemting to get episodes second time")
		ret, err = myshows.GetNextEpisodes(sm.Token)
	}
	if err != nil {
		log.Fatalln(err)
	}
	//log.Debugln("Successfuly got episode list")
	return ret
}

func renewToken() {
	defer func(now time.Time) {
		log.Debugf("Token renewal took: %v\n", time.Since(now))
	}(time.Now())
	
	token, err := myshows.GetToken(cfg.MyShows.Id, cfg.MyShows.Secret, cfg.MyShows.User, cfg.MyShows.Password)
	must(err)
	sm.Token = token
}
