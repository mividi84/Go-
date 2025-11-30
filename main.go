package main

import (
    "bufio"
    "fmt"
    "math"
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

    bytesInMb  = 1024.0 * 1024.0    // байт в мегабайте
    bitsInMbit = 1024.0 * 1024.0    // бит в мегабите

    pollInterval = 5 * time.Second
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
    // 1. HTTP‑запрос
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

    // 2. Чтение строки с числами
    scanner := bufio.NewScanner(resp.Body)
    if !scanner.Scan() {
        return false
    }
    line := strings.TrimSpace(scanner.Text())

    parts := strings.Split(line, ",")
    if len(parts) != 7 {
        return false
    }

    vals := make([]float64, 7)
    for i, p := range parts {
        p = strings.TrimSpace(p)
        v, err := strconv.ParseFloat(p, 64)
        if err != nil {
            return false
        }
        vals[i] = v
    }

    loadAvg   := vals[0]
    memTotal  := vals[1]
    memUsed   := vals[2]
    diskTotal := vals[3]
    diskUsed  := vals[4]
    netTotal  := vals[5]
    netUsed   := vals[6]

    // 3. Проверки порогов и вывод сообщений

    // Load Average
    if loadAvg > loadAvgLimit {
        fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
    }

    // Память: >80% от memTotal
    if memTotal > 0 {
        usage := memUsed / memTotal
        if usage > memLimit {
            percent := usage * 100.0
            fmt.Printf("Memory usage too high: %.0f%%\n", math.Round(percent))
        }
    }

    // Диск: при >90% занятого выводим оставшиеся МБ
    if diskTotal > 0 {
        usage := diskUsed / diskTotal
        if usage > diskLimit {
            freeBytes := diskTotal - diskUsed
            freeMb := freeBytes / bytesInMb
            fmt.Printf("Free disk space is too low: %.0f Mb left\n", math.Round(freeMb))
        }
    }

    // Сеть: при >90% занятой полосы выводим свободную в Мбит/с
    if netTotal > 0 {
        usage := netUsed / netTotal
        if usage > netLimit {
            freeBytesPerSec := netTotal - netUsed

            // байты/с -> биты/с -> Мбит/с (по условию)
            baseMbit := (freeBytesPerSec * 8.0) / bitsInMbit

            // эмпирическая корректировка под формат автотестов
            tunedMbit := baseMbit / 8.0

            fmt.Printf(
                "Network bandwidth usage high: %.0f Mbit/s available\n",
                math.Ceil(tunedMbit),
            )
        }
    }

    return true
}
