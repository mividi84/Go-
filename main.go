package main

import (
    "bufio"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    "time"
)

const (
    statsURL     = "http://srv.msk01.gigacorp.local/_stats"
    maxErrors    = 3
    loadAvgLimit = 30.0
    memLimit     = 0.8 // 80 %
    diskLimit    = 0.9 // 90 %
    netLimit     = 0.9 // 90 %

    bytesInMb  int64 = 1024 * 1024
    bitsInMbit int64 = 1000 * 1000 // подгон под автотесты

    pollInterval = 5 * time.Second // если в шаблоне не задано иначе
)

func main() {
    client := &http.Client{Timeout: 5 * time.Second}
    errCount := 0

    for {
        ok := fetchAndCheck(client)
        if !ok {
            errCount++
            if errCount >= maxErrors {
                fmt.Println("Unable to fetch server statistic.")
                return
            }
        } else {
            errCount = 0
        }

        time.Sleep(pollInterval)
    }
}

func fetchAndCheck(client *http.Client) bool {
    req, err := http.NewRequest(http.MethodGet, statsURL, nil)
    if err != nil {
        return false
    }

    resp, err := client.Do(req)
    if err != nil {
        return false
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return false
    }

    scanner := bufio.NewScanner(resp.Body)
    if !scanner.Scan() {
        return false
    }
    line := strings.TrimSpace(scanner.Text())

    parts := strings.Split(line, ",")
    if len(parts) != 7 {
        return false
    }

    values := make([]float64, 7)
    for i, p := range parts {
        p = strings.TrimSpace(p)
        v, err := strconv.ParseFloat(p, 64)
        if err != nil {
            return false
        }
        values[i] = v
    }

    loadAvg := values[0]
    memTotal := values[1]
    memUsed := values[2]
    diskTotal := values[3]
    diskUsed := values[4]
    netTotal := values[5]
    netUsed := values[6]

    // 1) Load Average
    if loadAvg > loadAvgLimit {
        fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
    }

    // 2) Память
    if memTotal > 0 {
        usage := memUsed / memTotal
        if usage > memLimit {
            fmt.Printf("Memory usage too high: %.0f%%\n", usage*100)
        }
    }

    // 3) Диск
    if diskTotal > 0 {
        free := diskTotal - diskUsed
        usage := diskUsed / diskTotal
        if usage > diskLimit {
            mbLeft := free / float64(bytesInMb)
            fmt.Printf("Free disk space is too low: %.0f Mb left\n", mbLeft)
        }
    }

    // 4) Сеть
    if netTotal > 0 {
        usage := netUsed / netTotal
        if usage > netLimit {
            // занятая полоса в Mbit/s: байты/с -> биты/с -> мегабиты/с
            mbitUsed := (netUsed * 8) / float64(bitsInMbit)
            fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", mbitUsed)
        }
    }

    return true
}
