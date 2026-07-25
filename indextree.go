package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rix4uni/indextree/banner"
	"github.com/spf13/pflag"
	"golang.org/x/net/html"
)

// ANSI terminal color codes
const (
	ansiReset  = "\033[0m"
	ansiBlue   = "\033[1;34m" // Bold Blue for directories
	ansiGreen  = "\033[32m"   // Green for files and success statistics
	ansiYellow = "\033[33m"   // Yellow for warnings and metadata details
	ansiRed    = "\033[31m"   // Red for errors
	ansiCyan   = "\033[36m"   // Cyan for request progress and tree branch lines
	ansiGray   = "\033[90m"   // Gray for dim/auxiliary labels
)

// Node represents a node in the directory tree
type Node struct {
	Name         string
	IsDir        bool
	Size         string
	LastModified string
	Children     []*Node
}

// Crawler config
type Config struct {
	StartURL   string
	MaxDepth   int
	Delay      time.Duration
	OutputFile string
	Verbose    bool
	Retries    int
	UseColor   bool
}

// Stats tracker
type Stats struct {
	mu           sync.Mutex
	DirsScanned  int
	FilesFound   int
	DirsFound    int
	RequestCount int
}

func (s *Stats) IncDirsScanned() {
	s.mu.Lock()
	s.DirsScanned++
	s.mu.Unlock()
}

func (s *Stats) IncFilesFound() {
	s.mu.Lock()
	s.FilesFound++
	s.mu.Unlock()
}

func (s *Stats) IncDirsFound() {
	s.mu.Lock()
	s.DirsFound++
	s.mu.Unlock()
}

func (s *Stats) IncRequestCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RequestCount++
	return s.RequestCount
}

func (s *Stats) Load() (int, int, int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.DirsScanned, s.DirsFound, s.FilesFound, s.RequestCount
}

// Sem represents a channel-based semaphore for concurrency control
type Sem struct {
	ch chan struct{}
}

func NewSem(n int) *Sem {
	if n <= 0 {
		n = 1
	}
	return &Sem{ch: make(chan struct{}, n)}
}

func (s *Sem) Acquire() {
	s.ch <- struct{}{}
}

func (s *Sem) Release() {
	<-s.ch
}

type parsedItem struct {
	href         string
	lastModified string
	size         string
}

