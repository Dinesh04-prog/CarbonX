package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

// ExpTerm stores the math for an exponential chunk, like -3 * e^(2x+1)
type ExpTerm struct {
	Coeffs     map[int]float64
	Base       float64
	Multiplier float64
}

type ExpSide struct {
	Exps        []ExpTerm
	OuterCoeffs map[int]float64
	SideMult    float64
}

// parseExpSide safely extracts MULTIPLE exponential blocks before applying string formatters!
func parseExpSide(sideStr string, sideMult float64, varName *string) ExpSide {
	side := ExpSide{
		Exps:        make([]ExpTerm, 0),
		OuterCoeffs: make(map[int]float64),
		SideMult:    sideMult,
	}

	nakedExpRe := regexp.MustCompile(`\^([a-z])`)
	sideStr = nakedExpRe.ReplaceAllString(sideStr, "^($1)")

	implicitERe := regexp.MustCompile(`(\d*\.?\d+)e\^`)
	sideStr = implicitERe.ReplaceAllString(sideStr, "${1}*e^")

	impliedOneRe := regexp.MustCompile(`(^|[^0-9.])([a-z])`)

	reExp := regexp.MustCompile(`([+-]?)\s*(?:(\d*\.?\d*)\s*\*?\s*)?\(?(e|\d+\.?\d*)\^\(([^)]+)\)\)?`)
	matches := reExp.FindAllStringSubmatch(sideStr, -1)

	for _, match := range matches {
		signStr := strings.TrimSpace(match[1])
		multStr := strings.TrimSpace(match[2])
		baseStr := strings.TrimSpace(match[3])
		innerStr := strings.TrimSpace(match[4])

		mult := 1.0
		if multStr != "" {
			parsedMult, err := strconv.ParseFloat(multStr, 64)
			if err == nil {
				mult = parsedMult
			}
		}
		if signStr == "-" {
			mult = -mult
		}

		base := math.E
		if baseStr != "e" {
			parsedBase, err := strconv.ParseFloat(baseStr, 64)
			if err == nil {
				base = parsedBase
			}
		}

		innerStr = impliedOneRe.ReplaceAllString(innerStr, "${1}1${2}")
		innerStr = distributeDivision(innerStr)
		innerStr = expandBrackets(innerStr)
		innerStr = strings.ReplaceAll(innerStr, "(", "")
		innerStr = strings.ReplaceAll(innerStr, ")", "")

		innerCoeffs := make(map[int]float64)
		v1 := extractPolyCoeffs(innerStr, 1.0, innerCoeffs)
		if v1 != "" { *varName = v1 }

		side.Exps = append(side.Exps, ExpTerm{Coeffs: innerCoeffs, Base: base, Multiplier: mult})
		sideStr = strings.Replace(sideStr, match[0], "", 1)
	}

	sideStr = impliedOneRe.ReplaceAllString(sideStr, "${1}1${2}")
	sideStr = distributeDivision(sideStr)
	sideStr = expandBrackets(sideStr)
	sideStr = strings.ReplaceAll(sideStr, "(", "")
	sideStr = strings.ReplaceAll(sideStr, ")", "")

	v2 := extractPolyCoeffs(sideStr, 1.0, side.OuterCoeffs)
	if v2 != "" { *varName = v2 }

	return side
}

// Eval calculates 'y' for 'x'
func (s ExpSide) Eval(x float64) float64 {
	sum := evaluatePolynomial(s.OuterCoeffs, x)

	for _, exp := range s.Exps {
		expVal := evaluatePolynomial(exp.Coeffs, x)
		val := exp.Multiplier * math.Pow(exp.Base, expVal)
		sum += val
	}

	return s.SideMult * sum
}

// THE EXPONENTIAL ENGINE: Solves variable-in-exponent equations via approximation
func solveExponential(equation string, conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte("> Initiating Transcendental (Exponential) Engine..."))

	equation = strings.ToLower(strings.ReplaceAll(equation, " ", ""))

	parts := strings.Split(equation, "=")
	if len(parts) != 2 { return }

	varName := "x"
	left := parseExpSide(parts[0], 1.0, &varName)
	right := parseExpSide(parts[1], -1.0, &varName)

	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Growth curves detected. Executing Deep Scan on variable '%s'...", varName)))

	var roots []float64
	step := 0.1
	checkCount := 0

	for x := -50.0; x < 50.0; x += step {
		checkCount++
		y1L := left.Eval(x)
		y1R := right.Eval(x)
		y2L := left.Eval(x + step)
		y2R := right.Eval(x + step)

		if math.IsInf(y1L, 0) || math.IsInf(y1R, 0) || math.IsInf(y2L, 0) || math.IsInf(y2R, 0) { continue }

		y1 := y1L + y1R
		y2 := y2L + y2R

		// 🌟 THE FIX: The Asymptote Shield!
		// Replaced the loose Absolute Value check with a strict Sign Change check (X-axis crossing).
		if y1 == 0 {
			isDuplicate := false
			for _, r := range roots {
				if math.Abs(r-x) < 0.01 { isDuplicate = true; break }
			}
			if !isDuplicate { roots = append(roots, x) }
			
		} else if (y1 > 0 && y2 < 0) || (y1 < 0 && y2 > 0) {
			low, high := x, x+step
			for i := 0; i < 30; i++ { 
				checkCount++
				mid := (low + high) / 2
				ymid := left.Eval(mid) + right.Eval(mid)

				if (y1 > 0 && ymid < 0) || (y1 < 0 && ymid > 0) {
					high = mid
				} else {
					low = mid
					y1 = ymid
				}
			}
			root := (low + high) / 2

			isDuplicate := false
			for _, r := range roots {
				if math.Abs(r-root) < 0.01 { isDuplicate = true; break }
			}
			if !isDuplicate {
				if math.Abs(root) < 0.0001 { root = 0 }
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
		conn.WriteMessage(websocket.TextMessage, []byte("No real solutions found within the standard growth scan."))
	}

	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("STATS:%d", checkCount)))
}