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
// THE PRE-PROCESSOR: Expands simple multipliers attached to brackets
// Converts "2*(5x+10y)" into "10x+20y"
func expandBrackets(input string) string {
	// Regex to find a number, an optional *, and a bracketed group
	// e.g. matches "2(5x+10y)" or "2*(5x+10y)"
	re := regexp.MustCompile(`(\d*\.?\d+)\*?\(([^)]+)\)`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		multiplierStr := submatches[1]
		insideBrackets := submatches[2]

		multiplier, _ := strconv.ParseFloat(multiplierStr, 64)

		// Now find all assets inside the bracket to multiply them
		assetRe := regexp.MustCompile(`([+-]?\d*\.?\d+)([a-zA-Z])`)
		
		expandedInside := assetRe.ReplaceAllStringFunc(insideBrackets, func(assetMatch string) string {
			assetSub := assetRe.FindStringSubmatch(assetMatch)
			costInside, _ := strconv.ParseFloat(assetSub[1], 64)
			variable := assetSub[2]
			
			newCost := multiplier * costInside
			
			// Format cleanly, adding a + if it's positive so it chains together well
			if newCost >= 0 {
				return fmt.Sprintf("+%.2f%s", newCost, variable)
			}
			return fmt.Sprintf("%.2f%s", newCost, variable)
		})

		return expandedInside
	})
}
// THE POLITE PARSER: Handles decimals (10.25) by converting them to cents (1025)
func parseEquation(input string) ([]Asset, int, error) {
	input = strings.ReplaceAll(input, " ", "")
	input = expandBrackets(input)
	
	// Failsafe 1: Check for the equals sign
	parts := strings.Split(input, "=")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("Oops! I couldn't find an '=' sign. Please add it so I know your target budget (e.g., = 100).")
	}

	// NEW DECIMAL LOGIC: Parse the target budget as a Float (decimal)
	rawTarget, err := strconv.ParseFloat(parts[1], 64)
	if err != nil { 
		return nil, 0, fmt.Errorf("I had trouble reading the budget after the '=' sign ('%s'). Please make sure it is a valid number.", parts[1]) 
	}

	// Failsafe 2: Prevent zero or negative budgets
	if rawTarget <= 0 {
		return nil, 0, fmt.Errorf("Your budget is currently set to %v. Please provide a budget greater than 0!", rawTarget)
	}

	// NEW DECIMAL LOGIC: Multiply by 100 to convert to whole integer cents
	targetInCents := int(rawTarget * 100)

	// NEW REGEX: \d*\.?\d+ captures whole numbers (10) AND decimals (10.25)
	re := regexp.MustCompile(`(\d*\.?\d+)([a-zA-Z])`)
	matches := re.FindAllStringSubmatch(parts[0], -1)

	// Failsafe 3: Check if any variables actually exist
	if len(matches) == 0 { 
		return nil, 0, fmt.Errorf("I couldn't find any valid assets in your equation. Try attaching a cost to a letter, like '10x' or '5.25y'.") 
	}

	var assets []Asset
	for _, match := range matches {
		// NEW DECIMAL LOGIC: Read the asset cost as a Float, then multiply by 100
		rawCost, _ := strconv.ParseFloat(match[1], 64)
		costInCents := int(rawCost * 100)
		
		// Failsafe 4: Prevent zero-cost items (The Infinity Trap!)
		if costInCents <= 0 {
			return nil, 0, fmt.Errorf("Asset '%s' has a cost of 0. Free assets create infinite loops! Please give it a valid price.", match[2])
		}
		
		assets = append(assets, Asset{Name: match[2], Cost: costInCents})
	}
	return assets, targetInCents, nil
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


// 1. ADD 'checkCount *int' TO THE END OF THE FUNCTION SIGNATURE
func solveRecursive(assets []Asset, target int, currentCombo map[string]int, conn *websocket.Conn, found *bool, minVals map[string]int, maxVals map[string]int, checkCount *int) {
	
	// Add to the counter every time the engine checks a base case
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
			time.Sleep(5 * time.Millisecond) // Lowered sleep to make it faster!
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
		// Add to the counter every time the engine branches
		*checkCount++ 
		
		newCombo := make(map[string]int)
		for k, v := range currentCombo { newCombo[k] = v }
		newCombo[currentAsset.Name] = i

		leftover := target - (i * currentAsset.Cost)
		solveRecursive(assets[1:], leftover, newCombo, conn, found, minVals, maxVals, checkCount) // Pass it down!
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

		// Inside handleWebSocket(), replace the solveRecursive call with this:

		minVals, maxVals := parseConstraints(payload.Constraints)

		found := false
		checkCount := 0 // Create the counter
		
		// Pass &checkCount to the engine
		solveRecursive(assets, target, make(map[string]int), conn, &found, minVals, maxVals, &checkCount)

		if !found {
			conn.WriteMessage(websocket.TextMessage, []byte("No whole number solutions exist within these constraints."))
		}
		
		// SEND THE FINAL STATS TO REACT BEFORE CLOSING
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("STATS:%d", checkCount)))
		conn.WriteMessage(websocket.TextMessage, []byte("FINISHED"))
	}
}

func main() {
	router := gin.Default()
	router.GET("/ws", handleWebSocket)
	fmt.Println("QuantSolve Master Engine running on http://localhost:8080")
	router.Run(":8080")
}