func main() {
	depth := pflag.Int("depth", 1, "Maximum recursion depth (0 for unlimited)")
	delay := pflag.Int("delay", 100, "Delay between requests in milliseconds")
	verbose := pflag.Bool("verbose", false, "Verbose output (show crawl details/progress)")
	output := pflag.String("output", "", "Output file path (default is stdout)")
	outputFormat := pflag.String("output-format", "tree", "Output format: tree or plain")
	concurrency := pflag.Int("concurrency", 50, "Maximum number of concurrent HTTP requests")
	retries := pflag.Int("retries", 3, "Number of retries for failed HTTP requests")
	color := pflag.String("color", "auto", "Color output (always, never, auto)")
	silent := pflag.Bool("silent", false, "Silent mode.")
	insecure := pflag.Bool("insecure", false, "Skip TLS certificate verification")
	version := pflag.Bool("version", false, "Print the version of the tool and exit.")
	pflag.Parse()

	if *version {
		banner.PrintBanner()
		banner.PrintVersion()
		return
	}

	if !*silent {
		banner.PrintBanner()
	}

	var useColor bool
	switch strings.ToLower(*color) {
	case "always":
		useColor = true
	case "never":
		useColor = false
	default: // "auto"
		useColor = (*output == "")
	}

	// Determine output destination
	var out io.Writer = os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
		if *verbose {
			fmt.Fprintf(os.Stderr, "Writing tree output to %s...\n", *output)
		}
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: *insecure,
		},
	}

	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}

	sem := NewSem(*concurrency)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		startURL := strings.TrimSpace(line)
		if startURL == "" {
			continue
		}

		// Normalize start URL
		parsedURL, err := url.Parse(startURL)
		if err != nil {
			if useColor {
				fmt.Fprintf(os.Stderr, "%sError: Invalid URL %q: %v%s\n", ansiRed, startURL, err, ansiReset)
			} else {
				fmt.Fprintf(os.Stderr, "Error: Invalid URL %q: %v\n", startURL, err)
			}
			continue
		}

		// If path doesn't end with slash, append it so baseURL.Parse handles relative paths correctly
		if !strings.HasSuffix(parsedURL.Path, "/") {
			parsedURL.Path += "/"
		}
		startURL = parsedURL.String()

		cfg := Config{
			StartURL:   startURL,
			MaxDepth:   *depth,
			Delay:      time.Duration(*delay) * time.Millisecond,
			OutputFile: *output,
			Verbose:    *verbose,
			Retries:    *retries,
			UseColor:   useColor,
		}

		if *verbose {
			if useColor {
				fmt.Fprintf(os.Stderr, "%sStarting crawl at:%s %s\n", ansiCyan, ansiReset, cfg.StartURL)
				if cfg.MaxDepth > 0 {
					fmt.Fprintf(os.Stderr, "%sMax depth limit:%s %s%d%s\n", ansiGray, ansiReset, ansiGreen, cfg.MaxDepth, ansiReset)
				} else {
					fmt.Fprintf(os.Stderr, "%sMax depth limit:%s %sUnlimited (0)%s\n", ansiGray, ansiReset, ansiGreen, ansiReset)
				}
				fmt.Fprintf(os.Stderr, "%sRequest delay:%s %s%v%s\n", ansiGray, ansiReset, ansiGreen, cfg.Delay, ansiReset)
				fmt.Fprintf(os.Stderr, "%sConcurrency limit:%s %s%d%s\n", ansiGray, ansiReset, ansiGreen, *concurrency, ansiReset)
				fmt.Fprintf(os.Stderr, "%sHTTP Retries:%s %s%d%s\n", ansiGray, ansiReset, ansiGreen, *retries, ansiReset)
			} else {
				fmt.Fprintf(os.Stderr, "Starting crawl at: %s\n", cfg.StartURL)
				if cfg.MaxDepth > 0 {
					fmt.Fprintf(os.Stderr, "Max depth limit: %d\n", cfg.MaxDepth)
				} else {
					fmt.Fprintln(os.Stderr, "Max depth limit: Unlimited (0)")
				}
				fmt.Fprintf(os.Stderr, "Request delay: %v\n", cfg.Delay)
				fmt.Fprintf(os.Stderr, "Concurrency limit: %d\n", *concurrency)
				fmt.Fprintf(os.Stderr, "HTTP Retries: %d\n", *retries)
			}
		}

		stats := &Stats{}
		visited := &sync.Map{}

		startTime := time.Now()
		rootNode, err := crawl(client, cfg.StartURL, parsedURL, 1, &cfg, visited, sem, stats)
		if err != nil {
			if useColor {
				fmt.Fprintf(os.Stderr, "%sError during crawl of %s: %v%s\n", ansiRed, startURL, err, ansiReset)
			} else {
				fmt.Fprintf(os.Stderr, "Error during crawl of %s: %v\n", startURL, err)
			}
			continue
		}

		duration := time.Since(startTime)

		// Print tree structure
		if rootNode != nil {
			switch *outputFormat {
			case "plain":
				PrintPlain(out, rootNode, cfg.StartURL, useColor)
			default:
				PrintTree(out, rootNode, useColor)
			}
		}

		// Print crawling statistics (only if verbose)
		if *verbose {
			scanned, foundDirs, foundFiles, totalReqs := stats.Load()
			summaryHeader := fmt.Sprintf("\nCrawl summary for %s:", startURL)
			if useColor {
				summaryHeader = ansiCyan + summaryHeader + ansiReset
			}
			fmt.Fprintln(os.Stderr, summaryHeader)

			printStatLine(os.Stderr, "Directories scanned", fmt.Sprintf("%d", scanned), useColor)
			printStatLine(os.Stderr, "Directories found", fmt.Sprintf("%d", foundDirs), useColor)
			printStatLine(os.Stderr, "Files found", fmt.Sprintf("%d", foundFiles), useColor)
			printStatLine(os.Stderr, "Total requests made", fmt.Sprintf("%d", totalReqs), useColor)
			printStatLine(os.Stderr, "Duration", fmt.Sprintf("%v", duration.Round(time.Millisecond)), useColor)
			fmt.Fprintln(os.Stderr, "")
		}
	}

	if err := scanner.Err(); err != nil {
		if useColor {
			fmt.Fprintf(os.Stderr, "%sError reading standard input: %v%s\n", ansiRed, err, ansiReset)
		} else {
			fmt.Fprintf(os.Stderr, "Error reading standard input: %v\n", err)
		}
	}
}

func printStatLine(out io.Writer, label, value string, useColor bool) {
	if useColor {
		fmt.Fprintf(out, "  - %s%s%s: %s%s%s\n", ansiGray, label, ansiReset, ansiGreen, value, ansiReset)
	} else {
		fmt.Fprintf(out, "  - %s: %s\n", label, value)
	}
}

