package main

import (
	"flag"
	"fmt"
	"log"
	"time"

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

func (aTaskCard TrelloCard) printAsCSV() {
	fmt.Print(aTaskCard.card.Name, ", ")
	fmt.Print(aTaskCard.card.Due.String(), ", ")
	for _, label := range aTaskCard.card.Labels {
		fmt.Print("#", label.Name, ", ")
	}
	fmt.Print(aTaskCard.durationInDay)
	fmt.Println()
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
		date := aTaskCard.Due.Format(time.RFC822)
		aList := mapAnnotatedCards[date]
		if aList == nil {
			aList = []TrelloCard{}
		}
		aList = append(aList, TrelloCard{aTaskCard, 0.0})
		mapAnnotatedCards[date] = aList
	}
	return mapAnnotatedCards
}

func computeDurations(mappedCards map[string][]TrelloCard) map[string][]TrelloCard {
	for t, list := range mappedCards {
		list = computeDuration(list)
		mappedCards[t] = list
	}
	return mappedCards
}

func computeDuration(dailyCards []TrelloCard) (result []TrelloCard) {
	nbTaskThatDay := float64(len(dailyCards))
	for i := range dailyCards {
		dailyCards[i].durationInDay = 1.0 / nbTaskThatDay
	}
	return dailyCards
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
	durationPerProject := map[string]float64{}
	for _, aList := range annotatedCards {
	CardLoop:
		for _, aTaskCard := range aList {
			for _, label := range aTaskCard.card.Labels {
				durationPerProject[label.Name] += aTaskCard.durationInDay
				continue CardLoop
			}
		}
	}

	totalDays := 0.0
	for projectName, durationInDay := range durationPerProject {
		fmt.Print(projectName, "\t", durationInDay, "\n")
		totalDays += durationInDay
	}
	fmt.Print("---------------------------\n")
	fmt.Print("Total\t", totalDays, "\n")

}
