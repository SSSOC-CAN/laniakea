/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10

Copyright (C) 2015-2018 Lightning Labs and The Lightning Network Developers

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package utils

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	bg "github.com/SSSOCPaulCote/blunderguard"
)

const (
	ErrFloatLargerThanOne = bg.Error("float is larger than or equal to 1")
)

var (
	letters     = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	pluginRegEx = `([a-zA-Z0-9_-]+)\:((?:datasource)|(?:controller))\:([a-zA-Z0-9_-]+(\.[a-zA-Z0-9]+)?)$`
)

// FileExists reports whether the named file or directory exists.
// This function is taken from https://github.com/lightningnetwork/lnd
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// UniqueFileName creates a unique file name if the provided one exists
func UniqueFileName(path string) string {
	counter := 1
	for FileExists(path) {
		ext := filepath.Ext(path)
		if counter > 1 && counter < 11 {
			path = path[:len(path)-len(ext)-4] + " (" + strconv.Itoa(counter) + ")" + ext
		} else if counter >= 11 {
			path = path[:len(path)-len(ext)-5] + " (" + strconv.Itoa(counter) + ")" + ext
		} else {
			path = path[:len(path)-len(ext)] + " (" + strconv.Itoa(counter) + ")" + ext
		}
		counter++
	}
	return path
}

// RandSeq generates a random string of length n
func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// NormalizeToNDecimalPlace will take any float below 1 and get the factor to transform it to 1 with equivalent decimal places
func NormalizeToNDecimalPlace(oldF float64) (float64, error) {
	if oldF >= 1 {
		return 0, ErrFloatLargerThanOne
	}
	s := fmt.Sprintf("%f", oldF)
	newS := strings.Replace(s, ".", "", -1)
	if oldF < 0 {
		newS = strings.Replace(newS, "-", "", -1)
	}
	i := 0
	for {
		if newS[0] == byte('0') {
			newS = strings.Replace(newS, "0", "", 1)
			i++
		} else {
			break
		}
	}
	newS = newS[:1] + "." + newS[1:]
	newF, err := strconv.ParseFloat(newS, 64)
	if err != nil {
		return 0, err
	}
	factor := oldF / newF
	factor = math.Round(factor*math.Pow(10, float64(i))) / math.Pow(10, float64(i))
	return factor, nil
}

// NumDecPlaces returns the number of decimal places a float is specified to
func NumDecPlaces(v float64) int {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i > -1 {
		return len(s) - i - 1
	}
	return 1
}

// VerifyPluginStringFormat checks if the plugin format in the config is in the acceptable format
func VerifyPluginStringFormat(s string) bool {
	match, _ := regexp.MatchString(pluginRegEx, s)
	return match
}
