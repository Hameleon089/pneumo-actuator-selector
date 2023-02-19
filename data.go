package main

type (
	safetyFactorSR struct {
		factorBTO, factorETO, factorBTC, factorETC float64
	}
	ActResult struct {
		model                                              string
		springNum                                          int
		torque, torqueBTO, torqueETO, torqueBTC, torqueETC float64
		optimal                                            bool
	}
)

const (
	initPressure    = 2.5
	maxPressure     = 8.0
	maxTorque       = 7521
	maxSafetyFactor = 3.5
	stepToOptimal   = 1
)

var (
	result       ActResult
	resultList   []ActResult
	ok, broken   bool
	safetyFactor float64
	sfStruct     safetyFactorSR
)
