package main

import (
	"fmt"
	"slices"
	"sort"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type Probability struct {
	symbol string
	value  float64
}

type Boundary struct {
	symbol string
	bottom float64
	top    float64
}

func calculateProbabilities(textToEncode string, collator *collate.Collator) (probabilities []Probability) {
	runedText := []rune(textToEncode)
	oneSymbolLength := 1.0 / float64(len(runedText))
	sort.Slice(runedText, func(i, j int) bool {
		return collator.CompareString(string(runedText[i]), string(runedText[j])) < 0
	})

	addProb := make(map[rune]float64)
	for _, rt := range runedText {
		addProb[rt] += float64(oneSymbolLength)
	}

	for _, rt := range runedText {
		if !slices.Contains(probabilities, Probability{string(rt), addProb[rt]}) {
			probabilities = append(probabilities, Probability{string(rt), addProb[rt]})
		}
	}
	return
}

func getSegments(probs []Probability) (segments []Boundary) {
	bottomBound := 0.0
	topBound := 0.0
	for _, p := range probs {
		bottomBound = topBound
		topBound = p.value + bottomBound
		segments = append(segments, Boundary{p.symbol, bottomBound, topBound})
	}
	return
}

func findBoundary(boundaries []Boundary, s string) (float64, float64) {
	for _, v := range boundaries {
		if v.symbol == s {
			return v.bottom, v.top
		}
	}
	panic(fmt.Errorf("No boundary found"))
}

func encodeString(segments []Boundary, textToEncode string) (boundaries []Boundary, encoded float64) {
	bot, top := 0.0, 1.0
	for _, c := range textToEncode {
		interval := top - bot
		symbol := string(c)
		segBot, segTop := findBoundary(segments, symbol)

		newBot := bot + interval*segBot
		top = bot + interval*segTop

		bot = newBot
		boundaries = append(boundaries, Boundary{symbol, bot, top})
	}
	return boundaries, boundaries[len(boundaries)-1].bottom
}

func decodeString(encodedValue float64, decodedLength int, segments []Boundary) (decodedString string) {
	currentInterval := [2]float64{0.0, 1.0}
	for len([]rune(decodedString)) < decodedLength {
		for _, s := range segments {
			newLower := currentInterval[0] + (currentInterval[1]-currentInterval[0])*s.bottom
			newUpper := currentInterval[0] + (currentInterval[1]-currentInterval[0])*s.top

			if newLower <= encodedValue && newUpper > encodedValue {
				decodedString += s.symbol
				currentInterval[0] = newLower
				currentInterval[1] = newUpper
				break
			}
		}
	}
	return
}

func main() {

	collator := collate.New(language.Ukrainian)
	testStirngs := []string{
		"ІНФОРМАЦІЯ",
		"КЛІШОВ_М_Р",
	}

	for _, s := range testStirngs {
		fmt.Println("--------------------TEXT ENCODE+DECODE-----------------------")
		fmt.Printf("Text to encode: %v\n\n", s)
		printResult(s, collator)
	}
}

func printResult(text string, collator *collate.Collator) {
	probabilities := calculateProbabilities(text, collator)
	fmt.Println("Appearance probabilities")
	for _, p := range probabilities {
		fmt.Printf("Symbol: %v | Value: %v\n", p.symbol, p.value)
	}

	segments := getSegments(probabilities)
	fmt.Println("-------------------------\nSegments")
	for _, s := range segments {
		fmt.Printf("Symbol: %v | %v < X < %v\n", s.symbol, s.bottom, s.top)
	}

	boundaries, encodedValue := encodeString(segments, text)
	fmt.Println("-------------------------\nBoundaries")
	for _, b := range boundaries {
		fmt.Printf("Symbol: %v | %v - %v\n", b.symbol, b.bottom, b.top)
	}

	fmt.Printf("\nFinal encoded value: %v\n", encodedValue)

	decodedText := decodeString(encodedValue, len([]rune(text)), segments)
	fmt.Printf("Decoded text: %v\n\n", decodedText)
}
