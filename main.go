package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/exp/slices"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

func fatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func readFloatParam() (float64, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err == nil {
		input = strings.TrimSpace(input)
		param, err := strconv.ParseFloat(input, 64)
		return param, err
	}
	return 0, err
}

func inputParam(invitation string) float64 {
	fmt.Print(invitation)
	param, err := readFloatParam()
	for err != nil {
		fmt.Printf("Ошибка! Введены недопустимые данные.\n%s", invitation)
		param, err = readFloatParam()
	}
	return param
}

func readMode() (int, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err == nil {
		input = strings.TrimSpace(input)
		mode, err := strconv.Atoi(input)
		return mode, err
	}
	return 0, err
}

func inputMode() int {
	invitation := "Введите число (1-5): "
	fmt.Print(invitation)
	mode, err := readMode()
	for err != nil || mode > 5 || mode < 1 {
		fmt.Printf("Ошибка! Введены недопустимые данные.\n%s", invitation)
		mode, err = readMode()
	}
	return mode
}

func checkParams(pressure, nomTorque, safetyFactor float64) bool {
	if pressure <= maxPressure && pressure >= initPressure && nomTorque*safetyFactor <= maxTorque {
		return true
	}
	return false
}

//func DASelector(pressure, nomTorque, safetyFactor float64) (bool, ActResult) {
//	for _, act := range actuatorsList {
//		actuatorTorque := act.torque * pressure / initPressure
//		if safetyFactor < actuatorTorque/nomTorque || safetyFactor-actuatorTorque/nomTorque <= 0.01 {
//			return true, ActResult{
//				model:  act.model,
//				torque: actuatorTorque,
//			}
//		}
//	}
//	return false, ActResult{}
//}

func choosePressure(pressureList []float64, curPressure float64) float64 {
	if !slices.Contains(pressureList, curPressure) {
		minI := 0
		minDelta := math.Abs(pressureList[minI] - curPressure)
		for i := 1; i < len(pressureList); i++ {
			delta := math.Abs(pressureList[i] - curPressure)
			if delta < minDelta {
				minDelta = delta
				minI = i
			}
		}
		return pressureList[minI]
	}
	return curPressure
}

func DASelector(curPressure, nomTorque, safetyFactor float64) (bool, ActResult) {
	query := "SELECT DISTINCT pressure FROM da_air_torque"
	db, err := sql.Open("sqlite3", "db.sqlite3")
	fatalErr(err)
	defer db.Close()
	rows, err := db.Query(query)
	fatalErr(err)
	defer rows.Close()
	var pressureList []float64
	for rows.Next() {
		var pressure float64
		err = rows.Scan(&pressure)
		fatalErr(err)
		pressureList = append(pressureList, pressure)
	}

	chPressure := choosePressure(pressureList, curPressure)

	query = fmt.Sprintf(
		"SELECT torque, actuator_model FROM da_air_torque WHERE pressure = %f AND torque >= %f ORDER BY torque",
		chPressure, nomTorque*safetyFactor*chPressure/curPressure,
	)

	var (
		model  string
		torque float64
	)
	err = db.QueryRow(query).Scan(&torque, &model)
	if err == nil {
		return true, ActResult{
			model:  model,
			torque: torque * curPressure / chPressure,
		}
	}
	return false, ActResult{}
}

