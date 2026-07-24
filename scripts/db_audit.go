package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "data/universalops.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tables := []string{"metrics", "logs", "events", "alerts", "conversations", "settings", "forensics", "reports", "baselines", "health_scores", "custom_workflows", "incidents", "schema_versions"}
	fmt.Println("=== ROW COUNTS ===")
	for _, t := range tables {
		var n int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + t).Scan(&n); err != nil {
			fmt.Printf("  %-20s ERROR: %v\n", t, err)
			continue
		}
		fmt.Printf("  %-20s %d\n", t, n)
	}

	fmt.Println()
	var pageCount, freeList int
	db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	db.QueryRow("PRAGMA freelist_count").Scan(&freeList)
	fmt.Printf("  fragmentation: %.1f%%\n", float64(freeList)/float64(pageCount)*100)
	var jMode, syncMode string
	db.QueryRow("PRAGMA journal_mode").Scan(&jMode)
	db.QueryRow("PRAGMA synchronous").Scan(&syncMode)
	fmt.Printf("  journal_mode: %s, synchronous: %s\n", jMode, syncMode)

	fmt.Println()
	var hCount int
	db.QueryRow("SELECT COUNT(*) FROM health_scores").Scan(&hCount)
	fmt.Printf("health_scores rows: %d\n", hCount)

	var errCount int
	db.QueryRow("SELECT COUNT(*) FROM logs WHERE level='ERROR'").Scan(&errCount)
	fmt.Printf("ERROR logs: %d\n", errCount)

	fmt.Println()
	fmt.Println("=== METRIC COUNTS PER NAME (top 10) ===")
	crows, err := db.Query("SELECT name, COUNT(*) as cnt FROM metrics GROUP BY name ORDER BY cnt DESC LIMIT 10")
	if err != nil {
		log.Fatal(err)
	}
	defer crows.Close()
	for crows.Next() {
		var name string
		var cnt int
		crows.Scan(&name, &cnt)
		fmt.Printf("  %-25s %d\n", name, cnt)
	}

	fmt.Println()
	fmt.Println("=== RECENT ALERTS (last 10) ===")
	arows, err := db.Query("SELECT id, level, metric, resolved FROM alerts ORDER BY id DESC LIMIT 10")
	if err != nil {
		log.Fatal(err)
	}
	defer arows.Close()
	for arows.Next() {
		var id int
		var level, metric string
		var resolved bool
		arows.Scan(&id, &level, &metric, &resolved)
		fmt.Printf("  [%d] %s %s resolved=%v\n", id, level, metric, resolved)
	}

	fmt.Println()
	fmt.Println("=== TABLE SCHEMAS ===")
	srows, err := db.Query("SELECT name, sql FROM sqlite_master WHERE type='table' AND name IN ('alerts','baselines','health_scores') ORDER BY name")
	if err != nil {
		log.Fatal(err)
	}
	defer srows.Close()
	for srows.Next() {
		var name, sql string
		srows.Scan(&name, &sql)
		fmt.Printf("-- %s --\n%s\n\n", name, sql)
	}

	fmt.Println("=== ALERTS (last 10, all columns) ===")
	arows2, err := db.Query("SELECT * FROM alerts ORDER BY rowid DESC LIMIT 10")
	if err != nil {
		log.Fatal(err)
	}
	defer arows2.Close()
	cols, _ := arows2.Columns()
	fmt.Printf("  columns: %v\n", cols)
	for arows2.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		arows2.Scan(ptrs...)
		fmt.Printf("  %v\n", vals)
	}

	fmt.Println()
	fmt.Println("=== BASELINES ===")
	brows, err := db.Query("SELECT * FROM baselines ORDER BY rowid DESC LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer brows.Close()
	bcols, _ := brows.Columns()
	fmt.Printf("  columns: %v\n", bcols)
	for brows.Next() {
		vals := make([]interface{}, len(bcols))
		ptrs := make([]interface{}, len(bcols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		brows.Scan(ptrs...)
		fmt.Printf("  %v\n", vals)
	}

	fmt.Println()
	fmt.Println("=== HEALTH SCORES ===")
	hrows, err := db.Query("SELECT * FROM health_scores ORDER BY rowid DESC LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer hrows.Close()
	hcols, _ := hrows.Columns()
	fmt.Printf("  columns: %v\n", hcols)
	for hrows.Next() {
		vals := make([]interface{}, len(hcols))
		ptrs := make([]interface{}, len(hcols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		hrows.Scan(ptrs...)
		fmt.Printf("  %v\n", vals)
	}
}
