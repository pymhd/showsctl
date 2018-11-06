package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	DataDir string `yaml:"data"`

	Telegram struct {
		Key      string `yaml:"key"`
		Watchers []int  `yaml:"notify"`
	} `yaml:"tlgrm"`

	MyShows struct {
		Id       string `yaml:"id"`
		Secret   string `yaml:"secret"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"myshows"`
}

func LoadCfg(f string) *Config {
	cfg := new(Config)

	fb, err := ioutil.ReadFile(f)
	//panics if err
	must(err)
	//panics if could not load
	must(yaml.Unmarshal(fb, cfg))

	return cfg
}
