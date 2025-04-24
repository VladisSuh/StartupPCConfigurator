package citilink

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"StartupPCConfigurator/internal/aggregator/usecase"
)

type CitilinkParser struct {
	logger *log.Logger
}

func NewStealthContext(parent context.Context, logger *log.Logger) (context.Context, context.CancelFunc) {
	// 1. Создаём allocator (браузер + флаги)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(
		parent,
		chromedp.ExecPath(`C:\Program Files\Google\Chrome\Application\chrome.exe`),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent(
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
				"AppleWebKit/537.36 (KHTML, like Gecko) "+
				"Chrome/135.0.0.0 Safari/537.36",
		),
	)

	// 2. Новый контекст с CDP‑слушателями
	ctx, cancelCtx := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(logger.Printf),
	)

	// 3. Устанавливаем таймаут на всё
	ctx, cancelTimeout := context.WithTimeout(ctx, 90*time.Second)

	// Вот здесь — ActionFunc, который внутри вызывает Do и игнорирует ScriptIdentifier:
	stealth := chromedp.ActionFunc(func(ctx context.Context) error {
		// включаем сеть, ставим доп. заголовки
		if err := network.Enable().Do(ctx); err != nil {
			return err
		}
		if err := network.SetExtraHTTPHeaders(network.Headers{
			"Accept-Language": "ru-RU,ru;q=0.9",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9",
		}).Do(ctx); err != nil {
			return err
		}
		// инжектим скрипт перед любой загрузкой страницы
		js := `(() => {
            Object.defineProperty(navigator, 'webdriver', { get: () => false });
            window.chrome = { runtime: {} };
            Object.defineProperty(navigator, 'languages', {
              get: () => ['ru-RU', 'ru', 'en-US', 'en']
            });
            Object.defineProperty(navigator, 'plugins', {
              get: () => [1, 2, 3, 4, 5]
            });
        })();`
		// AddScriptToEvaluateOnNewDocumentParams.Do возвращает (ScriptIdentifier, error)
		// мы используем ActionFunc, чтобы вернуть только ошибку
		if _, err := page.AddScriptToEvaluateOnNewDocument(js).Do(ctx); err != nil {
			return err
		}
		return nil
	})

	// Прогоним готовый stealth‑action один раз, чтобы он установился
	if err := chromedp.Run(ctx, stealth); err != nil {
		logger.Fatalf("failed to install stealth scripts: %v", err)
	}

	// Объединяем все cancel’ы в один
	cancel := func() {
		cancelTimeout()
		cancelCtx()
		cancelAlloc()
	}
	return ctx, cancel
}

func NewCitilinkParser(logger *log.Logger) *CitilinkParser {
	return &CitilinkParser{logger: logger}
}

func (p *CitilinkParser) ParseProductPage(ctx context.Context, url string) (*ParsedItem, error) {
	p.logger.Printf("Parsing Citilink page: %s", url)

	// (1) Создаём Stealth-контекст, как для DNS
	ctx, cancel := NewStealthContext(ctx, p.logger)
	defer cancel()

	// (2) Навигация + ожидание
	var html string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`.ProductCardLayout__price`, chromedp.ByQuery),
		chromedp.OuterHTML("html", &html),
	); err != nil {
		return nil, fmt.Errorf("chromedp run error: %w", err)
	}

	// (3) Goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("goquery parse error: %w", err)
	}

	// (4) Извлечение данных под новые селекторы Citilink
	item := &ParsedItem{}
	item.Name = doc.Find("h1.ProductHeading__title").Text()
	item.Price = doc.Find(".ProductCardLayout__price-current").Text()
	item.Description = doc.Find(".ProductCardDescription__text").Text()
	item.Availability = doc.Find(".ProductAvailability__status").Text()
	item.Category = doc.Find(".Breadcrumbs__item:last-child a").Text()

	// картинка и дополнительные изображения
	if src, ok := doc.Find(".ProductGallery__main img").Attr("src"); ok {
		item.MainImage = src
	}
	doc.Find(".ProductGallery__thumbs img").Each(func(i int, s *goquery.Selection) {
		if src, ok := s.Attr("src"); ok {
			item.Images = append(item.Images, src)
		}
	})

	// характеристики — прим. отличаются по классу
	doc.Find(".ProductSpecs__row").Each(func(i int, s *goquery.Selection) {
		k := s.Find(".ProductSpecs__name").Text()
		v := s.Find(".ProductSpecs__value").Text()
		if k != "" {
			item.Characteristics = append(item.Characteristics, KV{Key: k, Value: v})
		}
	})

	return item, nil
}

// Parse — обвязка под usecase.Parser
func (p *CitilinkParser) Parse(ctx context.Context, url string) (*usecase.ParsedItem, error) {
	prod, err := p.ParseProductPage(ctx, url)
	if err != nil {
		return nil, err
	}
	return &usecase.ParsedItem{
		Price:        prod.Price,
		Availability: prod.Availability,
		URL:          url,
	}, nil
}

// структура результата внутри пакета parser/citilink
type ParsedItem struct {
	Name            string
	Price           string
	Description     string
	Availability    string
	Category        string
	MainImage       string
	Images          []string
	Characteristics []KV
}

type KV struct {
	Key   string
	Value string
}
