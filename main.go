package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	tui "github.com/marcusolsson/tui-go"
)

type result struct {
	title string
	href  string
}

const AZLYRICS_URL = "https://search.azlyrics.com/search.php"

var artist = flag.String("artist", "", "Artist Name")
var song = flag.String("song", "", "Song Name")

func isLyricURL(in string) bool {
	url, err := url.Parse(in)
	if err != nil {
		log.Fatal(err)
	}
	for _, seg := range strings.Split(url.Path, "/") {
		if seg == "lyrics" {
			return true
		}
	}
	return false
}

func main() {

	flag.Parse()

	if *artist == "" || *song == "" {
		fmt.Fprintf(os.Stderr, "error: missing artist/song info")
		os.Exit(1)
	}

	// Build the URL
	target, err := url.Parse(AZLYRICS_URL)
	if err != nil {
		log.Fatal(err)
	}

	query := target.Query()
	query.Set("q", *song+" "+*artist)
	target.RawQuery = query.Encode()

	doc, err := goquery.NewDocument(target.String())
	if err != nil {
		log.Fatal(err)
	}

	results := make([]result, 0)

	doc.Find("table a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if isLyricURL(href) {
			results = append(results, result{s.Text(), href})
		}
	})

	library := tui.NewList()
	library.SetFocused(true)
	library.SetSelected(0)

	for _, result := range results {
		library.AddItems(result.title)
	}

	status := tui.NewStatusBar("Select a song from the list")
	root := tui.NewVBox(library, status)

	ui := tui.New(root)

	library.OnItemActivated(func(t *tui.List) {
		ui.Update(func() {

			chosen := results[t.Selected()]

			doc, err = goquery.NewDocument(chosen.href)
			if err != nil {
				log.Fatal(err)
			}

			lyrics := doc.Find("div.ringtone").NextAllFiltered("div").First()
			label := tui.NewLabel(strings.TrimSpace(lyrics.Text()))
			s := tui.NewScrollArea(label)
			scrollBox := tui.NewVBox(s)
			scrollBox.SetTitle(chosen.title + " " + "Lyrics")
			scrollBox.SetBorder(true)

			ui.SetKeybinding("Up", func() { s.Scroll(0, -1) })
			ui.SetKeybinding("k", func() { s.Scroll(0, -1) })

			ui.SetKeybinding("Down", func() { s.Scroll(0, 1) })
			ui.SetKeybinding("j", func() { s.Scroll(0, 1) })

			ui.SetKeybinding("Ctrl+U", func() { s.Scroll(0, -10) })
			ui.SetKeybinding("Ctrl+D", func() { s.Scroll(0, 10) })

			root.Remove(0)
			status.SetText("Press q to quit")
			root.Prepend(scrollBox)

		})
	})

	ui.SetKeybinding("Esc", func() { ui.Quit() })
	ui.SetKeybinding("q", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		panic(err)
	}

}
