package convert

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/labstack/echo/v4"
	"github.com/russross/blackfriday/v2"
)

func HandleMarkdownToPdf(ctx echo.Context) error {
	w := ctx.Response().Writer
	r := *ctx.Request()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=output.pdf")
	// 调用转换函数
	pdfData, err := convertMarkdownToPdf(body)
	if err != nil {
		http.Error(w, "Conversion failed", http.StatusInternalServerError)
		return err
	}
	// 以ArrayBuffer形式返回PDF数据
	_, writeErr := w.Write(pdfData)
	if writeErr != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
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
	// 1. 配置 chromedp 上下文选项：启用无头模式、设置超时、禁用图片加载（可选，提升速度）
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoSandbox,                     // 禁用沙箱（服务器环境必须，否则权限不足）
		chromedp.Headless,                      // 启用无头模式（无 GUI 环境必备）
		chromedp.DisableGPU,                    // 禁用 GPU 加速（服务器无显卡）
		chromedp.Flag("disable-images", false), // 若无需图片，设为 true 提升速度
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	}

	// 创建带配置的执行器上下文
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	// 2. 创建 chromedp 上下文，设置整体超时（避免无限阻塞）
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	// 设置上下文超时（根据实际需求调整，如 30 秒）
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 3. 对 HTML 内容进行 URL 编码，避免特殊字符导致页面解析失败
	encodedHTML := url.QueryEscape(htmlContent)
	// 若 HTML 内容过大，改用 base64 编码（QueryEscape 对大内容支持不佳）
	// encodedHTML = "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(htmlContent))

	// 4. 自定义 PDF 生成配置：页面尺寸、边距、是否打印背景等
	pdfParams := page.PrintToPDFParams{
		//PageWidth:        8.27,  // A4 宽度（英寸）
		//PageHeight:       11.69, // A4 高度（英寸）
		MarginTop:         0.4, // 上边距（英寸）
		MarginBottom:      0.4,
		MarginLeft:        0.4,
		MarginRight:       0.4,
		PrintBackground:   true, // 打印背景色和图片（关键，否则 CSS 背景失效）
		PreferCSSPageSize: true, // 优先使用 CSS 定义的页面尺寸
	}

	var pdfData []byte
	// 5. 执行 chromedp 任务链：导航到 HTML 内容 → 等待页面加载 → 生成 PDF
	err := chromedp.Run(ctx,
		// 导航到编码后的 HTML 数据 URI
		chromedp.Navigate("data:text/html;charset=utf-8,"+encodedHTML),
		// 等待页面完全加载（比 WaitVisible 更严谨，等待 DOMContentLoaded 事件）
		chromedp.WaitReady("body", chromedp.ByQuery),
		// 可选：等待额外时间，确保动态内容（如 JS 渲染）加载完成
		chromedp.Sleep(500*time.Millisecond),
		// 执行 PDF 生成操作
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 调用 Chrome DevTools Protocol 的 PrintToPDF 方法
			buf, _, err := pdfParams.Do(ctx)
			if err != nil {
				return err
			}
			pdfData = buf
			return nil
		}),
	)

	if err != nil {
		// 补充错误提示，方便排查
		return nil, fmt.Errorf("chromedp 生成 PDF 失败: %w", err)
	}

	return pdfData, nil
}
