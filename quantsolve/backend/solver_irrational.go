package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

// RadicalTerm stores the math inside a root, its sign, and its degree (2=sqrt, 3=cube root)
type RadicalTerm struct {
	Coeffs     map[int]float64
	Sign       float64
	RootDegree float64
}

// IrrationalSide separates an infinite number of roots/fractional exponents from the outer math
type IrrationalSide struct {
	Radicals    []RadicalTerm
	OuterCoeffs map[int]float64
	Multiplier  float64
}

// parseIrrationalSide safely extracts MULTIPLE roots before applying string formatters!
func parseIrrationalSide(sideStr string, multiplier float64, varName *string) IrrationalSide {
	side := IrrationalSide{
		Radicals:    make([]RadicalTerm, 0),
		OuterCoeffs: make(map[int]float64),
		Multiplier:  multiplier,
	}

	// 🌟 THE FIX: Forgive missing parentheses! Automatically turns "sqrtx" into "sqrt(x)"
	missingParenRe := regexp.MustCompile(`sqrt([a-z0-9.]+)`)
	sideStr = missingParenRe.ReplaceAllString(sideStr, "sqrt($1)")

	impliedOneRe := regexp.MustCompile(`(^|[^0-9.])([a-z])`)

	// --- REGEX 1: Find standard sqrt(...) ---
	reSqrt := regexp.MustCompile(`([+-]?)\s*sqrt\(([^)]+)\)`)
	matchesSqrt := reSqrt.FindAllStringSubmatch(sideStr, -1)

	for _, match := range matchesSqrt {
		signStr := strings.TrimSpace(match[1])
		radStr := match[2]

		sign := 1.0
		if signStr == "-" {
			sign = -1.0
		}

		radStr = impliedOneRe.ReplaceAllString(radStr, "${1}1${2}")
		radStr = distributeDivision(radStr)
		radStr = expandBrackets(radStr)
		radStr = strings.ReplaceAll(radStr, "(", "")
		radStr = strings.ReplaceAll(radStr, ")", "")

		radCoeffs := make(map[int]float64)
		v1 := extractPolyCoeffs(radStr, 1.0, radCoeffs)
		if v1 != "" { *varName = v1 }

		side.Radicals = append(side.Radicals, RadicalTerm{Coeffs: radCoeffs, Sign: sign, RootDegree: 2.0})
		sideStr = strings.Replace(sideStr, match[0], "", 1)
	}

	// --- REGEX 2: Find fractional exponents like (x+2)^(1/3) ---
	reFrac := regexp.MustCompile(`([+-]?)\s*\(([^)]+)\)\^\(1\/(\d+)\)`)
	matchesFrac := reFrac.FindAllStringSubmatch(sideStr, -1)

	for _, match := range matchesFrac {
		signStr := strings.TrimSpace(match[1])
		radStr := match[2]
		degreeStr := match[3]

		sign := 1.0
		if signStr == "-" { sign = -1.0 }

		degree, _ := strconv.ParseFloat(degreeStr, 64)
		if degree == 0 { degree = 1 } // Failsafe

		radStr = impliedOneRe.ReplaceAllString(radStr, "${1}1${2}")
		radStr = distributeDivision(radStr)
		radStr = expandBrackets(radStr)
		radStr = strings.ReplaceAll(radStr, "(", "")
		radStr = strings.ReplaceAll(radStr, ")", "")

		radCoeffs := make(map[int]float64)
		v1 := extractPolyCoeffs(radStr, 1.0, radCoeffs)
		if v1 != "" { *varName = v1 }

		// Add it to our memory with the correct Root Degree!
		side.Radicals = append(side.Radicals, RadicalTerm{Coeffs: radCoeffs, Sign: sign, RootDegree: degree})
		sideStr = strings.Replace(sideStr, match[0], "", 1)
	}

	// Now apply string fixes safely to the remaining outer math
	sideStr = impliedOneRe.ReplaceAllString(sideStr, "${1}1${2}")
	sideStr = distributeDivision(sideStr)
	sideStr = expandBrackets(sideStr)
	sideStr = strings.ReplaceAll(sideStr, "(", "")
	sideStr = strings.ReplaceAll(sideStr, ")", "")

	v2 := extractPolyCoeffs(sideStr, 1.0, side.OuterCoeffs)
	if v2 != "" { *varName = v2 }

	return side
}

// Eval calculates 'y' for 'x', correctly handling negative bases for odd vs even roots
func (s IrrationalSide) Eval(x float64) (float64, bool) {
	sum := evaluatePolynomial(s.OuterCoeffs, x)

	for _, rad := range s.Radicals {
		radVal := evaluatePolynomial(rad.Coeffs, x)

		// Domain check: Even roots (like 2, 4) cannot handle negative numbers!
		isEvenRoot := math.Mod(rad.RootDegree, 2) == 0
		if isEvenRoot && radVal < 0 {
			return 0, false // Domain Error
		}

		// Mathematical trick: Go's math.Pow() returns NaN for negative bases.
		// For odd roots (like ^(1/3)), we force the absolute value, then re-apply the negative sign!
		rootVal := math.Pow(math.Abs(radVal), 1.0/rad.RootDegree)
		if radVal < 0 {
			rootVal = -rootVal
		}

		sum += rad.Sign * rootVal
	}

	return s.Multiplier * sum, true
}

// THE RADICAL ENGINE: Safely finds roots while avoiding extraneous solutions
func solveIrrational(equation string, conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte("> Initiating Universal Root Engine..."))

	equation = strings.ToLower(strings.ReplaceAll(equation, " ", ""))

	parts := strings.Split(equation, "=")
	if len(parts) != 2 { return }

	varName := "x"
	left := parseIrrationalSide(parts[0], 1.0, &varName)
	right := parseIrrationalSide(parts[1], -1.0, &varName)

	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Radical boundaries detected. Domain scanning variable '%s'...", varName)))

	var roots []float64
	step := 0.1
	checkCount := 0

	for x := -100.0; x < 100.0; x += step {
		checkCount++
		y1L, ok1L := left.Eval(x)
		y1R, ok1R := right.Eval(x)
		y2L, ok2L := left.Eval(x + step)
		y2R, ok2R := right.Eval(x + step)

		ok1 := ok1L && ok1R
		ok2 := ok2L && ok2R

		y1 := y1L + y1R
		y2 := y2L + y2R

		if ok1 && math.Abs(y1) < 0.001 {
			isDuplicate := false
			for _, r := range roots {
				if math.Abs(r-x) < 0.01 { isDuplicate = true; break }
			}
			if !isDuplicate { roots = append(roots, x) }
			
		} else if ok1 && ok2 && y1*y2 <= 0 {
			low, high := x, x+step
			for i := 0; i < 20; i++ {
				checkCount++
				mid := (low + high) / 2
				ymidL, okmidL := left.Eval(mid)
				ymidR, okmidR := right.Eval(mid)

				if !okmidL || !okmidR {
					low = mid 
					continue
				}

				ymid := ymidL + ymidR

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
		conn.WriteMessage(websocket.TextMessage, []byte("No real solutions found, or answers were extraneous and rejected."))
	}

	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("STATS:%d", checkCount)))
}