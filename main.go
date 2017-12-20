package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	"gopkg.in/headzoo/surf.v1"
)

var (
	logger *log.Logger = log.New(ioutil.Discard, "", 0)
	debug              = false
)

type episode struct {
	Index int    `json:"index"`
	Title string `json:"title"`
	Url   string `json:"url"`
	Id    string `json:"id"`
}

type ByIndex []*episode

func (eps ByIndex) Len() int {
	return len(eps)
}

func (eps ByIndex) Swap(i, j int) {
	eps[i], eps[j] = eps[j], eps[i]
}

func (eps ByIndex) Less(i, j int) bool {
	return eps[i].Index < eps[j].Index
}

func saveToFile(bow *browser.Browser, filename string) error {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = bow.Download(w)
	if err != nil {
		return err
	}

	return nil
}

func download(url string, filename string, dir string) error {
	/*
		cmd := exec.Command("wget", url, "-O", dir+"/"+filename)
		err := cmd.Start()
		if err != nil {
			return err
		}

		err = cmd.Wait()
		if err != nil {
			return err
		}

		return nil
	*/
	return download_pipe(url, filename, dir)
}

func newEpisode(title, href string) *episode {
	title = strings.TrimSpace(title[strings.Index(title, "English Café"):])
	index, err := strconv.Atoi(title[14:])
	if err != nil {
		panic(err)
	}

	url := "https://www.lizhi.fm" + href
	id := strings.Split(href, "/")[2]
	return &episode{index, title, url, id}
}

// 1. get all hrefs of each episode
func fetchHrefs() []*episode {
	bow := surf.NewBrowser()
	bow.SetUserAgent(agent.Firefox())

	var eps []*episode
	// TODO: this number(92) and user ID should be set automatically
	for i := 92; i > 0; i-- {
		url := fmt.Sprintf("https://www.lizhi.fm/user/6362505/p/%d.html", i)
		logger.Println("== parse", url)
		err := bow.Open(url)
		if err != nil {
			saveToFile(bow, "err_log.html")
			panic(err)
		}
		sel := bow.Find("a")
		sel.Each(func(_ int, s *goquery.Selection) {
			title, ok := s.Attr("title")
			if ok && strings.Contains(title, "English Café") {
				href, _ := s.Attr("href")
				fmt.Printf("title=%s\nhref=%s\n", title, href)
				ep := newEpisode(title, href)
				eps = append(eps, ep)
			}
		})
		time.Sleep(time.Millisecond * 500)
	}
	sort.Sort(ByIndex(eps))

	// print into stdout in json format
	data, err := json.MarshalIndent(eps, " ", " ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	return eps
}

// 2. dowload all episodes
func downloadAllEpisodes(eps []*episode, filename string) {
	if eps == nil {
		eps = make([]*episode, 0)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(data, &eps)
		if err != nil {
			panic(err)
		}
	}
	for _, ep := range eps {
		fmt.Printf("== %s\n   %s\n", ep.Title, ep.Url)
		downloadEpisode(ep)
		time.Sleep(time.Second * 2)
	}
}

func downloadEpisode(ep *episode) {
	resp, err := http.Get("https://www.lizhi.fm/media/url/" + ep.Id)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	logger.Println(string(body))

	jdata := make(map[string]interface{})
	err = json.Unmarshal(body, &jdata)
	if err != nil {
		panic(err)
	}
	logger.Println(jdata)

	dataMap, _ := jdata["data"].(map[string]interface{})
	audioUrl, ok := dataMap["url"].(string)
	if !ok {
		panic("no audio url")
	}

	name := strings.Replace(ep.Title, " ", "_", -1)
	name = fmt.Sprintf("%03d_%s.mp3", ep.Index, name)
	fmt.Printf("   audio URL: %s\n", audioUrl)
	download(audioUrl, name, "episodes")
	fmt.Printf("== %s finished\n\n", ep.Title)
}

func init() {
	flag.BoolVar(&debug, "d", false, "enable debug log")
	flag.Parse()
	if debug {
		logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

func main() {
	// eps := fetchHrefs()
	downloadAllEpisodes(nil, "list.json")
}
