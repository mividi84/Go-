package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	const url = "http://srv.msk01.gigacorp.local/_stats"
	const pollInterval = 1 * time.Second
	var errorCount int

	for {
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(pollInterval)
			continue
		}

		errorCount = 0 // сброс ошибок при успешном запросе

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(pollInterval)
			continue
		}

		data := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(data) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(pollInterval)
			continue
		}

		loadAvg, _ := strconv.ParseFloat(data[0], 64)
		memTotal, _ := strconv.ParseUint(data[1], 10, 64)
		memUsed, _ := strconv.ParseUint(data[2], 10, 64)
		diskTotal, _ := strconv.ParseUint(data[3], 10, 64)
		diskUsed, _ := strconv.ParseUint(data[4], 10, 64)
		netTotal, _ := strconv.ParseUint(data[5], 10, 64)
		netUsed, _ := strconv.ParseUint(data[6], 10, 64)

		// Load Average
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
		}

		// Memory usage
		if memTotal > 0 {
			memPercent := float64(memUsed) / float64(memTotal) * 100
			if memPercent > 80 {
				fmt.Printf("Memory usage too high: %.0f%%\n", memPercent)
			}
		}

		// Disk space
		if diskTotal > 0 {
			diskPercent := float64(diskUsed) / float64(diskTotal) * 100
			if diskPercent > 90 {
				freeMb := int((diskTotal - diskUsed) / 1024 / 1024)
				fmt.Printf("Free disk space is too low: %d Mb left\n", freeMb)
			}
		}

		// Network bandwidth
		if netTotal > 0 {
			netPercent := float64(netUsed) / float64(netTotal) * 100
			if netPercent > 90 {
				freeMbit := int((float64(netTotal-netUsed) * 8) / (1024 * 1024))
				fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbit)
			}
		}

		time.Sleep(pollInterval)
	}
}
