package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "../test.db") // or ../featureplus.db if that's your main DB
	if err != nil {
		panic(err)
	}
	defer db.Close()

	baseDir := "../../frontend/out/data/projects"
	os.MkdirAll(baseDir, 0755)

	// Get all project IDs
	projectRows, err := db.Query("SELECT id FROM projects")
	if err != nil {
		panic(err)
	}
	defer projectRows.Close()

	var projectIDs []int
	for projectRows.Next() {
		var pid int
		if err := projectRows.Scan(&pid); err == nil {
			projectIDs = append(projectIDs, pid)
		}
	}

	for _, projectID := range projectIDs {
		// 1. Export features list for this project
		featuresDir := filepath.Join(baseDir, fmt.Sprintf("%d/features", projectID))
		os.MkdirAll(featuresDir, 0755)
		featuresListPath := filepath.Join(baseDir, fmt.Sprintf("%d/features.json", projectID))

		featuresRows, err := db.Query("SELECT * FROM features WHERE project_id = ?", projectID)
		if err != nil {
			fmt.Println("Error querying features for project", projectID, ":", err)
			continue
		}
		var featuresList []map[string]interface{}
		cols, _ := featuresRows.Columns()
		for featuresRows.Next() {
			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}
			if err := featuresRows.Scan(columnPointers...); err != nil {
				fmt.Println("Error scanning feature row:", err)
				continue
			}
			feature := make(map[string]interface{})
			var featureID interface{}
			for i, colName := range cols {
				val := columnPointers[i].(*interface{})
				feature[colName] = *val
				if colName == "id" {
					featureID = *val
				}
			}
			featuresList = append(featuresList, feature)
			// 2. Export feature details for this project/feature
			if featureID != nil {
				featurePath := filepath.Join(featuresDir, fmt.Sprintf("%v.json", featureID))
				f, err := os.Create(featurePath)
				if err == nil {
					json.NewEncoder(f).Encode(feature)
					f.Close()
				}
			}
		}
		featuresRows.Close()
		// Write features list for this project
		f, err := os.Create(featuresListPath)
		if err == nil {
			json.NewEncoder(f).Encode(featuresList)
			f.Close()
		}
		fmt.Printf("Exported features for project %d\n", projectID)
	}

	// Export all projects to index.json
	projectsRows, err := db.Query("SELECT * FROM projects")
	if err != nil {
		panic(err)
	}
	defer projectsRows.Close()

	var projectsList []map[string]interface{}
	projectCols, _ := projectsRows.Columns()
	for projectsRows.Next() {
		columns := make([]interface{}, len(projectCols))
		columnPointers := make([]interface{}, len(projectCols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := projectsRows.Scan(columnPointers...); err != nil {
			continue
		}
		project := make(map[string]interface{})
		for i, colName := range projectCols {
			val := columnPointers[i].(*interface{})
			project[colName] = *val
		}
		projectsList = append(projectsList, project)
	}
	projectsIndexPath := filepath.Join(baseDir, "index.json")
	f, err := os.Create(projectsIndexPath)
	if err == nil {
		json.NewEncoder(f).Encode(projectsList)
		f.Close()
	}
	fmt.Println("Exported all projects to index.json.")

	fmt.Println("Exported all project features to JSON.")
}
