package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/MyFirstGo/internal/env"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Koneksi Database
	db, err := sql.Open("postgres", env.GetString("DB_MIGRATOR_ADDR", "postgres://admin:adminpassword@localhost/nutritrack?sslmode=disable"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Definisi Master Nutrients & Units
	nutrientUnits := map[string]string{
		"Caloric Value": "kcal", "Fat": "g", "Saturated Fats": "g",
		"Monounsaturated Fats": "g", "Polyunsaturated Fats": "g",
		"Carbohydrates": "g", "Sugars": "g", "Protein": "g",
		"Dietary Fiber": "g", "Cholesterol": "mg", "Sodium": "g",
		"Water": "g", "Vitamin A": "mg", "Vitamin B1": "mg",
		"Vitamin B11": "mg", "Vitamin B12": "mg", "Vitamin B2": "mg",
		"Vitamin B3": "mg", "Vitamin B5": "mg", "Vitamin B6": "mg",
		"Vitamin C": "mg", "Vitamin D": "mg", "Vitamin E": "mg",
		"Vitamin K": "mg", "Calcium": "mg", "Copper": "mg",
		"Iron": "mg", "Magnesium": "mg", "Manganese": "mg",
		"Phosphorus": "mg", "Potassium": "mg", "Selenium": "mg",
		"Zinc": "mg", "Nutrition Density": "index",
	}

	ctx := context.Background()

	// 3. SEED MASTER NUTRIENTS
	// Simpan ID nutrient ke map untuk referensi cepat
	nutrientIDMap := make(map[string]int64)
	for name, unit := range nutrientUnits {
		var id int64
		err := db.QueryRowContext(ctx,
			"INSERT INTO nutrients (name, unit) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET unit = EXCLUDED.unit RETURNING id",
			name, unit).Scan(&id)
		if err != nil {
			log.Printf("Gagal insert nutrient %s: %v", name, err)
			continue
		}
		nutrientIDMap[name] = id
	}
	fmt.Println("✅ Master nutrients seeded.")

	// 4. SEED FOODS & PIVOT
	for i := range 5 {
		file, err := os.Open(fmt.Sprintf("FOOD-DATA-GROUP%d.csv", i+1)) // Asumsi file CSV namannya data.csv
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		header, _ := reader.Read() // Ambil header: [,Unnamed, food, Caloric Value, dst]

		// Iterasi tiap baris data
		for {
			record, err := reader.Read()
			if err != nil {
				break
			}

			// Mulai Transaksi per baris food (Atomic)
			tx, _ := db.BeginTx(ctx, nil)

			// Insert Food (Asumsi data per 100g sesuai CSV-mu)
			foodName := record[2]
			var foodID int64
			err = tx.QueryRowContext(ctx,
				"INSERT INTO foods (name, serving_size, serving_unit) VALUES ($1, 100, 'g') RETURNING id",
				foodName).Scan(&foodID)

			if err != nil {
				tx.Rollback()
				log.Printf("Gagal insert food %s: %v", foodName, err)
				continue
			}

			// Insert Nutrients (Mulai dari kolom index 3 dst)
			for i := 3; i < len(header); i++ {
				nutrientName := header[i]
				nutrientID, exists := nutrientIDMap[nutrientName]
				if !exists {
					continue
				}

				val, _ := strconv.ParseFloat(record[i], 64)

				_, err = tx.ExecContext(ctx,
					"INSERT INTO food_nutrients (food_id, nutrient_id, amount) VALUES ($1, $2, $3)",
					foodID, nutrientID, val)

				if err != nil {
					log.Printf("Gagal insert pivot %s-%s: %v", foodName, nutrientName, err)
				}
			}

			tx.Commit()
			fmt.Printf("🚀 Seeded: %s\n", foodName)
		}
	}
}
