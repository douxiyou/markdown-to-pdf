package chrome

import (
	"context"
	"fmt"
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
	mu              sync.RWMutex
	lastUsed        time.Time
	initialized     bool
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

// updateLastUsedInternal 在已经持有写锁的情况下更新实例最后使用时间
// 注意：调用此函数前必须已经持有mu的写锁
func updateLastUsedInternal() {
	lastUsed = time.Now()
}

// IsInitialized 检查Chrome实例是否已初始化
func IsInitialized() bool {
	mu.RLock()
	defer mu.RUnlock()
	return initialized && GlobalAllocCtx != nil && GlobalCancel != nil
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
	initialized = false
}

// InitGlobalChrome 初始化全局 Chrome 实例
func InitGlobalChrome() error {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		return nil // 已经初始化
	}

	err := initializeChrome()
	if err != nil {
		return err
	}

	GlobalPdfParams = page.PrintToPDFParams{
		MarginTop:         0.4,
		MarginBottom:      0.4,
		MarginLeft:        0.4,
		MarginRight:       0.4,
		PrintBackground:   false,
		PreferCSSPageSize: true,
	}

	initialized = true
	return nil
}

// initializeChrome 初始化Chrome实例的实际实现
func initializeChrome() error {
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
		// 移除可能导致问题的 single-process 标志
		chromedp.Flag("disable-background-networking", true), // 禁用后台网络请求
	)

	// 创建全局 ExecAllocator（Chrome 进程）
	GlobalAllocCtx, GlobalCancel = chromedp.NewExecAllocator(context.Background(), opts...)

	// 测试连接确保实例正常工作
	ctx, cancel := chromedp.NewContext(GlobalAllocCtx)
	defer cancel()

	// 增加超时时间以适应较慢的环境
	testCtx, testCancel := context.WithTimeout(ctx, 30*time.Second)
	defer testCancel()

	// 简单测试确保Chrome可以正常工作
	err := chromedp.Run(testCtx, chromedp.Navigate("about:blank"))
	if err != nil {
		// 初始化失败，清理资源
		if GlobalCancel != nil {
			GlobalCancel()
		}
		GlobalAllocCtx = nil
		GlobalCancel = nil
		return fmt.Errorf("failed to test chrome connection: %w", err)
	}

	log.Println("全局 Chrome 实例初始化完成")
	// 使用内部函数更新时间戳，因为我们已经持有锁
	updateLastUsedInternal()
	return nil
}
