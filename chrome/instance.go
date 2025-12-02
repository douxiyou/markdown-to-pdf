package chrome

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// 全局 Chrome 实例池（复用 ExecAllocator）
var (
	GlobalAllocCtx  context.Context
	GlobalCancel    context.CancelFunc
	GlobalPdfParams page.PrintToPDFParams
	poolOnce        sync.Once
	mu              sync.RWMutex
	lastUsed        time.Time
)

// GetLastUsed 获取实例最后使用时间
func GetLastUsed() time.Time {
	mu.RLock()
	defer mu.RUnlock()
	return lastUsed
}

// UpdateLastUsed 更新实例最后使用时间
func UpdateLastUsed() {
	mu.Lock()
	defer mu.Unlock()
	lastUsed = time.Now()
}

// IsInitialized 检查Chrome实例是否已初始化
func IsInitialized() bool {
	mu.RLock()
	defer mu.RUnlock()
	return GlobalAllocCtx != nil && GlobalCancel != nil
}

// Reset 重置Chrome实例
func Reset() {
	mu.Lock()
	defer mu.Unlock()

	if GlobalCancel != nil {
		GlobalCancel()
	}

	GlobalAllocCtx = nil
	GlobalCancel = nil
}

// InitGlobalChrome 初始化全局 Chrome 实例（仅启动一次）
func InitGlobalChrome() {
	poolOnce.Do(func() {
		initializeChrome()
	})

	GlobalPdfParams = page.PrintToPDFParams{
		MarginTop:         0.4,
		MarginBottom:      0.4,
		MarginLeft:        0.4,
		MarginRight:       0.4,
		PrintBackground:   false,
		PreferCSSPageSize: true,
	}
}

// initializeChrome 初始化Chrome实例的实际实现
func initializeChrome() {
	// 优化后的启动参数（禁用所有冗余功能）
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),                              // 新无头模式（更快）
		chromedp.Flag("disable-gpu", true),                            // 禁用 GPU
		chromedp.Flag("no-sandbox", true),                             // 禁用沙箱（Linux 必需）
		chromedp.Flag("disable-dev-shm-usage", true),                  // 规避共享内存限制
		chromedp.Flag("disable-extensions", true),                     // 禁用扩展
		chromedp.Flag("disable-plugins", true),                        // 禁用插件
		chromedp.Flag("disable-images", true),                         // 禁用图片加载（非必需图片）
		chromedp.Flag("disable-javascript", false),                    // 按需禁用 JS（若 HTML 无需 JS 渲染）
		chromedp.Flag("disable-css", false),                           // 禁用 CSS（仅纯文本 HTML 可用）
		chromedp.Flag("disable-web-security", true),                   // 禁用跨域检查（加速资源加载）
		chromedp.Flag("ignore-certificate-errors", true),              // 忽略 SSL 错误
		chromedp.Flag("disable-background-timer-throttling", true),    // 禁用后台定时器节流
		chromedp.Flag("disable-renderer-backgrounding", true),         // 禁用渲染器后台化
		chromedp.Flag("disable-backgrounding-occluded-windows", true), // 禁用后台窗口
		chromedp.NoFirstRun,                                           // 禁用首次运行引导
		chromedp.NoDefaultBrowserCheck,                                // 禁用默认浏览器检查
		chromedp.Flag("single-process", true),                         // 单进程模式（减少进程切换开销）
		chromedp.Flag("disable-background-networking", true),          // 禁用后台网络请求
	)

	// 创建全局 ExecAllocator（Chrome 进程）
	GlobalAllocCtx, GlobalCancel = chromedp.NewExecAllocator(context.Background(), opts...)
	log.Println("全局 Chrome 实例初始化完成")
	UpdateLastUsed()
}
