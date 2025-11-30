package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get("http://srv.msk01.gigacorp.local/_stats")
		if err != nil || resp.StatusCode != http.StatusOK {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(time.Second)
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(time.Second)
			continue
		}

		data := strings.Split(strings.TrimSpace(string(bodyBytes)), ",")
		if len(data) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(time.Second)
			continue
		}

		errorCount = 0

		vals := make([]float64, 7)
		for i, v := range data {
			vals[i], err = strconv.ParseFloat(v, 64)
			if err != nil {
				errorCount++
				if errorCount >= 3 {
					fmt.Println("Unable to fetch server statistic")
				}
				continue
			}
		}

		loadAvg := vals[0]
		memTotal := vals[1]
		memUsed := vals[2]
		diskTotal := vals[3]
		diskUsed := vals[4]
		netTotal := vals[5]
		netUsed := vals[6]

		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %d\n", int(loadAvg))
		}

		memPerc := int(memUsed * 100 / memTotal)
		if memPerc >= 80 {
			fmt.Printf("Memory usage too high: %d%%\n", memPerc)
		}

		if diskUsed*100/diskTotal > 90 {
			freeMb := int((diskTotal - diskUsed) / 1024 / 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeMb)
		}

		if netUsed*100/netTotal > 90 {
			freeNetMbit := int((netTotal - netUsed) / 1_000_000)
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeNetMbit)
		}

		time.Sleep(time.Second)
	}
}
