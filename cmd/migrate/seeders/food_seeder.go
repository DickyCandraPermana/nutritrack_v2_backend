package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// 1. Koneksi Database
	connStr := "user=admin password=adminpassword dbname=social sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Buka File CSV
	file, err := os.Open("FOOD-DATA-GROUP5.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Baca Header
	header, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	// 3. Mulai Transaksi agar insert cepat
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("INSERT INTO foods (name, nutrition) VALUES ($1, $2)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	count := 0
	for {
		// Baca baris satu per satu (Streaming)
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Gagal membaca baris: %v", err)
			continue
		}

		// Berdasarkan CSV kamu:
		// Kolom 0: (kosong/index)
		// Kolom 1: Unnamed: 0
		// Kolom 2: food (Nama Makanan)
		foodName := row[2]

		nutrients := make(map[string]interface{})

		// Loop mulai dari kolom ke-3 (Caloric Value) sampai terakhir
		for i := 3; i < len(header); i++ {
			valStr := strings.TrimSpace(row[i])
			if valStr == "" || valStr == "0" { // Skip jika kosong
				continue
			}

			// Coba konversi ke float64
			val, err := strconv.ParseFloat(valStr, 64)
			if err == nil {
				nutrients[header[i]] = val
			} else {
				// Jika bukan angka, simpan sebagai string saja
				nutrients[header[i]] = valStr
			}
		}

		// Konversi Map ke JSON
		nutrientsJSON, _ := json.Marshal(nutrients)

		// Eksekusi Insert
		_, err = stmt.Exec(foodName, nutrientsJSON)
		if err != nil {
			log.Printf("Gagal insert %s: %v", foodName, err)
			continue
		}

		count++
		if count%100 == 0 {
			fmt.Printf("Sedang memproses... %d data berhasil\n", count)
		}
	}

	// 4. Commit Transaksi
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Selesai! Total %d makanan diimpor ke database.\n", count)
}
