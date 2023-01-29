package main

import (
	"bufio"
	"fmt"
	"log"
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

func fatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func DASelector(pressure, nomTorque, safetyFactor float64) (bool, int) {
	if pressure <= maxPressure && pressure >= initPressure {
		for num, actuator := range actuatorsList {
			actuatorTorque := actuator.torque * pressure / initPressure
			if nomTorque*safetyFactor <= actuatorTorque {
				return true, num
			}
		}
	}
	return false, 0
}

//func SRSelector(pressure, nomTorque, float64, safetyFactorSR safetyFactorSR) (bool, int, int) {
//
//}

func printResultDA(actuatorNum int, pressure, nomTorque float64) {
	curActuator := actuatorsList[actuatorNum]
	fmt.Printf(
		"Модель привода - %sDA\nКоэффициент запаса - %.2f\n",
		curActuator.model,
		curActuator.torque*pressure/initPressure/nomTorque,
	)
}

func mainMenu() {
	nomTorque := inputParam("Введите номинальный крутящий момент: ")
	pressure := inputParam("Введите рабочее давление: ")
	safetyFactor := inputParam("Введите коэффициент запаса: ")
	ok, actuatorNum := DASelector(pressure, nomTorque, safetyFactor)
	if ok {
		printResultDA(actuatorNum, pressure, nomTorque)
	} else {
		fmt.Println()
	}
	fmt.Print("Для завершения работы программы нажмите enter")
	_, _ = readFloatParam()
}

func main() {
	mainMenu()
}
