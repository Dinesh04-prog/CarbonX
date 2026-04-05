package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// THE RECURSIVE ENGINE: Tracks node checks via a pointer
func solveRecursive(assets []Asset, target int, currentCombo map[string]int, conn *websocket.Conn, found *bool, minVals map[string]int, maxVals map[string]int, checkCount *int) {

	*checkCount++

	if len(assets) == 1 {
		lastAsset := assets[0]
		if target%lastAsset.Cost == 0 {
			count := target / lastAsset.Cost

			if min, ok := minVals[lastAsset.Name]; ok && count < min { return }
			if max, ok := maxVals[lastAsset.Name]; ok && count > max { return }

			currentCombo[lastAsset.Name] = count

			var parts []string
			for k, v := range currentCombo {
				parts = append(parts, fmt.Sprintf("%s=%d", k, v))
			}
			conn.WriteMessage(websocket.TextMessage, []byte("Solution Found: "+strings.Join(parts, ", ")))
			time.Sleep(5 * time.Millisecond)
			*found = true
		}
		return
	}

	currentAsset := assets[0]
	maxAmount := target / currentAsset.Cost

	if max, ok := maxVals[currentAsset.Name]; ok && maxAmount > max {
		maxAmount = max
	}

	minAmount := 0
	if min, ok := minVals[currentAsset.Name]; ok {
		minAmount = min
	}

	for i := minAmount; i <= maxAmount; i++ {
		*checkCount++

		newCombo := make(map[string]int)
		for k, v := range currentCombo {
			newCombo[k] = v
		}
		newCombo[currentAsset.Name] = i

		leftover := target - (i * currentAsset.Cost)
		solveRecursive(assets[1:], leftover, newCombo, conn, found, minVals, maxVals, checkCount)
	}
}