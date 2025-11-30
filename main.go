package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	const url = "http://srv.msk01.gigacorp.local/_stats"

	errorCount := 0

	for {
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != 200 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(1 * time.Second)
			continue
		}

		errorCount = 0

		scanner := bufio.NewScanner(resp.Body)
		if !scanner.Scan() {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			resp.Body.Close()
			time.Sleep(1 * time.Second)
			continue
		}
		line := scanner.Text()
		resp.Body.Close()

		parts := strings.Split(line, ",")
		if len(parts) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(1 * time.Second)
			continue
		}

		vals := make([]int64, 7)
		ok := true
		for i, p := range parts {
			num, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
			if err != nil {
				ok = false
				break
			}
			vals[i] = num
		}

		if !ok {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(1 * time.Second)
			continue
		}

		loadAvg := int(vals[0])
		memTotal := vals[1]
		memUsed := vals[2]
		diskTotal := vals[3]
		diskUsed := vals[4]
		netTotal := vals[5]
		netUsed := vals[6]

		// ========== Load Average ==========
		if loadAvg >= 30 {
			fmt.Printf("Load Average is too high: %d\n", loadAvg)
		}

		// ========== Memory ==========
		if memTotal > 0 {
			memPercent := float64(memUsed) / float64(memTotal) * 100
			if memPercent >= 80 {
				fmt.Printf("Memory usage too high: %d%%\n", int(memPercent))
			}
		}

		// ========== Disk ==========
		if diskTotal > 0 {
			diskPercent := float64(diskUsed) / float64(diskTotal) * 100
			if diskPercent > 90 {
				freeBytes := diskTotal - diskUsed
				freeMb := freeBytes / (1024 * 1024)
				fmt.Printf("Free disk space is too low: %d Mb left\n", freeMb)
			}
		}

		// ========== Network ==========
		if netTotal > 0 {
			netPercent := float64(netUsed) / float64(netTotal) * 100
			if netPercent > 90 {
				freeBytes := netTotal - netUsed
				freeMbit := (freeBytes * 8) / (1024 * 1024)
				fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbit)
			}
		}

		time.Sleep(1 * time.Second)
	}
}
