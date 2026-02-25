package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

var (
	punctuationMarksFromEdgesRegexp = regexp.MustCompile(`^\p{P}+|\p{P}+$`)
	dashesRegexp                    = regexp.MustCompile(`^-+$`)
)

type wordCounter struct {
	text          string
	wordsCountMap map[string]int
	wordsResult   []string
}

func newWordCounter(text string) *wordCounter {
	return &wordCounter{text: text, wordsCountMap: make(map[string]int), wordsResult: make([]string, 0)}
}

func (w *wordCounter) getTop10Word() []string {
	for _, word := range strings.Fields(w.text) {
		if word != "-" {
			w.processAndCountWord(word)
		}
	}

	w.sortWords()

	return w.getTop10FromResult()
}

func (w *wordCounter) processAndCountWord(word string) {
	if !dashesRegexp.MatchString(word) {
		word = punctuationMarksFromEdgesRegexp.ReplaceAllString(strings.ToLower(word), "")
	}

	if _, ok := w.wordsCountMap[word]; ok {
		w.wordsCountMap[word]++
	} else {
		w.wordsCountMap[word] = 1
		w.wordsResult = append(w.wordsResult, word)
	}
}

func (w *wordCounter) sortWords() {
	sort.Slice(w.wordsResult, func(i, j int) bool {
		iCount := w.wordsCountMap[w.wordsResult[i]]
		jCount := w.wordsCountMap[w.wordsResult[j]]

		if iCount == jCount {
			return strings.Compare(w.wordsResult[i], w.wordsResult[j]) < 0
		}

		return iCount > jCount
	})
}

func (w *wordCounter) getTop10FromResult() []string {
	if len(w.wordsResult) < 10 {
		return w.wordsResult[:len(w.wordsResult):len(w.wordsResult)]
	}

	return w.wordsResult[:10:10]
}

func Top10(text string) []string {
	if text == "" {
		return []string{}
	}

	wordCounter := newWordCounter(text)

	return wordCounter.getTop10Word()
}
