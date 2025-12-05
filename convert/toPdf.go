package convert

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/douxiyou/md-to-pdf/chrome"
	"github.com/labstack/echo/v4"
	"github.com/russross/blackfriday/v2"
)

func HandleMarkdownToPdf(ctx echo.Context) error {
	startTime := time.Now()
	w := ctx.Response().Writer
	r := ctx.Request() // 直接使用指针，不需要解引用
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=output.pdf")

	// 调用转换函数
	pdfData, err := convertMarkdownToPdf(body)
	if err != nil {
		http.Error(w, "Conversion failed: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	fmt.Println("处理耗时：", time.Since(startTime))

	// 以ArrayBuffer形式返回PDF数据
	_, writeErr := w.Write(pdfData)
	if writeErr != nil {
		http.Error(w, "Error writing response: "+writeErr.Error(), http.StatusInternalServerError)
		return writeErr
	}
	return nil
}

func convertMarkdownToPdf(markdownData []byte) ([]byte, error) {
	// 1. 将ArrayBuffer解码为字符串
	markdownText := string(markdownData)

	// 2. Markdown转换为HTML
	htmlBytes := blackfriday.Run([]byte(markdownText))

	// 3. HTML转换为PDF
	return convertHtmlToPdf(string(htmlBytes))
}

// convertHtmlToPdf 基于 chromedp 将 HTML 内容转换为 PDF 二进制数据
// htmlContent: 待转换的 HTML 字符串（需包含完整的 HTML 结构）
// 返回：PDF 二进制数据、错误信息
func convertHtmlToPdf(htmlContent string) ([]byte, error) {
	// 检查Chrome实例是否已初始化
	if !chrome.IsInitialized() {
		log.Println("Chrome实例未初始化或已崩溃，重新初始化...")
		chrome.Reset()
		err := chrome.InitGlobalChrome()
		if err != nil {
			log.Printf("Chrome实例初始化失败: %v", err)
			return nil, fmt.Errorf("failed to initialize chrome: %w", err)
		}
	}

	// 更新最后使用时间
	chrome.UpdateLastUsed()

	// 创建 chromedp 上下文，设置整体超时（避免无限阻塞）
	ctx, cancel := chromedp.NewContext(chrome.GlobalAllocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// 设置上下文超时（根据实际需求调整，如 30 秒）
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pdfData []byte

	// 执行 chromedp 任务链：导航到 HTML 内容 → 等待页面加载 → 生成 PDF
	err := chromedp.Run(ctx,
		// 导航到 HTML 数据 URI
		chromedp.Navigate("data:text/html;charset=utf-8,"+htmlContent),
		// 等待 body 元素可见
		chromedp.WaitVisible("body", chromedp.ByQuery),
		// 执行 PDF 生成操作
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 调用 Chrome DevTools Protocol 的 PrintToPDF 方法
			buf, _, err := chrome.GlobalPdfParams.Do(ctx)
			if err != nil {
				return err
			}
			pdfData = buf
			return nil
		}),
	)

	if err != nil {
		// 如果出错，标记实例可能已损坏
		log.Printf("ChromeDP执行出错: %v", err)
		// 补充错误提示，方便排查
		return nil, fmt.Errorf("chromedp 生成 PDF 失败: %w", err)
	}

	return pdfData, nil
}
