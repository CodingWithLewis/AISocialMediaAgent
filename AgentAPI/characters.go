package main

import (
	_ "github.com/lib/pq"
)

//
//func main() {
//
//	err := godotenv.Load(".env")
//
//	if err != nil {
//		log.Fatal("Error loading.env file")
//	}
//
//	connStr := "postgresql://elebumm:" + os.Getenv("NEON_API_KEY") + "@ep-cool-moon-a5ygl65p-pooler.us-east-2.aws.neon.tech/aitwitter?sslmode=require"
//	db, err := sql.Open("postgres", connStr)
//	if err != nil {
//		panic(err)
//	}
//	defer db.Close()
//
//	// Create table if not exists
//	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id serial PRIMARY KEY, username VARCHAR(255), password VARCHAR(255), " +
//		"email VARCHAR(255), bio VARCHAR(255), image VARCHAR(255), created_at TIMESTAMP," +
//		"updated_at TIMESTAMP)")
//
//}
