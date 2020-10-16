package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	Run()
}

func Run() {
	r := gin.New()
	gin.DisableConsoleColor()

	accessLogger := log.New(os.Stdout, "", log.LstdFlags)
	MakeRotateLogger("log", "access_", "2010-10-16", accessLogger)

	r.Use(gin.Recovery())
	r.Use(func (c *gin.Context) {
		start := time.Now()

		c.Next()

		end := time.Now()

		loggingData := []string {
			c.ClientIP(),
			start.Format(time.RFC3339),
			end.Format(time.RFC3339),
			fmt.Sprintf("%.3f", (float64(end.Sub(start).Nanoseconds()) / (1000 * 1000.0))),
			c.Request.Method,
			c.Request.URL.String(),
			c.Request.Proto,
			c.Request.Referer(),
			c.Request.UserAgent(),
			c.Request.Host,
			fmt.Sprintf("%d", c.Writer.Status()),
			fmt.Sprintf("%d", c.Writer.Size()),
			c.Request.Header.Get("Content-length"),
			fmt.Sprintf("%+v", c.Errors),
			fmt.Sprintf("%+v", c.Keys)}

		for idx, val := range loggingData {
			if len(val) == 0 {
				loggingData[idx] = `""`
			}
		}
		accessLogger.Printf("%s", strings.Join(loggingData, "\t"))
	})

	r.GET("/", imgShot)

	srv := &http.Server{
		Addr: ":8888",
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

type urlArg struct {
	Url string `form:"url"`
}

func imgShot(c *gin.Context)  {
	arg := urlArg{}
	if err := c.Bind(&arg); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Start Chrome
	// Remove the 2nd param if you don't need debug information logged
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
	defer cancel()

	userAgent := "Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko"
	width := 1920
	height := 3840

	var imageBuf []byte
	if err := chromedp.Run(ctx, ScreenshotTasks(arg.Url, userAgent, width, height, &imageBuf)); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "image/png", imageBuf)
}

func ScreenshotTasks(url, userAgent string, width, height int, imageBuf *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		emulation.SetDeviceMetricsOverride(int64(width), int64(height), 1.0, false),
		emulation.SetUserAgentOverride(userAgent),
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			*imageBuf, err = page.CaptureScreenshot().WithQuality(90).Do(ctx)
			return err
		}),
	}
}
