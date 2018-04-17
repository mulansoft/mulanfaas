package utils

import (
	log "github.com/Sirupsen/logrus"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	// letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterBytes   = "1234567890"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func IntToString(n int) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}

func Int32ToString(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}

func Uint64ToString(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func StringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.WithFields(log.Fields{
			"s":   s,
			"err": err,
		}).Error("StringToInt error")
		return 0
	}
	return i
}

func StringToUint64(s string) uint64 {
	v64, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		log.WithFields(log.Fields{
			"s":   s,
			"err": err,
		}).Error("StringToUint64 error")
		return 0
	}
	return v64
}

func StringToInt64(s string) int64 {
	v64, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		log.WithFields(log.Fields{
			"s":   s,
			"err": err,
		}).Error("StringToInt64 error")
		return 0
	}
	return v64
}

func TrimAllWhitespace(str string) string {
	return strings.Join(strings.Fields(str), "")
}
