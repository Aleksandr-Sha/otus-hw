package hw02unpackstring

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

// / это особая ситуация
// состояние последнего символа для повторения

const BackSlashCode = 92

const (
	digit  = iota
	letter // Тут возможно поправить
	backSlash
)

type Character struct {
	typeValue int
	value     rune
}

func newCharacter(r rune) *Character {
	if unicode.IsDigit(r) {
		return &Character{typeValue: digit, value: r}
	}

	if r == BackSlashCode {
		return &Character{typeValue: backSlash, value: r}
	}

	return &Character{typeValue: letter, value: r}
}

type Unpacker struct {
	str        []rune
	currentLet *Character
	index      int
	builder    strings.Builder
}

func newUnpacker(str string) (*Unpacker, error) {
	runes := []rune(str)

	if isDigit(runes[0]) {
		return nil, fmt.Errorf("first letter must not be a digit")
	}

	return &Unpacker{runes, newCharacter(runes[0]), 1, strings.Builder{}}, nil
}

func (u *Unpacker) getNext() *Character {
	return newCharacter(u.str[u.index])
}

//нужно как-то запомнить состояние когда идёт \

func (u *Unpacker) unpack() (string, error) {
	for u.index < len(u.str) {
		next := u.getNext()

		switch next.typeValue {
		case digit:
		}
	}

	return "", nil
}

func Unpack(str string) (string, error) {
	if len(str) == 0 {
		return "", nil
	}

	unpacker, err := newUnpacker(str)

	if err != nil {
		return "", err
	}

	result, err := unpacker.unpack()

	if err != nil {
		return "", err
	}

	return result, nil
}

func isDigit(letter rune) bool {
	if letter >= '0' && letter <= '9' {
		return true
	}

	return false
}
