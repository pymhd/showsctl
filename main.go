package main

import (
	"flag"
	"fmt"
	log "github.com/pymhd/go-logging"
	"myshows"
	"sort"
	bot "tlgrm-bot"
)

var (
	//sm StorManager initialezed in stor.go - json cache
	cfg      *Config
	top      int
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
		ids := parseIds()
		log.Debugln(ids, "will be marked as mon")
		sm.AddToWatchlist(ids...)
	case unwatch:
		ids := parseIds()
		sm.DelFromWatchlist(ids...)
	case list:
		ListAllShows()
	case len(search) > 0:
		//
	case top > 0:
		ShowTopEpisodes(top)
	default:
		//if no flags (main purpose of script)
		NotifyUsers()
	}

}

func NotifyUsers() {
	u := getUnwatchedEpisodes()
	for _, obj := range u {
		if sm.IsMonitored(obj.Show.Id) {
			if !sm.IsSent(obj.Episode.Id) {
				cap := genNotificationCaption(obj)
				//sm.MarkAsSent(obj.Episode.Id)
				if skipflag {
					//just skip sending notification
					sm.MarkAsSent(obj.Episode.Id)
					log.Warningln(cap)
					continue
				}

				for _, chatId := range cfg.Telegram.Watchers {
					var err error
					switch {
					case len(obj.Episode.Image) > 0: //we have special image for episode
						_, _, err = bot.SendPhoto(cfg.Telegram.Key, chatId, obj.Episode.Image, cap, 0)
					case len(obj.Show.Image) > 0: //at least we have image for show
						_, _, err = bot.SendPhoto(cfg.Telegram.Key, chatId, obj.Show.Image, cap, 0)
					default: // well that's just plain text notification
						_, err = bot.SendTextMessage(cfg.Telegram.Key, chatId, cap, 0)
					}

					// no need to continue if problems with notifications
					// need to handle them first
					must(err)
				}
				sm.MarkAsSent(obj.Episode.Id)
			}
		}
	}
}

func ListAllShows() {
	shows := getShowList()

	order := make([]int, len(shows))
	showsMap := make(map[int]myshows.Show, 0)

	for n, s := range shows {
		showsMap[s.Show.Id] = s.Show
		order[n] = s.Show.Id
	}

	sort.Ints(order)
	for _, id := range order {
		fmt.Printf("%d : %s\n", id, showsMap[id].TitleOriginal)
	}

}

func ShowTopEpisodes(t int) {
	shows, err := myshows.GetTopShows(t)
	must(err)
	
	for n, s := range shows {
		fmt.Printf("%3d %-33s %.2f   (id: %d)\n", n + 1, s.Show.TitleOriginal, s.Show.Rating, s.Show.Id)
	}
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
		renewToken()
		ret, err = myshows.GetNextEpisodes(sm.Token)
	}
	if err != nil {
		log.Fatalln(err)
	}
	return ret
}

func renewToken() {
	token, err := myshows.GetToken(cfg.MyShows.Id, cfg.MyShows.Secret, cfg.MyShows.User, cfg.MyShows.Password)
	must(err)
	sm.Token = token
}
