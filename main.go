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

		// Данные корректны — сбрасываем ошибки
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

		// Load Average
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
		}

		// Memory usage
		memPerc := memUsed / memTotal * 100
		if memPerc > 80 {
			fmt.Printf("Memory usage too high: %.0f%%\n", memPerc)
		}

		// Disk free space
		diskFree := diskTotal - diskUsed
		if diskUsed/diskTotal > 0.9 {
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", diskFree/1024/1024)
		}

		// Network
		if netUsed/netTotal > 0.9 {
			freeMbit := (netTotal - netUsed) * 8 / 1_000_000
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeMbit)
		}

		time.Sleep(time.Second)
	}
}
