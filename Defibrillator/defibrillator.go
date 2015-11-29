package main

import "fmt"
import "os"
import "bufio"
import "math"

import "strings"

//import "strconv"

func getDistanceFromLatLonInKm(lat1, lon1, lat2, lon2 float64) float64 {
	R := 6371.0 // Radius of the earth in km
	dLat := deg2rad(lat2 - lat1)
	dLon := deg2rad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(deg2rad(lat1))*math.Cos(deg2rad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := R * c // Distance in km
	return d
}

func deg2rad(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	var LON float64
	scanner.Scan()
	fmt.Sscan(strings.Replace(scanner.Text(), ",", ".", -1), &LON)

	var LAT float64
	scanner.Scan()
	fmt.Sscan(strings.Replace(scanner.Text(), ",", ".", -1), &LAT)

	var N int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &N)

	fmt.Fprintf(os.Stderr, "%v %v %v\n", LON, LAT, N)

	var minDist float64 = -1
	result := "No defibrillator around!!"

	for i := 0; i < N; i++ {
		scanner.Scan()
		fmt.Fprint(os.Stderr, scanner.Text())
		DEFIB := strings.Split(scanner.Text(), ";")

		var defibLon, defibLat float64
		fmt.Sscan(strings.Replace(DEFIB[4], ",", ".", -1), &defibLon)
		fmt.Sscan(strings.Replace(DEFIB[5], ",", ".", -1), &defibLat)

		dist := getDistanceFromLatLonInKm(defibLat, defibLon, LAT, LON)

		fmt.Fprintf(os.Stderr, "%v %v %v dist=%v\n", defibLon, defibLat, DEFIB[1], dist)

		if minDist == -1 || dist < minDist {
			minDist = dist
			result = DEFIB[1]
			fmt.Fprint(os.Stderr, "  --> chosen\n")
		}

	}

	fmt.Println(result) // Write answer to stdout
}
