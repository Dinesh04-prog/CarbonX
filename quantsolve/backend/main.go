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

// Payload handles the JSON arriving from the React Rule Engine
type Payload struct {
	Equation    string `json:"equation"`
	Constraints string `json:"constraints"`
}

// THE POLITE PARSER: Handles equations and catches user errors
func parseEquation(input string) ([]Asset, int, error) {
	input = strings.ReplaceAll(input, " ", "")
	
	// Failsafe 1: Check for the equals sign
	parts := strings.Split(input, "=")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("Oops! I couldn't find an '=' sign. Please add it so I know your target budget (e.g., = 100).")
	}

	// Failsafe 2: Check if the budget is a real number
	target, err := strconv.Atoi(parts[1])
	if err != nil { 
		return nil, 0, fmt.Errorf("I had trouble reading the budget after the '=' sign ('%s'). Please make sure it is a whole number without letters.", parts[1]) 
	}

	// Failsafe 3: Prevent zero or negative budgets
	if target <= 0 {
		return nil, 0, fmt.Errorf("Your budget is currently set to %d. Please provide a budget greater than 0 so we can buy some assets!", target)
	}

	re := regexp.MustCompile(`(\d+)([a-zA-Z])`)
	matches := re.FindAllStringSubmatch(parts[0], -1)

	// Failsafe 4: Check if any variables actually exist
	if len(matches) == 0 { 
		return nil, 0, fmt.Errorf("I couldn't find any valid assets in your equation. Try attaching a cost to a letter, like '10x' or '5y'.") 
	}

	var assets []Asset
	for _, match := range matches {
		cost, _ := strconv.Atoi(match[1])
		
		// Failsafe 5: Prevent zero-cost items (The Infinity Trap!)
		if cost <= 0 {
			return nil, 0, fmt.Errorf("Asset '%s' has a cost of 0 (or negative). Free assets create infinite loops! Please give it a price of 1 or more.", match[2])
		}
		
		assets = append(assets, Asset{Name: match[2], Cost: cost})
	}
	return assets, target, nil
}

// THE RULES PARSER: Translates "x>5" into math limits
func parseConstraints(input string) (map[string]int, map[string]int) {
	minVals := make(map[string]int)
	maxVals := make(map[string]int)

	if input == "" { return minVals, maxVals }
	input = strings.ReplaceAll(input, " ", "")
	rules := strings.Split(input, ",")

	re := regexp.MustCompile(`([a-zA-Z]+)(>=|<=|>|<|=)(\d+)`)
	
	for _, rule := range rules {
		match := re.FindStringSubmatch(rule)
		if len(match) == 4 {
			v := match[1]
			op := match[2]
			val, _ := strconv.Atoi(match[3])

			switch op {
			case ">": minVals[v] = val + 1
			case ">=": minVals[v] = val
			case "<": maxVals[v] = val - 1
			case "<=": maxVals[v] = val
			case "=":
				minVals[v] = val
				maxVals[v] = val
			}
		}
	}
	return minVals, maxVals
}

// THE MATH ENGINE: Recursively finds solutions within the rules
func solveRecursive(assets []Asset, target int, currentCombo map[string]int, conn *websocket.Conn, found *bool, minVals map[string]int, maxVals map[string]int) {
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
			time.Sleep(20 * time.Millisecond)
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
		newCombo := make(map[string]int)
		for k, v := range currentCombo { newCombo[k] = v }
		newCombo[currentAsset.Name] = i

		leftover := target - (i * currentAsset.Cost)
		solveRecursive(assets[1:], leftover, newCombo, conn, found, minVals, maxVals)
	}
}

// THE WEBSOCKET CONTROLLER
func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil { return }
	defer conn.Close()

	for {
		var payload Payload
		err := conn.ReadJSON(&payload)
		if err != nil { break }

		assets, target, err := parseEquation(payload.Equation)
		if err != nil {
			// If the user made a mistake, stream the polite error back to React!
			conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			conn.WriteMessage(websocket.TextMessage, []byte("FINISHED"))
			continue
		}

		minVals, maxVals := parseConstraints(payload.Constraints)

		found := false
		solveRecursive(assets, target, make(map[string]int), conn, &found, minVals, maxVals)

		if !found {
			conn.WriteMessage(websocket.TextMessage, []byte("No whole number solutions exist within these constraints."))
		}
		conn.WriteMessage(websocket.TextMessage, []byte("FINISHED"))
	}
}

func main() {
	router := gin.Default()
	router.GET("/ws", handleWebSocket)
	fmt.Println("QuantSolve Master Engine running on http://localhost:8080")
	router.Run(":8080")
}