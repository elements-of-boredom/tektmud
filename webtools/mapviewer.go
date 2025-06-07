package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"tektmud/internal/rooms"
)

// go run webtools/mapviewer.go
func setupMapRoutes() {
	http.HandleFunc("/map", serveMapPage)
	http.HandleFunc("/3dmap", serve3dMapPage)
	http.HandleFunc("/api/areas", serveAreasData)
}

func serveMapPage(w http.ResponseWriter, r *http.Request) {
	// Serve the HTML map generator page
	http.ServeFile(w, r, "webtools/web/map.html")
}
func serve3dMapPage(w http.ResponseWriter, r *http.Request) {
	// Serve the HTML map generator page
	http.ServeFile(w, r, "webtools/web/map3d.html")
}

func serveAreasData(w http.ResponseWriter, r *http.Request) {
	// Load areas data and return as JSON
	areas, err := rooms.LoadAreas("_data/world")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load areas: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for development

	if err := json.NewEncoder(w).Encode(areas); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode areas: %v", err), http.StatusInternalServerError)
		return
	}
}

// Enhanced version that can serve specific area data, needs a diff endpoint
func serveAreaData(w http.ResponseWriter, r *http.Request) {
	areaId := r.URL.Query().Get("id")
	if areaId == "" {
		http.Error(w, "Area ID required", http.StatusBadRequest)
		return
	}

	areas, err := rooms.LoadAreas("_data/world")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load areas: %v", err), http.StatusInternalServerError)
		return
	}

	area, exists := areas[areaId]
	if !exists {
		http.Error(w, "Area not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(area); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode area: %v", err), http.StatusInternalServerError)
		return
	}
}

// CLI tool to generate a JSON export of your areas
func exportAreasToJSON(worldPath, outputPath string) error {
	areas, err := rooms.LoadAreas(worldPath)
	if err != nil {
		return fmt.Errorf("failed to load areas: %w", err)
	}

	// Write to file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print

	if err := encoder.Encode(areas); err != nil {
		return fmt.Errorf("failed to encode areas to JSON: %w", err)
	}

	fmt.Printf("Areas exported to %s\n", outputPath)
	return nil
}

func main() {
	// Option 1: Export to JSON file for use with the web tool
	if len(os.Args) > 1 && os.Args[1] == "export" {
		if err := exportAreasToJSON("_data/world", "webtools/web/areas_export.json"); err != nil {
			fmt.Printf("Export failed: %v\n", err)
			return
		}
		return
	}

	// Option 2: Serve the map via HTTP
	setupMapRoutes()

	fmt.Println("Map server starting on http://localhost:8080/map")
	fmt.Println("3d Map server starting on http://localhost:8080/3dmap")
	fmt.Println("Areas API available at http://localhost:8080/api/areas")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
