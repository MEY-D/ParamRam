package main

import (
"bufio"
"flag"
"fmt"
"io/ioutil"
"net/http"
"os"
"path/filepath"
"regexp"
"github.com/PuerkitoBio/goquery"
)

func main() {
save := flag.Bool("s", false, "Save the result")
flag.Parse()

scanner := bufio.NewScanner(os.Stdin)
urls := make([]string, 0)
for scanner.Scan() {
urls = append(urls, scanner.Text())
}

if err := scanner.Err(); err != nil {
fmt.Printf("Error reading from stdin: %v\n", err)
return
}

parameters := make(map[string]struct{})
for _, url := range urls {
params, err := extractParameters(url)
if err != nil {
fmt.Printf("Error extracting parameters from %s: %v\n", url, err)
continue
}

for _, param := range params {
parameters[param] = struct{}{}
}
}

if *save {
homeDir, err := os.UserHomeDir()
if err != nil {
fmt.Printf("Error getting home directory: %v\n", err)
return
}

outputPath := filepath.Join(homeDir, "database", "parameters.txt")
err = os.MkdirAll(filepath.Dir(outputPath), 0755)
if err != nil {
fmt.Printf("Error creating directory: %v\n", err)
return
}

err = saveParameters(outputPath, parameters)
if err != nil {
fmt.Printf("Error saving parameters: %v\n", err)
return
}
} else {
for param := range parameters {
fmt.Println(param)
}
}
}

func extractParameters(url string) ([]string, error) {
resp, err := http.Get(url)
if err != nil {
return nil, err
}
defer resp.Body.Close()

doc, err := goquery.NewDocumentFromReader(resp.Body)
if err != nil {
return nil, err
}

regex := regexp.MustCompile(`[a-zA-Z]+[a-zA-Z0-9_]*`)
parameters := make([]string, 0)

doc.Find("*").Each(func(_ int, s *goquery.Selection) {
for _, attr := range []string{"name", "id", "class"} {
value, exists := s.Attr(attr)
if exists {
matches := regex.FindAllString(value, -1)
parameters = append(parameters, filterMatches(matches)...)
}
}
})

doc.Find("link[href], script[src]").Each(func(_ int, s *goquery.Selection) {
var attr string
if s.Is("link") {
attr = "href"
} else {
attr = "src"
}

href, exists := s.Attr(attr)
if exists {
resp, err := http.Get(href)
if err != nil {
return
}
defer resp.Body.Close()

content, err := ioutil.ReadAll(resp.Body)
if err != nil {
return
}

matches := regex.FindAllString(string(content), -1)
parameters = append(parameters, filterMatches(matches)...)
}
})

return parameters, nil
}

func filterMatches(matches []string) []string {
filtered := make([]string, 0)
underscoreRegex := regexp.MustCompile(`^[a-zA-Z]+(_[a-zA-Z0-9]+){0,2}$`)
for _, match := range matches {
if len(match) <= 15 && underscoreRegex.MatchString(match) {
filtered = append(filtered, match)
}
}
return filtered
}

func saveParameters(filename string, parameters map[string]struct{}) error {
file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
if err != nil {
return err
}
defer file.Close()

for param := range parameters {
_, err := file.WriteString(param + "\n")
if err != nil {
return err
}
}

return nil
}

