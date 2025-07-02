package screenshot

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"time"
	"github.com/chromedp/chromedp"
)

func TakeScreenshot(url, savePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ctx, cancelBrowser := chromedp.NewContext(ctx)
	defer cancelBrowser()

	var buf []byte

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		return err
	}

	return os.WriteFile(savePath, buf, 0644)
}

func ExtractText(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ctx, cancelBrowser := chromedp.NewContext(ctx)
	defer cancelBrowser()

	var visibleText string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.Text("body", &visibleText, chromedp.NodeVisible),
	)
	if err == nil && len(visibleText) > 0 {
		return visibleText, nil
	}

	// fallback 
	resp, err2 := http.Get(url)
	if err2 != nil {
		return "", errors.New("chromedp error: " + err.Error() + "; http fallback error: " + err2.Error())
	}
	defer resp.Body.Close()

	bodyBytes, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		return "", errors.New("chromedp error: " + err.Error() + "; http fallback read error: " + err2.Error())
	}

	return string(bodyBytes), nil
}
