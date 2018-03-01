package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/adlio/trello"
)

var appKey = flag.String("appKey", "", "the appKey")
var token = flag.String("token", "", "the token")
var boardID = flag.String("board", "", "the board to extract from")
var finishedListName = flag.String("finish", "Fini", "Name of the finished list")

type TrelloCard struct {
	card          *trello.Card
	durationInDay float64
}

type cardByTime []*TrelloCard

func (a cardByTime) Len() int      { return len(a) }
func (a cardByTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a cardByTime) Less(i, j int) bool {
	return a[i].card.Due.Before(*a[j].card.Due) && !a[i].card.Due.Equal(*a[j].card.Due)
}

func (aTaskCard TrelloCard) printAsCSV() {
	d := aTaskCard.card.Due
	fmt.Printf("\"%d-%02d-%02d\",", d.Year(), d.Month(), d.Day())
	fmt.Printf("\"%.2f\",", aTaskCard.durationInDay)
	fmt.Print("\"", aTaskCard.card.Name, "\",\"")
	for _, label := range aTaskCard.card.Labels {
		fmt.Print("#", strings.Replace(label.Name, " ", "_", -1), ", ")
	}
	fmt.Println("\"")
}

type trelloList []*trello.Card

func (cards trelloList) printAsCSV() {
	for _, aTaskCard := range cards {
		TrelloCard{aTaskCard, 0.0}.printAsCSV()
	}
}

func groupByDate(cards []*trello.Card) map[string][]TrelloCard {
	mapAnnotatedCards := map[string][]TrelloCard{}
	for _, aTaskCard := range cards {
		d := aTaskCard.Due
		date := fmt.Sprintf("\"%d-%02d-%02d\", ", d.Year(), d.Month(), d.Day())
		aList := mapAnnotatedCards[date]
		if aList == nil {
			aList = []TrelloCard{}
		}
		aList = append(aList, TrelloCard{aTaskCard, 0.0})
		mapAnnotatedCards[date] = aList
	}
	return mapAnnotatedCards
}

func computeDurations(mappedCards map[string][]TrelloCard) {
	for t, list := range mappedCards {
		computeDuration(list)
		mappedCards[t] = list
	}
}

func computeDuration(dailyCards []TrelloCard) {
	nbTaskThatDay := float64(len(dailyCards))
	for i := range dailyCards {
		dailyCards[i].durationInDay = 1.0 / nbTaskThatDay
	}
}

func main() {
	flag.Parse()
	client := trello.NewClient(*appKey, *token)
	board, err := client.GetBoard(*boardID, trello.Defaults())
	if err != nil {
		log.Println("Cannot get board", err)
		return
	}

	lists, err := board.GetLists(trello.Defaults())
	if err != nil {
		log.Println("Cannot get lists", err)
		return
	}
	var finishedList *trello.List
	for _, aList := range lists {
		if aList.Name == *finishedListName {
			finishedList = aList
		}
	}
	if finishedList == nil {
		log.Println("Cannot find finished list:", *finishedListName)
	}
	cards, err := finishedList.GetCards(trello.Defaults())
	if err != nil {
		log.Println("Cannot get cards", err)
		return
	}

	annotatedCards := groupByDate(cards)
	computeDurations(annotatedCards)

	keys := make([]string, 0, len(annotatedCards))
	for k := range annotatedCards {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, aDay := range keys {
		cardsWithDuration := annotatedCards[aDay]
		for _, aCardWithDuration := range cardsWithDuration {
			aCardWithDuration.printAsCSV()
		}
	}
}