func SRSelector(curPressure, nomTorque float64, safetyFactorSR safetyFactorSR) (bool, []ActResult) {
	query := "SELECT DISTINCT pressure FROM sr_air_torque"
	db, err := sql.Open("sqlite3", "db.sqlite3")
	fatalErr(err)
	defer db.Close()
	rows, err := db.Query(query)
	fatalErr(err)
	defer rows.Close()
	var pressureList []float64
	for rows.Next() {
		var pressure float64
		err = rows.Scan(&pressure)
		fatalErr(err)
		pressureList = append(pressureList, pressure)
	}

	chPressure := choosePressure(pressureList, curPressure)
	torqueBTO := safetyFactorSR.factorBTO * nomTorque
	torqueETO := safetyFactorSR.factorETO * nomTorque
	torqueBTC := safetyFactorSR.factorBTC * nomTorque
	torqueETC := safetyFactorSR.factorETC * nomTorque

	query = fmt.Sprintf(
		`SELECT sra.torque_0, sra.torque_90, number, optimal, s.torque_90, s.torque_0, s.actuator_model 
		FROM sr_air_torque sra JOIN springs s on s.text_id = sra.spring_id 
		WHERE pressure = %f AND s.torque_90 >= %f AND s.torque_0 >= %f 
		  AND (s.torque_0 + sra.torque_0)  * %f - s.torque_0 >= %f 
		  AND (s.torque_90 + sra.torque_90) * %f - s.torque_90 >= %f
		ORDER BY id`,
		chPressure, torqueBTC, torqueETC, curPressure/chPressure, torqueBTO, curPressure/chPressure, torqueETO,
	)

	rows, err = db.Query(query)
	if err != nil {
		return false, []ActResult{}
	}
	var actList []ActResult

	for rows.Next() {
		var (
			airTorque0, airTorque90, sprTorque90, sprTorque0 float64
			sprNum                                           int
			model                                            string
			optimal                                          bool
		)
		err = rows.Scan(&airTorque0, &airTorque90, &sprNum, &optimal, &sprTorque90, &sprTorque0, &model)
		fatalErr(err)
		actList = append(actList, ActResult{
			model:     model,
			springNum: sprNum,
			torqueBTO: (airTorque0+sprTorque0)*curPressure/chPressure - sprTorque0,
			torqueETO: (airTorque90+sprTorque90)*curPressure/chPressure - sprTorque90,
			torqueBTC: sprTorque90,
			torqueETC: sprTorque0,
			optimal:   optimal,
		})
	}
	return true, actList
}

//func SRSelector(pressure, nomTorque float64, safetyFactorSR safetyFactorSR) (bool, ActResult) {
//	torqueBTO := safetyFactorSR.factorBTO * nomTorque
//	torqueETO := safetyFactorSR.factorETO * nomTorque
//	torqueBTC := safetyFactorSR.factorBTC * nomTorque
//	torqueETC := safetyFactorSR.factorETC * nomTorque
//	for _, act := range actuatorsList {
//		for springNum, spr := range act.springs {
//			actTorque0 := act.torque*pressure/initPressure - spr.torque0
//			actTorque90 := act.torque*pressure/initPressure - spr.torque90
//			if minSpringPressure[springNum] <= pressure &&
//				maxSpringPressure[springNum] >= pressure &&
//				(spr.torque90 > torqueBTC || safetyFactorSR.factorBTC-spr.torque90/nomTorque <= 0.01) &&
//				(spr.torque0 > torqueETC || safetyFactorSR.factorETC-spr.torque0/nomTorque <= 0.01) &&
//				(actTorque0 > torqueBTO || safetyFactorSR.factorBTO-actTorque0/nomTorque <= 0.01) &&
//				(actTorque90 > torqueETO || safetyFactorSR.factorETO-actTorque90/nomTorque <= 0.01) {
//				return true, ActResult{
//					model:     act.model,
//					springNum: springNum + 5,
//					torqueBTO: actTorque0,
//					torqueETO: actTorque90,
//					torqueBTC: spr.torque90,
//					torqueETC: spr.torque0,
//				}
//			}
//		}
//	}
//	return false, ActResult{}
//}

func printResultDA(result ActResult, nomTorque float64) bool {
	broken := false
	daSafetyFactor := result.torque / nomTorque
	fmt.Printf(
		"\nМодель привода - %sDA\nКрутящий момент привода - %.2f Н*м\nКоэффициент запаса - %.2f\n",
		result.model,
		result.torque,
		daSafetyFactor,
	)
	if daSafetyFactor >= maxSafetyFactor {
		broken = true
	}
	return broken
}

