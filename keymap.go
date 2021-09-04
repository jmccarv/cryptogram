package main

import (
	"bytes"
	"fmt"
	"regexp"
)

type keyMap struct {
	key        [91]byte
	letterUsed [91]bool
}

func newKeyMap() keyMap {
	k := keyMap{}

	k.key['\''] = '\''
	k.letterUsed['\''] = true

	return k
}

func parseKeyMap(line []byte) (keyMap, bool) {
	k := newKeyMap()
	rxKey := regexp.MustCompile(`\s*([A-Z']+=[A-Z']+)(?:[ ,]|$)`)

	mappings := rxKey.FindAllSubmatch(bytes.ToUpper(line), -1)
	if len(mappings) == 0 {
		return k, false
	}

	for _, m := range mappings {
		kv := bytes.SplitN(m[1], []byte("="), 2)
		if len(kv) != 2 || len(kv[0]) != len(kv[1]) {
			fmt.Printf("Invalid Key Map %s in %s\n", string(m[1]), string(line))
			return k, false
		}

		// kv[0] is the left side of the '=', i.e. the encrypted character
		// kv[1] is the right side of the '=', i.e. the decrypted character
		for i, cc := range kv[0] {
			wc := kv[1][i]
			k.key[cc] = wc
			k.letterUsed[wc] = true
		}
	}

	return k, true
}

func (km keyMap) String() string {
	var ret string

	for i, c := range km.key {
		if c == 0 {
			continue
		}

		ret += fmt.Sprintf("%c=%c ", i, c)
	}

	return ret
}
