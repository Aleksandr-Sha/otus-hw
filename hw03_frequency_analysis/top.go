package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

var punctuationMarksFromEdgesPattern = regexp.MustCompile(`^\p{P}+|\p{P}+$`)

type wordCounter struct {
	text     string
	countMap map[string]int
	result   []string
}

func newWordCounter(text string) *wordCounter {
	return &wordCounter{text: text, countMap: make(map[string]int), result: make([]string, 0)}
}

func (w *wordCounter) getTop10Word() []string {
	for _, word := range strings.Fields(w.text) {
		if word == "-" {
			continue
		}

		w.processAndCountWord(word)
	}

	w.sortWords()

	return w.getTop10FromResult()
}

func (w *wordCounter) processAndCountWord(word string) {
	word = punctuationMarksFromEdgesPattern.ReplaceAllString(strings.ToLower(word), "")

	if _, ok := w.countMap[word]; ok {
		w.countMap[word]++
	} else {
		w.countMap[word] = 1
		w.result = append(w.result, word)
	}
}

func (w *wordCounter) sortWords() {
	sort.Slice(w.result, func(i, j int) bool {
		iCount := w.countMap[w.result[i]]
		jCount := w.countMap[w.result[j]]

		if iCount == jCount {
			if strings.Compare(w.result[i], w.result[j]) < 0 {
				return true
			}

			return false
		}

		return iCount > jCount
	})
}

func (w *wordCounter) getTop10FromResult() []string {
	if len(w.result) < 10 {
		return w.result[:len(w.result):len(w.result)]
	} else {
		return w.result[:10:10]
	}
}

func Top10(text string) []string {
	result := make([]string, 0)

	if text == "" {
		return result
	}

	newWordCounter := newWordCounter(text)

	return newWordCounter.getTop10Word()
}
