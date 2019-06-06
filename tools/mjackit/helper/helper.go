package helper

import (
	"errors"
	"fmt"
)

//Hex converter functions taken from net/parse.go

const big = 0xFFFFFF

func Xtoi(s string) (n int, i int, ok bool) {
	n = 0
	for i = 0; i < len(s); i++ {
		if '0' <= s[i] && s[i] <= '9' {
			n *= 16
			n += int(s[i] - '0')
		} else if 'a' <= s[i] && s[i] <= 'f' {
			n *= 16
			n += int(s[i]-'a') + 10
		} else if 'A' <= s[i] && s[i] <= 'F' {
			n *= 16
			n += int(s[i]-'A') + 10
		} else {
			break
		}
		if n >= big {
			return 0, i, false
		}
	}
	if i == 0 {
		return 0, i, false
	}
	return n, i, true
}


func Xtoi2(s string, e byte) (byte, bool) {
	if len(s) > 2 && s[2] != e {
		return 0, false
	}
	n, ei, ok := Xtoi(s[:2])
	return byte(n), ok && ei == 2
}


func Select(prompt string, options []string) (index int, err error) {
	for i,o := range options {
		fmt.Printf("%d) %s\n", i+1, o)
	}

	fmt.Print(prompt)
	_, err = fmt.Scan(&index)
	if err != nil {
		return
	}
	if index < 1 || index > len(options) {
		err = errors.New("invalid option")
		return
	}

	index--
	return
}
