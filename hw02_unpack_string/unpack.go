package hw02unpackstring

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const BackSlashCode = 92

var ErrInvalidString = errors.New("invalid string")

const (
	Digit = iota
	Other
)

type symbol struct {
	val        rune
	symbolType int
}

func (s *symbol) isDigit() bool {
	return s.symbolType == Digit
}

type unpacker struct {
	currentIndex   int
	runs           []rune
	builder        strings.Builder
	symbolForWrite *symbol
}

func newUnpacker(str string) *unpacker {
	return &unpacker{
		currentIndex: 0,
		runs:         []rune(str),
	}
}

func (u *unpacker) UnpackString() (string, error) {
	err := u.updateSymbolForWrite()
	if err != nil {
		return "", fmt.Errorf("read first symbol: %w", err)
	}

	for u.isReadNotAll() {
		if u.symbolForWrite.isDigit() {
			return "", fmt.Errorf(
				"first symbol before write iteration digit %s: %w",
				string(u.symbolForWrite.val),
				ErrInvalidString,
			)
		}

		if u.isReadFinish() {
			u.writeCurrentSymbolAndSetNew(nil)
			continue
		}

		nextSymbol, err := u.getNextSymbol()
		if err != nil {
			return "", fmt.Errorf("read next symbol: %w", err)
		}

		if nextSymbol.isDigit() {
			err := u.repeatAndWriteCurrentSymbol(nextSymbol)
			if err != nil {
				return "", fmt.Errorf("repeat and write: %w", err)
			}

			err = u.updateSymbolForWrite()
			if err != nil {
				return "", fmt.Errorf("after repeat and write: %w", err)
			}
		} else {
			u.writeCurrentSymbolAndSetNew(nextSymbol)
		}
	}

	return u.builder.String(), nil
}

func (u *unpacker) updateSymbolForWrite() error {
	newSymbolForWrite, err := u.getNextSymbol()
	if err != nil {
		return fmt.Errorf("update symbol for write: %w", err)
	}

	u.symbolForWrite = newSymbolForWrite
	return nil
}

func (u *unpacker) writeCurrentSymbolAndSetNew(newSymbol *symbol) {
	u.builder.WriteRune(u.symbolForWrite.val)
	u.symbolForWrite = newSymbol
}

func (u *unpacker) repeatAndWriteCurrentSymbol(countSymbol *symbol) error {
	countRepeat, err := strconv.Atoi(string(countSymbol.val))
	if err != nil {
		return fmt.Errorf("parse digit: %w", err)
	}

	repeatPart := strings.Repeat(string(u.symbolForWrite.val), countRepeat)

	u.builder.WriteString(repeatPart)

	return nil
}

func (u *unpacker) getNextSymbol() (*symbol, error) {
	if u.isReadFinish() {
		return nil, nil
	}

	newRune := u.runs[u.currentIndex]
	u.currentIndex++

	if unicode.IsDigit(newRune) {
		return &symbol{val: newRune, symbolType: Digit}, nil
	}

	if newRune == BackSlashCode {
		if !u.isReadFinish() {
			symbolAfterBackSlashCode := u.runs[u.currentIndex]
			u.currentIndex++

			if unicode.IsDigit(symbolAfterBackSlashCode) || symbolAfterBackSlashCode == BackSlashCode {
				return &symbol{val: symbolAfterBackSlashCode, symbolType: Other}, nil
			}
		}

		return nil, fmt.Errorf("get escaped symbol: %w", ErrInvalidString)
	}

	return &symbol{val: newRune, symbolType: Other}, nil
}

func (u *unpacker) isReadFinish() bool {
	return u.currentIndex == len(u.runs)
}

func (u *unpacker) isReadNotAll() bool {
	return u.symbolForWrite != nil
}

func Unpack(str string) (string, error) {
	if str == "" {
		return "", nil
	}

	unpacker := newUnpacker(str)

	unpackString, err := unpacker.UnpackString()
	if err != nil {
		return "", fmt.Errorf("unpack: %w", err)
	}

	return unpackString, nil
}
