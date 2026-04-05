package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Asset struct {
	Name string
	Cost int
}

// 1. THIS IS NEW: Tells Go how to read the JSON payload from React
type Payload struct {
	Equation    string `json:"equation"`
	Constraints string `json:"constraints"`
}

func parseEquation(input string) ([]Asset, int, error) {
	input = strings.ReplaceAll(input, " ", "")
	parts := strings.Split(input, "=")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("invalid format: missing '='")
	}

	target, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, fmt.Errorf("invalid target budget")
	}

	re := regexp.MustCompile(`(\d+)([a-zA-Z])`)
	matches := re.FindAllStringSubmatch(parts[0], -1)

	if len(matches) == 0 {
		return nil, 0, fmt.Errorf("could not detect any variables")
	}

	var assets []Asset
	for _, match := range matches {
		cost, _ := strconv.Atoi(match[1])
		assets = append(assets, Asset{Name: match[2], Cost: cost})
	}

	return assets, target, nil
}

// 2. THIS IS NEW: Parses rule constraints like "a>5"
// Updated parseEquation function
func parseEquation(input string, conn *websocket.Conn) ([]Asset, int, error) {
	// Send Step 1
	conn.WriteMessage(websocket.TextMessage, []byte("LOG: [STEP 1] Receiving raw input string..."))
	time.Sleep(300 * time.Millisecond) // Slight delay for visual effect

	input = strings.ReplaceAll(input, " ", "")
	
	// Send Step 2
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("LOG: [STEP 2] Tokenizing and cleaning whitespace -> '%s'", input)))
	time.Sleep(300 * time.Millisecond)

	parts := strings.Split(input, "=")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("invalid format: missing '='")
	}

	target, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, fmt.Errorf("invalid target budget")
	}

	// Send Step 3
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("LOG: [STEP 3] Budget isolated -> Target: %d", target)))
	time.Sleep(300 * time.Millisecond)

	re := regexp.MustCompile(`(\d+)([a-zA-Z])`)
	matches := re.FindAllStringSubmatch(parts[0], -1)

	if len(matches) == 0 {
		return nil, 0, fmt.Errorf("no variables found")
	}

	var assets []Asset
	var assetNames []string
	for _, match := range matches {
		cost, _ := strconv.Atoi(match[1])
		name := match[2]
		assets = append(assets, Asset{Name: name, Cost: cost})
		assetNames = append(assetNames, fmt.Sprintf("%s(cost:%d)", name, cost))
	}

	// Send Step 4
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("LOG: [STEP 4] Custom Regex extraction complete -> Variables: %s", strings.Join(assetNames, ", "))))
	time.Sleep(300 * time.Millisecond)
	
	// Send Step 5
	conn.WriteMessage(websocket.TextMessage, []byte("LOG: [STEP 5] Initializing Recursive DFS Algorithm..."))
	time.Sleep(200 * time.Millisecond)

	return assets, target, nil
}
// 3. UPDATED: Now enforces min/max limits during the recursion
func solveRecursive(assets []Asset, target int, currentCombo map[string]int, conn *websocket.Conn, found *bool, minVals map[string]int, maxVals map[string]int) {
	if len(assets) == 1 {
		lastAsset := assets[0]
		if target%lastAsset.Cost == 0 {
			count := target / lastAsset.Cost
			
			// Enforce constraints on the final asset
			if min, ok := minVals[lastAsset.Name]; ok && count < min { return }
			if max, ok := maxVals[lastAsset.Name]; ok && count > max { return }

			currentCombo[lastAsset.Name] = count
			
			var parts []string
			for k, v := range currentCombo {
				parts = append(parts, fmt.Sprintf("%s=%d", k, v))
			}
			conn.WriteMessage(websocket.TextMessage, []byte("Solution Found: "+strings.Join(parts, ", ")))
			time.Sleep(20 * time.Millisecond)
			*found = true
		}
		return
	}

	currentAsset := assets[0]
	maxAmount := target / currentAsset.Cost

	// Adjust max loop based on constraints
	if max, ok := maxVals[currentAsset.Name]; ok && maxAmount > max {
		maxAmount = max
	}
	
	// Adjust min loop based on constraints
	minAmount := 0
	if min, ok := minVals[currentAsset.Name]; ok {
		minAmount = min
	}

	for i := minAmount; i <= maxAmount; i++ {
		newCombo := make(map[string]int)
		for k, v := range currentCombo { newCombo[k] = v }
		newCombo[currentAsset.Name] = i

		leftover := target - (i * currentAsset.Cost)
		solveRecursive(assets[1:], leftover, newCombo, conn, found, minVals, maxVals)
	}
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil { return }
	defer conn.Close()

	for {
		// 4. THIS IS NEW: Reads the JSON object from React instead of plain string
		var payload Payload
		err := conn.ReadJSON(&payload)
		if err != nil { break }

		assets, target, err := parseEquation(payload.Equation)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
			conn.WriteMessage(websocket.TextMessage, []byte("FINISHED"))
			continue
		}

		minVals, maxVals := parseConstraints(payload.Constraints)

		found := false
		solveRecursive(assets, target, make(map[string]int), conn, &found, minVals, maxVals)

		if !found {
			conn.WriteMessage(websocket.TextMessage, []byte("No whole number solutions exist."))
		}
		conn.WriteMessage(websocket.TextMessage, []byte("FINISHED"))
	}
}

func main() {
	router := gin.Default()
	router.GET("/ws", handleWebSocket)
	fmt.Println("QuantSolve V3 (JSON Engine) running on http://localhost:8080")
	router.Run(":8080")
}