package main

type (
	spring struct {
		torque90, torque0 float64
	}
	actuator struct {
		model   string
		torque  float64
		springs [8]spring
	}
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
	stepToOptimal   = 2
)

var (
	result        ActResult
	resultList    []ActResult
	ok, broken    bool
	safetyFactor  float64
	sfStruct      safetyFactorSR
	actuatorsList = [14]actuator{
		{
			model:  "PA2",
			torque: 8.3,
			springs: [8]spring{
				{4.9, 3.4},
				{5.8, 4.1},
				{6.8, 4.7},
				{7.8, 5.4},
				{8.8, 6.1},
				{9.7, 6.8},
				{10.7, 7.4},
				{11.7, 8.1},
			},
		},
		{
			model:  "PA3",
			torque: 14.7,
			springs: [8]spring{
				{8.5, 5.5},
				{10.2, 6.7},
				{11.8, 7.8},
				{13.5, 8.9},
				{15.2, 10},
				{16.9, 11.1},
				{18.6, 12.2},
				{20.3, 13.3},
			},
		},
		{
			model:  "PA7",
			torque: 29.1,
			springs: [8]spring{
				{17.3, 11.1},
				{20.8, 13.3},
				{24.2, 15.5},
				{27.7, 17.7},
				{31.1, 19.9},
				{34.6, 22.1},
				{38.1, 24.3},
				{41.5, 26.5},
			},
		},
		{
			model:  "PA11",
			torque: 45.7,
			springs: [8]spring{
				{28.9, 18.3},
				{34.7, 22},
				{40.4, 25.7},
				{46.2, 29.4},
				{52, 33},
				{57.8, 36.7},
				{63.5, 40.4},
				{69.3, 44},
			},
		},
		{
			model:  "PA16",
			torque: 66.5,
			springs: [8]spring{
				{39.4, 25.3},
				{47.3, 30.4},
				{55.2, 35.4},
				{63.1, 40.5},
				{71, 45.5},
				{78.8, 50.6},
				{86.7, 55.6},
				{94.6, 60.7},
			},
		},
		{
			model:  "PA25",
			torque: 107,
			springs: [8]spring{
				{65.6, 41},
				{78.7, 49.3},
				{91.8, 57.5},
				{105, 65.7},
				{118, 74},
				{131, 82},
				{144, 90.3},
				{157, 98.5},
			},
		},
		{
			model:  "PA33",
			torque: 138,
			springs: [8]spring{
				{85.2, 52.5},
				{98.9, 62.9},
				{115, 73.4},
				{132, 83.9},
				{148, 94.4},
				{165, 105},
				{181, 115},
				{198, 126},
			},
		},
		{
			model:  "PA52",
			torque: 217,
			springs: [8]spring{
				{129, 82.3},
				{155, 98.7},
				{181, 115},
				{206, 132},
				{232, 148},
				{258, 165},
				{284, 181},
				{310, 197},
			},
		},
		{
			model:  "PA68",
			torque: 283,
			springs: [8]spring{
				{166, 112},
				{199, 135},
				{233, 157},
				{266, 179},
				{299, 202},
				{332, 224},
				{365, 247},
				{399, 269},
			},
		},
		{
			model:  "PA91",
			torque: 383,
			springs: [8]spring{
				{237, 158},
				{284, 190},
				{322, 221},
				{379, 253},
				{426, 284},
				{474, 316},
				{521, 347},
				{569, 379},
			},
		},
		{
			model:  "PA120",
			torque: 531,
			springs: [8]spring{
				{315, 212},
				{378, 255},
				{441, 297},
				{504, 340},
				{567, 382},
				{630, 425},
				{693, 467},
				{756, 510},
			},
		},
		{
			model:  "PA220",
			torque: 935,
			springs: [8]spring{
				{616, 434},
				{740, 521},
				{863, 608},
				{986, 695},
				{1109, 852},
				{1233, 869},
				{1356, 955},
				{1479, 1042},
			},
		},
		{
			model:  "PA320",
			torque: 1347,
			springs: [8]spring{
				{783, 567},
				{939, 680},
				{1096, 794},
				{1252, 907},
				{1409, 1021},
				{1565, 1134},
				{1722, 1247},
				{1878, 1361},
			},
		},
		{
			model:  "PA560",
			torque: 2350,
			springs: [8]spring{
				{1334, 1017},
				{1600, 1221},
				{1867, 1424},
				{2134, 1628},
				{2400, 1831},
				{2667, 2035},
				{2934, 2238},
				{3200, 2442},
			},
		},
	}

	minSpringPressure = [8]float64{2.5, 2.5, 3, 3.5, 4, 4.5, 5, 5.5}
	maxSpringPressure = [8]float64{5, 5.5, 6, 7, 8, 8, 8, 8}
)
