package main

import (
	"fmt"
	"math"
	"gonum.org/v1/gonum/stat/distuv"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/vg"
)

type ThreePointEstimate struct {
	lowPoint float64
	midPoint float64
	highPoint float64
}

type RiskEventInput struct {
	estimate ThreePointEstimate
	likelihood float64
}

type RiskItemInput struct {
	estimate ThreePointEstimate
}

type Distribution struct {
	points []float64
	minPoint float64
	maxPoint float64
}

func main() {
	// copied from the "labor" item in the excel spreadsheet
	testRiskItemInputs := []RiskItemInput {
		RiskItemInput {
			ThreePointEstimate {
				lowPoint: 7.5,
				midPoint: 10.0, 
				highPoint: 15.0,
			},
		},
		RiskItemInput {
			ThreePointEstimate {
				lowPoint: 6.0,
				midPoint: 8.5, 
				highPoint: 12.5,
			},
		},
		RiskItemInput {
			ThreePointEstimate {
				lowPoint: 5.0,
				midPoint: 7.5, 
				highPoint: 13.0,
			},
		},
		RiskItemInput {
			ThreePointEstimate {
				lowPoint: 7.0,
				midPoint: 11.0, 
				highPoint: 15.0,
			},
		},
	}

	dist, minPoint, maxPoint := logisticPoolingRiskItems(testRiskItemInputs)

	// fmt.Println(dist)

	plotDistribution(dist, minPoint, maxPoint)

}


func computeMean(est ThreePointEstimate) float64 {
	return (est.lowPoint + 4*est.midPoint + est.highPoint)/6
}

func computeStdDev(est ThreePointEstimate) float64 {
	mean := computeMean(est)
	return math.Sqrt((mean - est.lowPoint)*(est.highPoint - mean)/7)
}

func linearPoolingRiskItems(inputs []RiskItemInput) ([]float64, float64, float64) {
	ests := []ThreePointEstimate{}

	for i := 0; i < len(inputs); i++ {
		ests = append(ests, inputs[i].estimate)
	}

	return linearPooling(ests)
}

func linearPoolingRiskEvents(inputs []RiskEventInput) ([]float64, float64, float64, float64) {
	ests := []ThreePointEstimate{}
	sumLikelihood := float64(0)

	for i := 0; i < len(inputs); i++ {
		ests = append(ests, inputs[i].estimate)
		sumLikelihood += inputs[i].likelihood
	}

	dist, minPoint, maxPoint := linearPooling(ests)

	return dist, minPoint, maxPoint, sumLikelihood / float64(len(inputs))
}


func linearPooling(inputs []ThreePointEstimate) ([]float64, float64, float64) {
	inputBases, minPoint, maxPoint := constructInputBases(inputs)

	pooledBasis := []float64{}

	for i := 0; i < len(inputBases[0]); i++ {
		point := float64(0)
		for j := 0; j < len(inputBases); j++ {
			point += inputBases[j][i]
		}

		pooledBasis = append(pooledBasis, point / float64(len(inputBases)))
	}

	return pooledBasis, minPoint, maxPoint
}


func logisticPoolingRiskItems(inputs []RiskItemInput) ([]float64, float64, float64) {
	ests := []ThreePointEstimate{}

	for i := 0; i < len(inputs); i++ {
		ests = append(ests, inputs[i].estimate)
	}

	return logisticPooling(ests)
}

func logisticPoolingRiskEvents(inputs []RiskEventInput) ([]float64, float64, float64, float64) {
	ests := []ThreePointEstimate{}
	sumLikelihood := float64(0)

	for i := 0; i < len(inputs); i++ {
		ests = append(ests, inputs[i].estimate)
		sumLikelihood += inputs[i].likelihood
	}

	dist, minPoint, maxPoint := logisticPooling(ests)

	return dist, minPoint, maxPoint, sumLikelihood /float64(len(inputs))
}


func logisticPooling(inputs []ThreePointEstimate) ([]float64, float64, float64) {
	inputBases, minPoint, maxPoint := constructInputBases(inputs)

	pooledBasis := []float64{}

	for i := 0; i < len(inputBases[0]); i++ {
		point := 1.0
		for j := 0; j < len(inputBases); j++ {
			point *= inputBases[j][i]
		}

		pooledBasis = append(pooledBasis, math.Pow(point, float64(1) / float64(len(inputBases))))
	}

	return pooledBasis, minPoint, maxPoint
}

func constructInputBases(inputs []ThreePointEstimate) ([][]float64, float64, float64) {

	inputBases := [][]float64{}
	minPoint, maxPoint := determineOutputRange(inputs)
	interval := (maxPoint - minPoint) / 1000

	for i := 0; i < len(inputs); i++ {
		mean := computeMean(inputs[i])
		std := computeStdDev(inputs[i])
		dist := distuv.Normal {
			Mu: mean,
			Sigma: std,
		}

		basis := []float64{}

		for j := minPoint; j < maxPoint; j += interval  {
			point := dist.CDF(j + interval) - dist.CDF(j)
			basis = append(basis, point)
		}

		inputBases = append(inputBases, basis)
	}
	return inputBases, minPoint, maxPoint
}

func determineOutputRange(inputs []ThreePointEstimate) (float64, float64) {
	minPoint := computeMean(inputs[0]) - 8 * computeStdDev(inputs[0])
	maxPoint := computeMean(inputs[0]) + 8 * computeStdDev(inputs[0])

	for i := 1; i < len(inputs); i++ {
		mean := computeMean(inputs[i])
		std := computeStdDev(inputs[i])

		min := mean - 8 * std
		max := mean + 8 * std

		if(min < minPoint) {
			minPoint = min
			maxPoint = max
		}
	}

	return minPoint, maxPoint
}

func plotDistribution(points []float64, minPoint float64, maxPoint float64) {
	p := plot.New()

	p.Title.Text = "Test plot" 

	pts := make(plotter.XYs, len(points))
	for i := range pts {
		pts[i].X = minPoint + float64(i)*((maxPoint - minPoint) / float64(len(points)))
		pts[i].Y = points[i]
	}

	fmt.Println(pts)

	s, err := plotter.NewScatter(pts)

	if err != nil {
        panic(err)
    }

    p.Add(s)

    if err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}

