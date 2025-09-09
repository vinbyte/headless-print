// Command pdf is a chromedp example demonstrating how to capture a pdf of a
// page.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// TabPool manages a pool of browser tabs for parallel processing
type TabPool struct {
	browserCtx    context.Context
	browserCancel context.CancelFunc
	pool          chan context.Context
	mutex         sync.Mutex
}

// Global tab pool
var tabPool *TabPool

var isUseCDPServer = true

// default cdp server
var cdpServerURL = "ws://127.0.0.1:9222"

// NewTabPool creates and initializes a new tab pool
func NewTabPool(size int) (*TabPool, error) {
	// Setup Chrome options for Alpine environment
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// Disables Chrome's security sandbox
		chromedp.Flag("no-sandbox", true),
		// Disables setuid sandbox (often used with --no-sandbox)
		chromedp.Flag("disable-setuid-sandbox", true),
		// Disables zygote (chrome process for faster startup) process
		chromedp.Flag("no-zygote", true),
		// Uses /tmp instead of shared memory for Docker
		chromedp.Flag("disable-dev-shm-usage", true),
		// Disables GPU hardware acceleration
		chromedp.Flag("disable-gpu", true),
		// Disables browser extensions
		chromedp.Flag("disable-extensions", true),
		// Disables plugins
		chromedp.Flag("disable-plugins", true),
		// Prevents Chrome from slowing down timers in background tabs
		chromedp.Flag("disable-background-timer-throttling", true),
		// Keeps occluded windows active
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		// Prevents renderer processes from backgrounding
		chromedp.Flag("disable-renderer-backgrounding", true),
		// Disables web security
		chromedp.Flag("disable-web-security", true),
		// Reduces CPU usage for rendering
		chromedp.Flag("disable-software-rasterizer", true),
		// Prevents loading of default Chrome apps
		chromedp.Flag("disable-default-apps", true),
		// Disables component extensions with background pages
		chromedp.Flag("disable-component-extensions-with-background-pages", true),
		// Disables Google Account syncing
		chromedp.Flag("disable-sync", true),
		// Runs Chrome in headless mode
		chromedp.Flag("headless", true),
	)

	// Create allocator context
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	if isUseCDPServer {
		log.Println("Using CDPServer:", cdpServerURL)
		// Create a remote allocator with NoModifyURL option
		allocCtx, cancel = chromedp.NewRemoteAllocator(
			context.Background(),
			cdpServerURL,
			// chromedp.NoModifyURL,
		)
	}

	// Create browser context
	ctx, browserCancel := chromedp.NewContext(allocCtx)

	// Start the browser
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start browser: %v", err)
	}

	// Create the tab pool
	pool := &TabPool{
		browserCtx:    ctx,
		browserCancel: browserCancel,
		pool:          make(chan context.Context, size),
		mutex:         sync.Mutex{},
	}

	// Initialize tabs in the pool
	for i := 0; i < size; i++ {
		tabCtx, _ := chromedp.NewContext(ctx)
		// Pre-navigate to blank page to initialize the tab
		err := chromedp.Run(tabCtx, chromedp.Navigate("about:blank"))
		if err != nil {
			log.Printf("Warning: Failed to initialize tab %d: %v", i, err)
		}
		pool.pool <- tabCtx
	}

	return pool, nil
}

// GetTab gets a tab from the pool
func (p *TabPool) GetTab() context.Context {
	return <-p.pool
}

// ReleaseTab returns a tab to the pool
func (p *TabPool) ReleaseTab(ctx context.Context) {
	p.pool <- ctx
}

// Close closes all tabs and the browser
func (p *TabPool) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Close the browser context which will close all tabs
	p.browserCancel()
}

// Calculate optimal tab pool size based on available resources
func calculateOptimalPoolSize() int {
	// Get number of CPUs available to the container
	numCPU := runtime.NumCPU()

	// Each Chrome tab can use approximately 0.25 CPU cores
	// So we can have 4 tabs per CPU core
	optimalSize := numCPU * 4

	// Limit to a reasonable range
	if optimalSize < 4 {
		optimalSize = 4 // Minimum 4 tabs
	} else if optimalSize > 32 {
		optimalSize = 32 // Maximum 32 tabs
	}

	return optimalSize
}

func main() {
	useCDPServer := flag.Bool("use-cdp-server", false, "use cdp server")
	url := flag.String("url", "ws://127.0.0.1:9222", "devtools url")
	flag.Parse()
	cdpServerURL = *url
	isUseCDPServer = *useCDPServer

	// Initialize the tab pool with optimal size based on available resources
	log.Println("Initializing browser tab pool...")
	poolSize := calculateOptimalPoolSize()
	log.Printf("Calculated optimal pool size: %d tabs based on available resources", poolSize)

	var err error
	tabPool, err = NewTabPool(poolSize)
	if err != nil {
		log.Fatalf("Failed to initialize tab pool: %v", err)
	}
	log.Printf("Browser tab pool initialized successfully with %d tabs", poolSize)

	// Ensure browser is closed when the program exits
	defer tabPool.Close()

	// Define HTTP routes
	http.HandleFunc("/generate-pdf", generatePDFHandler)

	// Start the server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

// generatePDFHandler handles the PDF generation request
func generatePDFHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := createPDF()
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=generated.pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(buf)))

	// Write PDF to response
	if _, err := w.Write(buf); err != nil {
		log.Printf("Error writing response: %v", err)
	}
	log.Println("PDF generated and sent to client")
}

// generatePDF generates a PDF and saves it to disk
func generatePDF() {
	buf, err := createPDF()
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("sample.pdf", buf, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote sample.pdf")
}

func createPDF() ([]byte, error) {
	// Get a tab from the pool
	log.Println("Getting tab from pool")
	tabCtx := tabPool.GetTab()
	defer tabPool.ReleaseTab(tabCtx) // Return the tab to the pool when done

	log.Println("Read HTML file")
	html, err := os.ReadFile("index.html")
	if err != nil {
		return nil, err
	}

	var buf []byte
	log.Println("Using tab from pool")

	// Create a context with timeout
	ctxWithTimeout, cancelTimeout := context.WithTimeout(tabCtx, 10*time.Second)
	defer cancelTimeout()

	if err := chromedp.Run(ctxWithTimeout,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Set page content")

			// Get the frame tree
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				log.Println("Error getting frame tree:", err)
				return err
			}

			// Set the document content
			if err := page.SetDocumentContent(frameTree.Frame.ID, string(html)).Do(ctx); err != nil {
				log.Println("Error setting document content:", err)
				return err
			}

			// Wait a short time for the content to load
			log.Println("Waiting for content to load...")
			time.Sleep(1 * time.Second)
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Print to PDF")
			var err error
			buf, _, err = page.PrintToPDF().WithPrintBackground(false).Do(ctx)
			if err != nil {
				log.Println("Error printing to PDF:", err)
				return err
			}
			log.Println("PDF generated successfully")
			return nil
		}),
	); err != nil {
		log.Println("Error running chromedp:", err)
		return nil, err
	}

	log.Println("Return buffer")
	return buf, nil
}
