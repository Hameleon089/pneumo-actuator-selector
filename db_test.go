package main

import (
	"database/sql"
	"fmt"
	"testing"

	"math"

	_ "github.com/mattn/go-sqlite3"
)

// Check DA actuators data in db
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