// crawl recursively fetches and parses the HTML directory index
func crawl(client *http.Client, currentTarget string, rootURL *url.URL, currentDepth int, cfg *Config, visited *sync.Map, sem *Sem, stats *Stats) (*Node, error) {
	currentURL, err := url.Parse(currentTarget)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", currentTarget, err)
	}

	// Mark path as visited to prevent duplicate scans
	visited.Store(currentURL.Path, true)
	stats.IncDirsScanned()

	// Delay between requests to respect server rates
	if cfg.Delay > 0 {
		time.Sleep(cfg.Delay)
	}

	// Acquire semaphore and fetch HTML listing
	sem.Acquire()

	var parsedItems []parsedItem
	var baseURL *url.URL

	err = func() error {
		defer sem.Release()

		reqCount := stats.IncRequestCount()
		req, err := http.NewRequest("GET", currentURL.String(), nil)
		if err != nil {
			return err
		}
		// Add user-agent header to appear polite and standard
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) indextree-crawler/1.0")

		if cfg.Verbose {
			if cfg.UseColor {
				fmt.Fprintf(os.Stderr, "%s[Request %d]%s Crawling: %s\n", ansiCyan, reqCount, ansiReset, currentURL.String())
			} else {
				fmt.Fprintf(os.Stderr, "[Request %d] Crawling: %s\n", reqCount, currentURL.String())
			}
		}

		resp, err := doRequestWithRetry(client, req, cfg.Retries, cfg.Verbose, cfg.UseColor)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		baseURL = resp.Request.URL
		parsedItems, err = parseDirectoryListing(resp.Body)
		return err
	}()

	if err != nil {
		return nil, err
	}

	// Determine node name for display
	displayName := path.Base(baseURL.Path)
	if displayName == "/" || displayName == "." || displayName == "" {
		displayName = rootURL.Host
	}

	node := &Node{
		Name:  displayName,
		IsDir: true,
	}

	type linkItem struct {
		name         string
		url          string
		isDir        bool
		size         string
		lastModified string
	}
	var items []linkItem

	for _, item := range parsedItems {
		// Resolve relative URL
		resolvedURL, err := baseURL.Parse(item.href)
		if err != nil {
			continue
		}

		// Ensure we stay on the same host and don't exit scope
		if resolvedURL.Host != rootURL.Host {
			continue
		}

		// Filter out query parameters (e.g. sorting columns '?C=N;O=D')
		if resolvedURL.RawQuery != "" || resolvedURL.Fragment != "" {
			continue
		}

		// Check path relationships: resolvedURL.Path must be a child of baseURL.Path
		if !strings.HasPrefix(resolvedURL.Path, baseURL.Path) {
			continue
		}

		relPath := resolvedURL.Path[len(baseURL.Path):]
		trimmed := strings.TrimSuffix(relPath, "/")

		// If there is an internal slash, it is a nested subdirectory, skip it
		if trimmed == "" || strings.Contains(trimmed, "/") {
			continue
		}

		// Also check that we are not going "up" or matching ourselves
		if resolvedURL.Path == baseURL.Path {
			continue
		}

		isDir := strings.HasSuffix(relPath, "/")
		items = append(items, linkItem{
			name:         trimmed,
			url:          resolvedURL.String(),
			isDir:        isDir,
			size:         item.size,
			lastModified: item.lastModified,
		})
	}

	// Sort items: directories first, then files, both alphabetically (case-insensitive)
	sort.Slice(items, func(i, j int) bool {
		if items[i].isDir && !items[j].isDir {
			return true
		}
		if !items[i].isDir && items[j].isDir {
			return false
		}
		return strings.ToLower(items[i].name) < strings.ToLower(items[j].name)
	})

	// Process each resolved child concurrently
	node.Children = make([]*Node, len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		if item.isDir {
			stats.IncDirsFound()
			parsedChildURL, err := url.Parse(item.url)
			if err != nil {
				node.Children[i] = &Node{
					Name:         item.name,
					IsDir:        true,
					LastModified: item.lastModified,
				}
				continue
			}

			// Base recursion check on depth limit and visited map
			if cfg.MaxDepth == 0 || currentDepth < cfg.MaxDepth {
				if _, loaded := visited.LoadOrStore(parsedChildURL.Path, true); !loaded {
					wg.Add(1)
					go func(idx int, it linkItem) {
						defer wg.Done()
						childNode, err := crawl(client, it.url, rootURL, currentDepth+1, cfg, visited, sem, stats)
						if err != nil {
							if cfg.Verbose {
								if cfg.UseColor {
									fmt.Fprintf(os.Stderr, "%sWarning: Failed to crawl subdirectory %s: %v%s\n", ansiRed, it.url, err, ansiReset)
								} else {
									fmt.Fprintf(os.Stderr, "Warning: Failed to crawl subdirectory %s: %v\n", it.url, err)
								}
							}
							node.Children[idx] = &Node{
								Name:         it.name,
								IsDir:        true,
								LastModified: it.lastModified,
							}
						} else if childNode != nil {
							node.Children[idx] = childNode
						}
					}(i, item)
				} else {
					node.Children[i] = &Node{
						Name:         item.name,
						IsDir:        true,
						LastModified: item.lastModified,
					}
				}
			} else {
				// We hit depth limit
				node.Children[i] = &Node{
					Name:         item.name,
					IsDir:        true,
					LastModified: item.lastModified,
				}
			}
		} else {
			stats.IncFilesFound()
			node.Children[i] = &Node{
				Name:         item.name,
				IsDir:        false,
				Size:         item.size,
				LastModified: item.lastModified,
			}
		}
	}
	wg.Wait()

	return node, nil
}

