package main

import (
	"encoding/json"
	"fmt"

	"github.com/motaz/redisaccess"

	"net/http"
	"strconv"
	"time"

	"github.com/motaz/codeutils"
)

func getUserKey(w http.ResponseWriter, req *http.Request) (userKey string) {

	userCookie, err := req.Cookie("userkey")
	secondCookie, err2 := req.Cookie("secondkey")
	var secondKey string

	if err == nil {
		userKey = userCookie.Value
		secondKey = userKey
	} else if userKey == "" {
		if err2 == nil {
			userKey = secondCookie.Value
		} else {
			userKey = setCookies(w, req)
		}
	}
	if err2 != nil && err == nil {
		secondKey = userKey
		expiration2 := time.Now().Add(time.Hour * 24 * 100)
		cookie2 := http.Cookie{Name: "secondkey", Value: secondKey, Expires: expiration2}

		http.SetCookie(w, &cookie2)

	}
	return
}

func setCookies(w http.ResponseWriter, r *http.Request) (userKey string) {

	expiration := time.Now().Add(time.Hour * 24 * 90)
	expiration2 := time.Now().Add(time.Hour * 24 * 100)
	userKey = strconv.Itoa(codeutils.GetRandom(10000000))
	cookie := http.Cookie{Name: "userkey", Value: userKey, Expires: expiration}
	cookie2 := http.Cookie{Name: "secondkey", Value: userKey, Expires: expiration2}

	http.SetCookie(w, &cookie)
	http.SetCookie(w, &cookie2)
	return
}

func writeLog(event string) {
	codeutils.WriteToLog(event, "uploaded")
}

func searchDirectory(userkey string, searchkey, sharekey string, files *[]FileInfoType) {

	keys, err := redisaccess.GetKeys(FILE_INFO_KEY + "*")

	if err != nil {
		fmt.Println("Error: ", err.Error())
	} else {
		for _, key := range keys {
			infoStr, _, _ := redisaccess.GetValue(key)
			var info FileInfoType
			err := json.Unmarshal([]byte(infoStr), &info)
			if err != nil {
				fmt.Println("Error: ", err.Error())
			} else {

				if ((userkey != "" && info.UserKey == userkey) ||
					(searchkey != "" && info.ShareKey == searchkey)) ||
					(sharekey != "" && info.ShareKeyHash == sharekey) && info.Filename != "" {
					*files = append(*files, info)
				}
			}
		}
	}

}

func suggestEntryID(akey string) (entry string) {

	anum := codeutils.GetRandom(100000)
	hash := codeutils.GetMD5(fmt.Sprintf("%s-%s-%d", akey, time.Now().String(), anum))
	entry = hash[:20]
	return
}
