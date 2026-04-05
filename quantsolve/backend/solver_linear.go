package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// THE RECURSIVE ENGINE: Now upgraded for continuous decimal fractions!
func solveRecursive(assets []Asset, target int, currentCombo map[string]int, conn *websocket.Conn, found *bool, minVals map[string]int, maxVals map[string]int, checkCount *int) {

	*checkCount++

	if len(assets) == 1 {
		lastAsset := assets[0]
		
		// 🌟 THE BREAKTHROUGH: No more integer modulo check! We calculate exact floats.
		exactCount := float64(target) / float64(lastAsset.Cost)

		// Check constraints (casting ints to floats to compare)
		if min, ok := minVals[lastAsset.Name]; ok && exactCount < float64(min) { return }
		if max, ok := maxVals[lastAsset.Name]; ok && exactCount > float64(max) { return }

		var parts []string
		for k, v := range currentCombo {
			parts = append(parts, fmt.Sprintf("%s=%d", k, v))
		}
		
		// Append the final exact decimal value (to 4 decimal places)
		parts = append(parts, fmt.Sprintf("%s=%.4f", lastAsset.Name, exactCount))

		conn.WriteMessage(websocket.TextMessage, []byte("Solution Found: "+strings.Join(parts, ", ")))
		time.Sleep(5 * time.Millisecond)
		*found = true
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