// doRequestWithRetry sends HTTP request and retries on transient errors with backoff
func doRequestWithRetry(client *http.Client, req *http.Request, maxRetries int, verbose bool, useColor bool) (*http.Response, error) {
	var resp *http.Response
	var err error
	backoff := 1 * time.Second

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			if verbose {
				if useColor {
					fmt.Fprintf(os.Stderr, "%s[Retry %d/%d] Retrying request to %s in %v...%s\n", ansiYellow, i, maxRetries, req.URL.String(), backoff, ansiReset)
				} else {
					fmt.Fprintf(os.Stderr, "[Retry %d/%d] Retrying request to %s in %v...\n", i, maxRetries, req.URL.String(), backoff)
				}
			}
			time.Sleep(backoff)
			backoff *= 2
		}

		resp, err = client.Do(req)
		if err == nil {
			// Check if status is successful (200 OK)
			if resp.StatusCode == http.StatusOK {
				return resp, nil
			}

			// If it's a transient server error or rate-limiting, we retry
			if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode <= 599) {
				resp.Body.Close()
				err = fmt.Errorf("HTTP status code %d", resp.StatusCode)
				continue
			}

			// For standard redirect/not found/unauthorized, we do not retry
			return resp, nil
		}
	}
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

// parseDirectoryListing parses table rows to extract metadata
func parseDirectoryListing(body io.Reader) ([]parsedItem, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	var items []parsedItem
	var findRows func(*html.Node)
	findRows = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			item, ok := parseRow(n)
			if ok {
				items = append(items, item)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findRows(c)
		}
	}
	findRows(doc)
	return items, nil
}

func parseRow(tr *html.Node) (parsedItem, bool) {
	var tds []*html.Node
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "td" {
			tds = append(tds, c)
		}
	}

	if len(tds) < 3 {
		return parsedItem{}, false
	}

	// Find the td that contains the anchor link (different Apache layouts
	// may have an icon column before the link column, so the anchor isn't
	// always in tds[0]).
	var linkIdx int = -1
	for i, td := range tds {
		aNode := findAnchorNode(td)
		if aNode != nil {
			for _, attr := range aNode.Attr {
				if attr.Key == "href" && attr.Val != "" {
					linkIdx = i
					break
				}
			}
			if linkIdx >= 0 {
				break
			}
		}
	}
	if linkIdx < 0 || linkIdx+2 >= len(tds) {
		return parsedItem{}, false
	}

	aNode := findAnchorNode(tds[linkIdx])
	var href string
	for _, attr := range aNode.Attr {
		if attr.Key == "href" {
			href = attr.Val
			break
		}
	}
	if href == "" {
		return parsedItem{}, false
	}

	// Skip parent directory
	if strings.Contains(getTextContent(tds[linkIdx]), "Parent Directory") || href == "../" || href == "/" {
		return parsedItem{}, false
	}

	lastModified := getTextContent(tds[linkIdx+1])
	lastModified = strings.TrimSpace(lastModified)
	if strings.Contains(lastModified, "\u00a0") || lastModified == "" {
		lastModified = ""
	}

	size := getTextContent(tds[linkIdx+2])
	size = strings.TrimSpace(size)
	if size == "-" || size == "" {
		size = ""
	}

	return parsedItem{
		href:         href,
		lastModified: lastModified,
		size:         size,
	}, true
}

