package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

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

func DASelector(pressure, nomTorque, safetyFactor float64) (bool, ActResult) {
	for _, act := range actuatorsList {
		actuatorTorque := act.torque * pressure / initPressure
		if nomTorque*safetyFactor <= actuatorTorque {
			return true, ActResult{
				model:  act.model,
				torque: actuatorTorque,
			}
		}
	}
	return false, ActResult{}
}

func SRSelector(pressure, nomTorque float64, safetyFactorSR safetyFactorSR) (bool, ActResult) {
	torqueBTO := safetyFactorSR.factorBTO * nomTorque
	torqueETO := safetyFactorSR.factorETO * nomTorque
	torqueBTC := safetyFactorSR.factorBTC * nomTorque
	torqueETC := safetyFactorSR.factorETC * nomTorque
	for _, act := range actuatorsList {
		for springNum, spr := range act.springs {
			actTorque0 := act.torque*pressure/initPressure - spr.torque0
			actTorque90 := act.torque*pressure/initPressure - spr.torque90
			if minSpringPressure[springNum] <= pressure &&
				maxSpringPressure[springNum] >= pressure &&
				spr.torque90 >= torqueBTC &&
				spr.torque0 >= torqueETC &&
				actTorque0 >= torqueBTO &&
				actTorque90 >= torqueETO {
				return true, ActResult{
					model:     act.model,
					springNum: springNum + 5,
					torqueBTO: actTorque0,
					torqueETO: actTorque90,
					torqueBTC: spr.torque90,
					torqueETC: spr.torque0,
				}
			}
		}
	}
	return false, ActResult{}
}

func printResultDA(result ActResult, nomTorque float64) {
	fmt.Printf(
		"\nМодель привода - %sDA\nКоэффициент запаса - %.2f\n",
		result.model,
		result.torque/nomTorque,
	)
}

func printResultSR(result ActResult, nomTorque float64, sfStruct safetyFactorSR) {
	fmt.Printf(
		"\nМодель привода - %sSR\nПружины номер - %d\nКоэффициенты запаса:\n",
		result.model, result.springNum,
	)
	fmt.Printf(
		" BTO - %.2f (задан - %.2f)\n ETO - %.2f (задан - %.2f)\n BTC - %.2f (задан - %.2f)\n ETC - %.2f (задан - %.2f)\n",
		result.torqueBTO/nomTorque,
		sfStruct.factorBTO,
		result.torqueETO/nomTorque,
		sfStruct.factorETO,
		result.torqueBTC/nomTorque,
		sfStruct.factorBTC,
		result.torqueETC/nomTorque,
		sfStruct.factorETC,
	)

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
			ok, result = SRSelector(pressure, nomTorque, sfStruct)
		case menu == 3:
			sfStruct = safetyFactorSR{safetyFactor, 0.8, safetyFactor, 0.8}
			ok, result = SRSelector(pressure, nomTorque, sfStruct)
		case menu == 4:
			factorBTO := inputParam("Введите коэффициент запаса BTO (например 1.25): ")
			factorETO := inputParam("Введите коэффициент запаса ETO (например 0.5): ")
			factorBTC := inputParam("Введите коэффициент запаса BTC (например 1.0): ")
			factorETC := inputParam("Введите коэффициент запаса ETC (например 1.25): ")
			sfStruct = safetyFactorSR{factorBTO, factorETO, factorBTC, factorETC}
			ok, result = SRSelector(pressure, nomTorque, sfStruct)
		}

		if ok {
			switch {
			case menu == 1:
				printResultDA(result, nomTorque)
			default:
				printResultSR(result, nomTorque, sfStruct)
			}
		}
	}

	if !ok {
		fmt.Println("Для заданных парметров не удалось ничего подобрать, попробуйте другие параметры.")
	}

	mainMenu()
}

func main() {
	fmt.Println("Подбор пневмоприводов ArTorq")
	fmt.Println("\nВНИМАНИЕ! Нецелые числа необходимо писать ЧЕРЕЗ ТОЧКУ.\nНапример: 1.25 - верно; 1,25 - неверно.")
	mainMenu()
}
