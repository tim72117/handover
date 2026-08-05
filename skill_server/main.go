package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var skillsDir string

func extractMetadata(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return nil, nil
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
		return nil, err
	}
	if metadata == nil {
		metadata = map[string]interface{}{}
	}

	metadata["id"] = filepath.Base(filepath.Dir(filePath))
	return metadata, nil
}

func getAllMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		http.Error(w, `{"detail":"Skills directory not found"}`, http.StatusNotFound)
		return
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		http.Error(w, `{"detail":"Skills directory not found"}`, http.StatusNotFound)
		return
	}

	results := []map[string]interface{}{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillFile := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			continue
		}
		metadata, err := extractMetadata(skillFile)
		if err != nil {
			log.Printf("解析 %s 錯誤: %v", skillFile, err)
			continue
		}
		if metadata != nil {
			results = append(results, metadata)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func getSkillContentHandler(w http.ResponseWriter, r *http.Request) {
	skillID := strings.TrimPrefix(r.URL.Path, "/skill/")

	if strings.Contains(skillID, "..") || strings.Contains(skillID, "/") || strings.Contains(skillID, "\\") {
		http.Error(w, `{"detail":"非法路徑名稱"}`, http.StatusBadRequest)
		return
	}

	skillFile := filepath.Join(skillsDir, skillID, "SKILL.md")

	absSkillFile, err := filepath.Abs(skillFile)
	if err != nil {
		http.Error(w, `{"detail":"非法路徑名稱"}`, http.StatusBadRequest)
		return
	}
	absSkillsDir, err := filepath.Abs(skillsDir)
	if err != nil {
		http.Error(w, `{"detail":"伺服器錯誤"}`, http.StatusInternalServerError)
		return
	}

	if !strings.HasPrefix(absSkillFile, absSkillsDir) {
		http.Error(w, `{"detail":"存取權限不足：不得讀取上層目錄"}`, http.StatusForbidden)
		return
	}

	content, err := os.ReadFile(skillFile)
	if err != nil {
		http.Error(w, `{"detail":"找不到技能 '`+skillID+`'"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":      skillID,
		"content": string(content),
	})
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	skillsDir = filepath.Join(filepath.Dir(exePath), "skills")

	// 若以 `go run` 執行，使用原始碼所在目錄
	if wd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(wd, "skills")); err == nil {
			skillsDir = filepath.Join(wd, "skills")
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", getAllMetadataHandler)
	mux.HandleFunc("/skill/", getSkillContentHandler)

	log.Println("Skill Management Server listening on 127.0.0.1:8000")
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", mux))
}