func printResultSR(result ActResult, nomTorque float64, sfStruct safetyFactorSR) bool {
	broken := false
	safetyFactorBTO := result.torqueBTO / nomTorque
	safetyFactorETO := result.torqueETO / nomTorque
	safetyFactorBTC := result.torqueBTC / nomTorque
	safetyFactorETC := result.torqueETC / nomTorque
	fmt.Printf(
		"\nМодель привода - %sSR\nПружины номер - %d\nКрутящие моменты привода:\n",
		result.model, result.springNum,
	)
	fmt.Printf(
		" BTO - %.2f Н*м\n ETO - %.2f Н*м\n BTC - %.2f Н*м\n ETC - %.2f Н*м\n",
		result.torqueBTO, result.torqueETO, result.torqueBTC, result.torqueETC,
	)
	fmt.Println("Коэффициенты запаса:")
	fmt.Printf(
		" BTO - %.2f (задан - %.2f)\n ETO - %.2f (задан - %.2f)\n BTC - %.2f (задан - %.2f)\n ETC - %.2f (задан - %.2f)\n",
		safetyFactorBTO,
		sfStruct.factorBTO,
		safetyFactorETO,
		sfStruct.factorETO,
		safetyFactorBTC,
		sfStruct.factorBTC,
		safetyFactorETC,
		sfStruct.factorETC,
	)
	fmt.Print("Оптимальный вариант (воздух/пружины): ")
	if result.optimal {
		fmt.Println("да")
	} else {
		fmt.Println("нет")
	}

	if safetyFactorBTO >= maxSafetyFactor ||
		safetyFactorETO >= maxSafetyFactor ||
		safetyFactorBTC >= maxSafetyFactor ||
		safetyFactorETC >= maxSafetyFactor {
		broken = true
	}
	return broken
}

func checkOptimalResult(actList []ActResult, nomTorque float64, sfStruct safetyFactorSR) bool {
	if !actList[0].optimal {
		count := stepToOptimal
		for i := 1; i <= count && i < len(actList); i++ {
			if actList[i].optimal {
				fmt.Println("\nВНИМАНИЕ! Возможно лучше взять оптимальный вариант привода (смотри ниже).")
				return printResultSR(actList[i], nomTorque, sfStruct)
			}
		}
	}
	return false
}

func mainMenu() {
	fmt.Println("\nВыберите режим работы программы:")
	fmt.Println(" 1 - подбор двухстороннего привода\n 2 - подбор НЗ привода с пружинами для затвора\n" +
		" 3 - подбор НЗ привода с пружинами для крана\n 4 - подбор НЗ привода с ручным вводом всех коэффициентов запаса (BTO, ETO, BTC, ETC)\n" +
		" 5 - завершить работу программы")
	menu := inputMode()
	if menu == 5 {
		return
	}
	ok = false
	nomTorque := inputParam("Введите номинальный крутящий момент (Н*м): ")
	pressure := inputParam("Введите рабочее давление (бар): ")
	if menu == 4 {
		safetyFactor = 1.25
	} else {
		safetyFactor = inputParam("Введите коэффициент запаса (например 1.25): ")
	}
	okParams := checkParams(pressure, nomTorque, safetyFactor)
	if okParams {
		switch {
		case menu == 1:
			ok, result = DASelector(pressure, nomTorque, safetyFactor)
		case menu == 2:
			sfStruct = safetyFactorSR{safetyFactor, 0.5, 1, safetyFactor}
			ok, resultList = SRSelector(pressure, nomTorque, sfStruct)
		case menu == 3:
			sfStruct = safetyFactorSR{safetyFactor, 0.8, safetyFactor, 0.8}
			ok, resultList = SRSelector(pressure, nomTorque, sfStruct)
		case menu == 4:
			factorBTO := inputParam("Введите коэффициент запаса BTO (например 1.25): ")
			factorETO := inputParam("Введите коэффициент запаса ETO (например 0.5): ")
			factorBTC := inputParam("Введите коэффициент запаса BTC (например 1.0): ")
			factorETC := inputParam("Введите коэффициент запаса ETC (например 1.25): ")
			sfStruct = safetyFactorSR{factorBTO, factorETO, factorBTC, factorETC}
			ok, resultList = SRSelector(pressure, nomTorque, sfStruct)
		}

		if ok {
			switch {
			case menu == 1:
				broken = printResultDA(result, nomTorque)
			default:
				broken = printResultSR(resultList[0], nomTorque, sfStruct)
				check := checkOptimalResult(resultList, nomTorque, sfStruct)
				if check {
					broken = true
				}
			}
			if broken {
				fmt.Println("ВНИМАНИЕ! Очень большой коэффициент запаса - возможно механическое разрушение штока арматуры.")
			}
		}
	}

	if !ok {
		fmt.Println("\nДля заданных парметров не удалось ничего подобрать. Попробуйте другие параметры.")
	}

	mainMenu()
}

func main() {
	fmt.Println("Подбор пневмоприводов ArTorq")
	fmt.Println("\nВНИМАНИЕ! Нецелые числа необходимо писать ЧЕРЕЗ ТОЧКУ.\nНапример: 1.25 - верно; 1,25 - неверно.")
	mainMenu()
}