func findAnchorNode(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "a" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res := findAnchorNode(c)
		if res != nil {
			return res
		}
	}
	return nil
}

func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getTextContent(c))
	}
	return sb.String()
}

// PrintTree prints the directory tree structure
func PrintTree(out io.Writer, root *Node, useColor bool) {
	if root == nil {
		return
	}
	// Print root node
	var meta []string
	if root.Size != "" {
		meta = append(meta, root.Size)
	}
	if root.LastModified != "" {
		meta = append(meta, root.LastModified)
	}

	displayName := root.Name + "/"
	if useColor {
		displayName = ansiBlue + displayName + ansiReset // Bold Blue
	}

	if len(meta) > 0 {
		metaStr := strings.Join(meta, ", ")
		if useColor {
			metaStr = ansiYellow + metaStr + ansiReset // Yellow
		}
		fmt.Fprintf(out, "%s (%s)\n", displayName, metaStr)
	} else {
		fmt.Fprintf(out, "%s\n", displayName)
	}

	// Print children recursively
	for i, child := range root.Children {
		printNode(out, child, "", i == len(root.Children)-1, useColor)
	}
}

// PrintPlain prints the directory listing as a flat list of full URLs
func PrintPlain(out io.Writer, root *Node, baseURL string, useColor bool) {
	if root == nil {
		return
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	for _, child := range root.Children {
		if child == nil {
			continue
		}
		printPlainNode(out, child, baseURL, useColor)
	}
}

func printPlainNode(out io.Writer, node *Node, parentURL string, useColor bool) {
	if node == nil {
		return
	}
	fullURL := strings.TrimSuffix(parentURL, "/") + "/" + node.Name
	if node.IsDir {
		fullURL += "/"
	}

	displayURL := fullURL
	if useColor {
		if node.IsDir {
			displayURL = ansiBlue + displayURL + ansiReset
		} else {
			displayURL = ansiGreen + displayURL + ansiReset
		}
	}

	var meta []string
	if node.Size != "" {
		meta = append(meta, node.Size)
	}
	if node.LastModified != "" {
		meta = append(meta, node.LastModified)
	}

	if len(meta) > 0 {
		metaStr := strings.Join(meta, ", ")
		if useColor {
			metaStr = ansiYellow + metaStr + ansiReset
		}
		fmt.Fprintf(out, "%s (%s)\n", displayURL, metaStr)
	} else {
		fmt.Fprintf(out, "%s\n", displayURL)
	}

	// Recurse into subdirectories
	if node.IsDir {
		for _, child := range node.Children {
			if child != nil {
				printPlainNode(out, child, fullURL, useColor)
			}
		}
	}
}

// printNode prints a child node recursively with appropriate tree borders
func printNode(out io.Writer, node *Node, prefix string, isLast bool, useColor bool) {
	if node == nil {
		return
	}

	marker := "├── "
	if isLast {
		marker = "└── "
	}

	displayName := node.Name
	if node.IsDir {
		displayName += "/"
		if useColor {
			displayName = ansiBlue + displayName + ansiReset // Bold Blue
		}
	} else {
		if useColor {
			displayName = ansiGreen + displayName + ansiReset // Green
		}
	}

	var meta []string
	if node.Size != "" {
		meta = append(meta, node.Size)
	}
	if node.LastModified != "" {
		meta = append(meta, node.LastModified)
	}

	formattedMarker := marker
	formattedPrefix := prefix
	if useColor {
		formattedMarker = ansiCyan + marker + ansiReset // Cyan marker
		// Color vertical bars in prefix too if they exist
		if prefix != "" {
			formattedPrefix = strings.ReplaceAll(prefix, "│", ansiCyan+"│"+ansiReset)
		}
	}

	if len(meta) > 0 {
		metaStr := strings.Join(meta, ", ")
		if useColor {
			metaStr = ansiYellow + metaStr + ansiReset // Yellow
		}
		fmt.Fprintf(out, "%s%s%s (%s)\n", formattedPrefix, formattedMarker, displayName, metaStr)
	} else {
		fmt.Fprintf(out, "%s%s%s\n", formattedPrefix, formattedMarker, displayName)
	}

	var nextPrefix string
	if isLast {
		nextPrefix = prefix + "    "
	} else {
		nextPrefix = prefix + "│   "
	}

	for i, child := range node.Children {
		printNode(out, child, nextPrefix, i == len(node.Children)-1, useColor)
	}
}
