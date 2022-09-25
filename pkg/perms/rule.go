package perms

import (
	"fmt"
	"strings"
)

type Rule struct {
	Actor        string
	Relationship string
	Object       string
}

func (r *Rule) ActorMatches(parameters map[string]string, actor string) bool {
	return r.matches(parameters, r.Actor, actor)
}
func (r *Rule) RelationshipMatches(parameters map[string]string, relationship string) bool {
	return r.matches(parameters, r.Relationship, relationship)
}
func (r *Rule) ObjectMatches(parameters map[string]string, object string) bool {
	return r.matches(parameters, r.Object, object)
}

func (r *Rule) matches(parameters map[string]string, needed string, def string) bool {
	result := needed
	if parameters != nil {
		for key, val := range parameters {
			result = strings.ReplaceAll(result, fmt.Sprintf("{%s}", key), val)
		}
	}
	all := strings.Split(strings.ToLower(result), ",")
	val := strings.ToLower(def)
	for _, a := range all {
		if isMatch(val, a) {
			return true
		}
	}
	return false
}

func isMatch(s string, p string) bool {
	runeInputArray := []rune(s)
	runePatternArray := []rune(p)
	if len(runeInputArray) > 0 && len(runePatternArray) > 0 {
		if runePatternArray[len(runePatternArray)-1] != '*' && runePatternArray[len(runePatternArray)-1] != '?' && runeInputArray[len(runeInputArray)-1] != runePatternArray[len(runePatternArray)-1] {
			return false
		}
	}
	return isMatchUtil([]rune(s), []rune(p), 0, 0, len([]rune(s)), len([]rune(p)))
}
func isMatchUtil(input, pattern []rune, inputIndex, patternIndex int, inputLength, patternLength int) bool {
	if inputIndex == inputLength && patternIndex == patternLength {
		return true
	} else if patternIndex == patternLength {
		return false
	} else if inputIndex == inputLength {
		if pattern[patternIndex] == '*' && restPatternStar(pattern, patternIndex+1, patternLength) {
			return true
		} else {
			return false
		}
	}

	if pattern[patternIndex] == '*' {
		return isMatchUtil(input, pattern, inputIndex, patternIndex+1, inputLength, patternLength) ||
			isMatchUtil(input, pattern, inputIndex+1, patternIndex, inputLength, patternLength)
	}
	if pattern[patternIndex] == '?' {
		return isMatchUtil(input, pattern, inputIndex+1, patternIndex+1, inputLength, patternLength)
	}
	if inputIndex < inputLength {
		if input[inputIndex] == pattern[patternIndex] {
			return isMatchUtil(input, pattern, inputIndex+1, patternIndex+1, inputLength, patternLength)
		} else {
			return false
		}
	}
	return false
}
func restPatternStar(pattern []rune, patternIndex int, patternLength int) bool {
	for patternIndex < patternLength {
		if pattern[patternIndex] != '*' {
			return false
		}
		patternIndex++
	}
	return true
}
