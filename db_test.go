package main

import (
	"database/sql"
	"fmt"
	"testing"

	"math"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/exp/slices"
)

// Check DA actuators data in DB
func TestDaAirTorque(t *testing.T) {
	db, err := sql.Open("sqlite3", "db.sqlite3")
	fatalErr(err)
	defer db.Close()
	query := "SELECT DISTINCT actuator_model FROM da_air_torque"
	rows, err := db.Query(query)
	fatalErr(err)
	defer rows.Close()

	var modelsList []string
	for rows.Next() {
		var model string
		err = rows.Scan(&model)
		fatalErr(err)
		modelsList = append(modelsList, model)
	}

	for _, model := range modelsList {
		modelMap := make(map[float64]float64)
		query = fmt.Sprintf("SELECT pressure, torque FROM da_air_torque WHERE actuator_model = '%s'", model)
		rows, err := db.Query(query)
		fatalErr(err)
		for rows.Next() {
			var pressure, torque float64
			err = rows.Scan(&pressure, &torque)
			fatalErr(err)
			modelMap[pressure] = torque
		}
		curPressure := 2.5
		curTorque := modelMap[curPressure]
		for pressure, torque := range modelMap {
			if math.Abs(pressure/curPressure-torque/curTorque) > 0.1 {
				t.Errorf(
					"Model - %s; curPressure=%f; curTorque=%f; pressure=%f; torque=%f;",
					model, curPressure, curTorque, pressure, torque,
				)
			}
		}
	}
}

// Check SR actuators data in DB
func TestSrAirTorque(t *testing.T) {
	exceptionsIdList := []int{147, 246, 295, 297}
	db, err := sql.Open("sqlite3", "db.sqlite3")
	fatalErr(err)
	defer db.Close()
	query := `SELECT id, sra.torque_0, sra.torque_90, number, pressure, s.torque_90, s.torque_0, s.actuator_model 
			FROM sr_air_torque sra JOIN springs s on s.text_id = sra.spring_id 
			ORDER BY id`
	rows, err := db.Query(query)
	fatalErr(err)
	defer rows.Close()
	for rows.Next() {
		var (
			airTorque0, airTorque90, springTorque0, springTorque90, pressure float64
			model                                                            string
			number, id                                                       int
		)
		err = rows.Scan(&id, &airTorque0, &airTorque90, &number, &pressure, &springTorque90, &springTorque0, &model)
		fatalErr(err)
		delta := (airTorque0 + springTorque0) - (airTorque90 + springTorque90)
		if math.Abs(delta) > 4 && !slices.Contains(exceptionsIdList, id) {
			t.Errorf(
				"ID - %d; Model - %s; Spring number - %d; Pressure = %.1f; (airTorque0 + springTorque0) - (airTorque90 + springTorque90) = (%f + %f) - (%f + %f) = %.2f",
				id, model, number, pressure, airTorque0, springTorque0, airTorque90, springTorque90, delta,
			)
		}
	}
}

// Check optimal SR actuators
func TestOptimal(t *testing.T) {
	db, err := sql.Open("sqlite3", "db.sqlite3")
	fatalErr(err)
	defer db.Close()

	pressureNumberMap := map[float64]int{
		2.5: 5,
		3:   6,
		3.5: 7,
		4:   8,
		4.5: 9,
		5:   10,
		5.5: 11,
		6:   12,
		7:   12,
		8:   12,
	}
	query := `SELECT id, number, pressure, optimal, s.actuator_model 
			FROM sr_air_torque sra JOIN springs s on s.text_id = sra.spring_id 
			ORDER BY number`
	rows, err := db.Query(query)
	fatalErr(err)
	defer rows.Close()

	for rows.Next() {
		var (
			pressure            float64
			model               string
			number, id, optimal int
		)

		err = rows.Scan(&id, &number, &pressure, &optimal, &model)
		fatalErr(err)
		optimalNumber := pressureNumberMap[pressure]
		if (number == optimalNumber && optimal != 1) ||
			(number != optimalNumber && optimal != 0) {
			t.Errorf(
				"ID - %d; Model - %s; Spring number - %d; Pressure = %.1f; Optimal - %d",
				id, model, number, pressure, optimal,
			)
		}
	}
}
