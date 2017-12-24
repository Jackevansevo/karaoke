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

// [TODO] Refactor out common logic
// [TODO] Options to capture STDOUT and pipe shit elsewhere (e.g. less)
// [TODO] Link with dbus

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
		fmt.Fprintln(os.Stderr, "error: missing artist/song info")
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
	var chosen result

	doc.Find("table a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if isLyricURL(href) {
			results = append(results, result{s.Text(), href})
		}
	})

	switch len(results) {
	case 0:
		fmt.Fprintln(os.Stderr, "error: song not found")
		os.Exit(1)
	case 1:
		chosen = results[0]
	default:
		list := tui.NewList()
		list.SetFocused(true)
		list.SetSelected(0)

		for _, result := range results {
			list.AddItems(result.title)
		}

		status := tui.NewStatusBar("Select a song from the list")
		root := tui.NewVBox(list, status)

		ui := tui.New(root)

		ui.SetKeybinding("Esc", func() { ui.Quit() })
		ui.SetKeybinding("q", func() { ui.Quit() })

		// Quit once selected
		list.OnItemActivated(func(t *tui.List) {
			chosen = results[t.Selected()]
			ui.Quit()
		})

		if err := ui.Run(); err != nil {
			panic(err)
		}
	}

	// [TODO] Devise way of printing stuff to STDOUT

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

	status := tui.NewStatusBar("Press q to quit")
	root := tui.NewVBox(scrollBox, status)
	ui := tui.New(root)

	ui.SetKeybinding("Up", func() { s.Scroll(0, -1) })
	ui.SetKeybinding("k", func() { s.Scroll(0, -1) })

	ui.SetKeybinding("Down", func() { s.Scroll(0, 1) })
	ui.SetKeybinding("j", func() { s.Scroll(0, 1) })

	ui.SetKeybinding("Ctrl+U", func() { s.Scroll(0, -10) })
	ui.SetKeybinding("Ctrl+D", func() { s.Scroll(0, 10) })

	ui.SetKeybinding("Esc", func() { ui.Quit() })
	ui.SetKeybinding("q", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		panic(err)
	}

}
