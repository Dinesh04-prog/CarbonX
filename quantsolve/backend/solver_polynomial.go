package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

// THE EVALUATOR: Plugs an 'x' value into our mapped polynomial and returns 'y'
func evaluatePolynomial(coeffs map[int]float64, x float64) float64 {
	sum := 0.0
	for power, coef := range coeffs {
		sum += coef * math.Pow(x, float64(power))
	}
	return sum
}

// THE POLYNOMIAL EXTRACTOR: Maps standard forms like -3m^4, +m, and now +c/3 !
func extractPolyCoeffs(side string, multiplier float64, coeffs map[int]float64) string {
	if side == "" || side == "0" {
		return ""
	}
	if !strings.HasPrefix(side, "+") && !strings.HasPrefix(side, "-") {
		side = "+" + side
	}

	re := regexp.MustCompile(`([+-]\d*\.?\d*)([a-z])?(?:\^(\d+))?(?:\/(\d*\.?\d+))?`)
	matches := re.FindAllStringSubmatch(side, -1)

	varName := ""

	for _, match := range matches {
		coefStr := match[1]
		letter := match[2]
		powStr := match[3]
		divStr := match[4] 

		if letter != "" {
			varName = letter
		}

		if (coefStr == "+" || coefStr == "-") && letter == "" {
			continue
		}

		if coefStr == "+" {
			coefStr = "+1"
		}
		if coefStr == "-" {
			coefStr = "-1"
		}

		coef, err := strconv.ParseFloat(coefStr, 64)
		if err != nil {
			continue
		}

		if divStr != "" {
			divisor, err := strconv.ParseFloat(divStr, 64)
			if err == nil && divisor != 0 {
				coef /= divisor
			}
		}

		coef *= multiplier

		power := 0
		if letter != "" {
			if powStr != "" {
				power, _ = strconv.Atoi(powStr)
			} else {
				power = 1
			}
		}

		coeffs[power] += coef
	}
	return varName
}

// THE NUMERICAL ENGINE: Uses Bisection Method to find roots
func solvePolynomial(equation string, conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte("> Initiating Algebraic Numeric Engine..."))

	equation = strings.ToLower(strings.ReplaceAll(equation, " ", ""))

	impliedOneRe := regexp.MustCompile(`(^|[^0-9.])([a-z])`)
	equation = impliedOneRe.ReplaceAllString(equation, "${1}1${2}")

	equation = distributeDivision(equation)

	equation = expandBrackets(equation)
	equation = strings.ReplaceAll(equation, "(", "")
	equation = strings.ReplaceAll(equation, ")", "")

	parts := strings.Split(equation, "=")
	if len(parts) != 2 {
		return
	}

	coeffs := make(map[int]float64)
	varLeft := extractPolyCoeffs(parts[0], 1.0, coeffs)
	varRight := extractPolyCoeffs(parts[1], -1.0, coeffs)

	varName := "x"
	if varLeft != "" {
		varName = varLeft
	}
	if varRight != "" {
		varName = varRight
	}

	degree := 0
	for p, c := range coeffs {
		if math.Abs(c) > 0.0001 && p > degree {
			degree = p
		}
	}

	if degree <= 1 {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Standard Linear Algebra detected. Isolating variable '%s'...", varName)))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Degree %d polynomial detected. Executing Bisection Approximation...", degree)))
	}

	var roots []float64
	step := 0.1
	checkCount := 0

	for x := -100.0; x < 100.0; x += step {
		checkCount++
		y1 := evaluatePolynomial(coeffs, x)
		y2 := evaluatePolynomial(coeffs, x+step)

		if y1*y2 <= 0 {
			low, high := x, x+step
			for i := 0; i < 20; i++ {
				checkCount++
				mid := (low + high) / 2
				ymid := evaluatePolynomial(coeffs, mid)
				if y1*ymid <= 0 {
					high = mid
				} else {
					low = mid
					y1 = ymid
				}
			}
			root := (low + high) / 2

			isDuplicate := false
			for _, r := range roots {
				if math.Abs(r-root) < 0.01 {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				if math.Abs(root) < 0.0001 {
					root = 0
				}
				roots = append(roots, root)
			}
		}
	}

	if len(roots) > 0 {
		var out []string
		for i, r := range roots {
			if len(roots) == 1 {
				out = append(out, fmt.Sprintf("%s=%.4f", varName, r))
			} else {
				out = append(out, fmt.Sprintf("%s_%d=%.4f", varName, i+1, r))
			}
		}
		conn.WriteMessage(websocket.TextMessage, []byte("Solution Found: "+strings.Join(out, ", ")))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("No real solutions found in standard search space (-100 to 100)."))
	}

	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("STATS:%d", checkCount)))
}