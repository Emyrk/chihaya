package main

import (
	"fmt"
	"net/http"
	"time"
)

// TRACKER to pound
const TRACKER string = "http://10.41.0.240:6881/announce"

var responseTimes chan float64

func main() {
	responseTimes = make(chan float64, 1000)
	go drain()
	for {
		announce()
	}
}

func drain() {
	for {
		select {
		case <-responseTimes:
		}
	}
}

func announce() {
	client := &http.Client{Timeout: time.Second * 5}

	url := makeURL()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	now := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		return
	}
	responseTimes <- timeDiff(now)
}

func makeURL() string {
	ih := fmt.Sprintf("%020s", "985a0750f122eea4b0eee2875c7210e3b7ca7861")
	id := fmt.Sprintf("%020s", "0")
	port := fmt.Sprintf("%d", 6881)
	downloaded := fmt.Sprintf("%d", 0)
	left := fmt.Sprintf("%d", 0)
	event := "started"

	url := fmt.Sprintf("%s?info_hash=%s&peer_id=%s&port=%s&downloaded=%s&left=%s&event=%s",
		TRACKER,
		ih,
		id,
		port,
		downloaded,
		left,
		event)

	return url
}

func timeDiff(start time.Time) float64 {
	diff := float64(time.Now().UnixNano()) - float64(time.Now().UnixNano())
	return diff
}

/*
http://some.tracker.com:999/announce
?info_hash=12345678901234567890
&peer_id=ABCDEFGHIJKLMNOPQRST
&ip=255.255.255.255
&port=6881
&downloaded=1234
&left=98765
&event=stopped
*/

/*
https://wiki.theory.org/BitTorrent_Tracker_Protocol
*/

/*
	Description
100	Invalid request type: client request was not a HTTP GET.
101	Missing info_hash.
102	Missing peer_id.
103	Missing port.
150	Invalid infohash: infohash is not 20 bytes long.
151	Invalid peerid: peerid is not 20 bytes long.
152	Invalid numwant. Client requested more peers than allowed by tracker.
200	info_hash not found in the database. Sent only by trackers that do not automatically include new hashes into the database.
500	Client sent an eventless request before the specified time.
900	Generic error.
*/
