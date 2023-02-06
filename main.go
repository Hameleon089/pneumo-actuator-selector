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
		if safetyFactor < actuatorTorque/nomTorque || safetyFactor-actuatorTorque/nomTorque <= 0.01 {
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
				(spr.torque90 > torqueBTC || safetyFactorSR.factorBTC-spr.torque90/nomTorque <= 0.01) &&
				(spr.torque0 > torqueETC || safetyFactorSR.factorETC-spr.torque0/nomTorque <= 0.01) &&
				(actTorque0 > torqueBTO || safetyFactorSR.factorBTO-actTorque0/nomTorque <= 0.01) &&
				(actTorque90 > torqueETO || safetyFactorSR.factorETO-actTorque90/nomTorque <= 0.01) {
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

	if safetyFactorBTO >= maxSafetyFactor ||
		safetyFactorETO >= maxSafetyFactor ||
		safetyFactorBTC >= maxSafetyFactor ||
		safetyFactorETC >= maxSafetyFactor {
		broken = true
	}
	return broken
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
				broken = printResultDA(result, nomTorque)
			default:
				broken = printResultSR(result, nomTorque, sfStruct)
			}
			if broken {
				fmt.Println("ВНИМАНИЕ! Очень большой коэффициент запаса - возможно механическое разрушение штока арматуры.")
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
