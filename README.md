# Duplicate Query Finder

This is a somewhat simple Go program that finds duplicate SQL queries in a given directory.

This is still very much an "alpha" version and is not yet ready for production use.

## Usage

```bash
Usage of ./bin/duplicate-query:
  -folder string
        Folder path to scan (default ".")
  -ignore string
        Comma separated list of folders to ignore (default "vendor,node_modules")
  -type string
        File type to scan (default ".php")
  -workers int
        Number of worker goroutines (default Number of logical CPUs)
        
        
# Example
./bin/duplicate-query -folder=/path/to/folder -type=".php" -ignore="vendor,node_modules"
```