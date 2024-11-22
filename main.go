package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type QueryResult struct {
	FilePath   string
	Query      string
	Normalized string
}

type Config struct {
	FolderPath    string
	IgnoreFolders []string
	FileType      string
	NumWorkers    int
}

func parseFlags() Config {
	folderPath := flag.String("folder", ".", "Folder path to scan")
	ignoreFolders := flag.String("ignore", "vendor,node_modules", "Comma separated list of folders to ignore")
	fileType := flag.String("type", ".php", "File type to scan")
	numWorkers := flag.Int("workers", runtime.NumCPU(), "Number of worker goroutines")
	flag.Parse()

	return Config{
		FolderPath:    *folderPath,
		IgnoreFolders: strings.Split(*ignoreFolders, ","),
		FileType:      *fileType,
		NumWorkers:    *numWorkers,
	}
}

func findFiles(config Config) ([]string, error) {
	var files []string
	err := filepath.Walk(config.FolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for _, folder := range config.IgnoreFolders {
				if info.Name() == folder {
					return filepath.SkipDir
				}
			}
		}

		if !info.IsDir() && strings.HasSuffix(path, config.FileType) {
			files = append(files, path)
		}

		return nil
	})
	return files, err
}

func worker(jobs <-chan string, results chan<- []QueryResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for path := range jobs {
		if res, err := analyzeFile(path); err == nil {
			results <- res
		}
	}
}

func processFiles(files []string, config Config) []QueryResult {
	jobs := make(chan string, len(files))
	results := make(chan []QueryResult, len(files))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < config.NumWorkers; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	// Send jobs
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	// Wait for workers in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allQueries []QueryResult
	for result := range results {
		allQueries = append(allQueries, result...)
	}

	return allQueries
}

func analyzeFile(path string) ([]QueryResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	matches := findSQLQueries(string(data))
	results := make([]QueryResult, len(matches))
	for i, match := range matches {
		results[i] = QueryResult{
			FilePath:   path,
			Query:      match,
			Normalized: normalizeQuery(match),
		}
	}
	return results, nil
}

func findDuplicates(queries []QueryResult) map[string][]QueryResult {
	duplicates := make(map[string][]QueryResult)
	for _, query := range queries {
		duplicates[query.Normalized] = append(duplicates[query.Normalized], query)
	}

	for key, value := range duplicates {
		if len(value) == 1 {
			delete(duplicates, key)
		}
	}
	return duplicates
}

func printResults(duplicates map[string][]QueryResult) {
	if len(duplicates) == 0 {
		fmt.Println("No duplicate queries found")
		return
	}

	fmt.Printf("Found %d duplicate queries\n", len(duplicates))

	// Convert map keys to slice for sorting
	keys := make([]string, 0, len(duplicates))
	for k := range duplicates {
		keys = append(keys, k)
	}

	// Sort by number of values (descending) and alphabetically for equal counts
	sort.Slice(keys, func(i, j int) bool {
		if len(duplicates[keys[i]]) != len(duplicates[keys[j]]) {
			return len(duplicates[keys[i]]) > len(duplicates[keys[j]])
		}
		return keys[i] < keys[j]
	})

	// Print sorted results
	for _, k := range keys {
		fmt.Printf("Count: %d -- Normalized Query:\t %s\n", len(duplicates[k]), k)
	}
}

func normalizeQuery(query string) string {
	// First collapse all whitespace variants into single spaces
	normalized := regexp.MustCompile(`[\s\n\r\t]+`).ReplaceAllString(query, " ")
	normalized = strings.TrimSpace(normalized)
	normalized = strings.ToLower(normalized)

	replacements := []struct {
		pattern     string
		replacement string
	}{
		{`\s*=\s*`, " = "},  // Normalize spaces around equals
		{`\s*,\s*`, ", "},   // Normalize spaces around commas
		{`\s+`, " "},        // Any remaining multiple spaces to single
		{`\d+`, "N"},        // Numbers to N
		{`'[^']*'`, "S"},    // Quoted strings to S
		{`"[^"]*"`, "S"},    // Double quoted strings to S
		{`\s*\(\s*`, " ( "}, // Normalize spaces around parentheses
		{`\s*\)\s*`, " ) "},
	}

	for _, r := range replacements {
		re := regexp.MustCompile(r.pattern)
		normalized = re.ReplaceAllString(normalized, r.replacement)
	}

	return normalized
}

func findSQLQueries(text string) []string {
	// More comprehensive SQL pattern
	pattern := `(?i)(?:SELECT\s+[\s\S]+?(?:FROM[\s\S]+?)?|` +
		`INSERT\s+INTO[\s\S]+?|` +
		`UPDATE\s+\w+\s+SET[\s\S]+?|` +
		`DELETE\s+FROM[\s\S]+?|` +
		`CREATE\s+(?:TABLE|DATABASE|INDEX)[\s\S]+?|` +
		`ALTER\s+TABLE[\s\S]+?|` +
		`DROP\s+(?:TABLE|DATABASE)[\s\S]+?|` +
		`TRUNCATE\s+TABLE[\s\S]+?)` +
		`(?:;|$)` // Match until semicolon or end of string

	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(text, -1)

	// Clean and validate matches
	var result []string
	for _, match := range matches {
		// Clean up the match
		cleaned := strings.TrimSpace(match)

		// Basic validation that it looks like a SQL query
		if len(cleaned) > 0 &&
			(strings.HasSuffix(cleaned, ";") ||
				strings.Contains(strings.ToUpper(cleaned), "SELECT") ||
				strings.Contains(strings.ToUpper(cleaned), "INSERT") ||
				strings.Contains(strings.ToUpper(cleaned), "UPDATE")) {

			result = append(result, cleaned)
		}
	}
	return result
}

func main() {
	config := parseFlags()
	files, err := findFiles(config)
	if err != nil {
		fmt.Printf("Error walking folder: %v\n", err)
		return
	}

	queries := processFiles(files, config)
	duplicates := findDuplicates(queries)
	printResults(duplicates)
}
