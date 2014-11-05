package database

import "log"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "crypto/sha1"
import "hash"
import "time"
import "encoding/base64"
import "fmt"
import "math"

var sh hash.Hash = sha1.New()

func InitializeDatabase() error {
	db, err := sql.Open("mysql", "root:@/CAB_REPUBLIC")
	if err != nil {
		log.Fatal("Cannot connect to the database server.")
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Connection failed.")
	}

	defer db.Close()

	_, err = db.Query("CREATE DATABASE IF NOT EXISTS CAB_REPUBLIC")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Query("CREATE TABLE IF NOT EXISTS user (id INT(11) NOT NULL AUTO_INCREMENT, email VARCHAR(31) UNIQUE, password VARCHAR(63), age_min TINYINT(8), age_max TINYINT(8), gender_preference TINYINT(2), access_token VARCHAR(63), PRIMARY KEY(id))")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Query("CREATE TABLE IF NOT EXISTS intention (id INT(11) NOT NULL AUTO_INCREMENT, user_id INT(11), destination_longitude DOUBLE, destination_latitude DOUBLE, time_initiate_intention timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, is_deciding TINYINT(1), PRIMARY KEY(id), FOREIGN KEY (user_id) REFERENCES user(id))")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Query("CREATE TABLE IF NOT EXISTS pickup_location (id INT(11) NOT NULL AUTO_INCREMENT, name VARCHAR(31), longitude DOUBLE, latitude DOUBLE, PRIMARY KEY(id))")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Query("CREATE TABLE IF NOT EXISTS amatch (id INT(11) NOT NULL AUTO_INCREMENT, intention1 INT(11), intention2 INT(11), PRIMARY KEY(id), FOREIGN KEY (intention1) REFERENCES intention(id), FOREIGN KEY (intention2) REFERENCES intention(id))")
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func IsInitialized(db *sql.DB) bool {
	return false
}

func LogIn(email string, password string) (bool, int, int, int, string) {
	var ageMin int
	var ageMax int
	var genderPreference int
	db, err := sql.Open("mysql", "root:@/CAB_REPUBLIC")
	if err != nil {
		log.Fatal("Cannot connect to the database server.")
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Connection failed.")
	}

	defer db.Close()
	if err != nil {
		log.Fatal(err)
		return false, 0, 0, 0, ""
	}
	defer db.Close()

	err = db.QueryRow("SELECT age_min, age_max, gender_preference FROM user WHERE email=? and password=?", email, password).Scan(&ageMin, &ageMax, &genderPreference)
	switch {
	case err == sql.ErrNoRows:
		return false, -1, -1, -1, ""
	case err != nil:
		fmt.Println(err)
		return false, -2, -2, -2, ""
	default:
		sh.Write([]byte(email + password + time.Now().Format(time.ANSIC)))
		accessToken := base64.URLEncoding.EncodeToString(sh.Sum(nil))
		_, err := db.Exec("UPDATE user SET access_token=? WHERE email=?", accessToken, email)
		if err == nil {
			return true, ageMin, ageMax, genderPreference, accessToken
		} else {
			return false, -3, -3, -3, ""
		}
	}

}

func UpdateUser(ageMin int, ageMax int, gender int, token string) bool {
	db, err := sql.Open("mysql", "root:@/CAB_REPUBLIC")
	if err != nil {
		log.Fatal("Cannot connect to the database server.")
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Connection failed.")
	}

	defer db.Close()
	var id int
	err = db.QueryRow("SELECT id FROM user WHERE access_token=?", token).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		fmt.Println("No such user")
		return false
	case err != nil:
		fmt.Println(err)
		return false
	default:
		_, err := db.Exec("UPDATE user SET age_min=? , age_max=? , gender_preference=? WHERE id=?", ageMin, ageMax, gender, id)
		if err == nil {
			return true
		} else {
			fmt.Println(err)
			return false
		}
	}
}

func CreateIntention(latitude float64, longitude float64, token string) bool {
	db, err := sql.Open("mysql", "root:@/CAB_REPUBLIC")
	if err != nil {
		log.Fatal("Cannot connect to the database server.")
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Connection failed.")
	}

	defer db.Close()
	var id int
	err = db.QueryRow("SELECT id FROM user WHERE access_token=?", token).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		fmt.Println("No such user")
		return false
	case err != nil:
		fmt.Println(err)
		return false
	default:
		_, err := db.Exec("INSERT INTO intention (user_id, destination_latitude, destination_longitude, is_deciding) VALUES (?, ?, ?, 0)", id, latitude, longitude)
		if err == nil {
			return true
		} else {
			fmt.Println(err)
			return false
		}
	}
}

func FindMatch(emails []string, token string) (bool, string, string) {
	db, err := sql.Open("mysql", "root:@/CAB_REPUBLIC")
	if err != nil {
		log.Fatal("Cannot connect to the database server.")
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Connection failed.")
	}

	defer db.Close()
	var id int
	var longitude float64
	var latitude float64
	err = db.QueryRow("SELECT id FROM user WHERE access_token=?", token).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		fmt.Println("No such user")
		return false, "", ""
	case err != nil:
		fmt.Println(err)
		return false, "", ""
	default:
		err = db.QueryRow("SELECT destination_latitude, destination_longitude FROM intention WHERE user_id=?", id).Scan(&latitude, &longitude)
		if err == nil {
			var latitude1 float64
			var longitude1 float64
			var id1 int
			for _, email := range emails {
				err = db.QueryRow("SELECT id FROM user WHERE email=?", email).Scan(&id1)
				if err == nil {
					err = db.QueryRow("SELECT destination_latitude, destination_longitude FROM intention WHERE user_id=?", id1).Scan(&latitude1, &longitude1)
					if err == nil {
						rows, err := db.Query("SELECT name, longitude, latitude FROM pickup_location")
						if err != nil {
							log.Fatal(err)
						}
						min := 9999.999
						var minName string
						var distance float64
						var name string
						var lo float64
						var la float64
						defer rows.Close()
						for rows.Next() {
							if err = rows.Scan(&name, &lo, &la); err != nil {
								log.Fatal(err)
							}
							distance = math.Sqrt(math.Pow(lo-longitude, 2)+math.Pow(la-latitude, 2)) + math.Sqrt(math.Pow(lo-longitude1, 2)+math.Pow(la-latitude1, 2))
							if distance < min {
								min = distance
								minName = name
							}
						}
						if err := rows.Err(); err != nil {
							log.Fatal(err)
						}
						return true, email, minName
					}
				}
			}
		}

		return false, "", ""
	}
}

func openDB() error {
	db, err := sql.Open("mysql", "root:@/CAB_REPUBLIC")
	if err != nil {
		log.Fatal("Cannot connect to the database server.")
		return err
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Connection failed.")
		return err
	}
	return nil
}
