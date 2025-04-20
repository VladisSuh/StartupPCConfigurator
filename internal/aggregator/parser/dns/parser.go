package dns

import (
	"StartupPCConfigurator/internal/aggregator/usecase"
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// DNSParser содержит логику парсинга конкретного магазина DNS
type DNSParser struct {
	logger *log.Logger
}

// NewDNSParser конструктор (может принимать иные параметры)
func NewDNSParser(logger *log.Logger) *DNSParser {
	return &DNSParser{logger: logger}
}

func (p *DNSParser) ParseProductPage(ctx context.Context, url string) (*ParsedItem, error) {
	p.logger.Printf("Parsing page: %s", url)

	// 1. Создаём новый контекст chromedp
	// Можно создать глобально 1 браузер и переиспользовать => но для упрощения тут локально
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// 2. Случайная задержка (имитация человека, если нужно)
	pause := time.Duration(rand.Intn(5)+5) * time.Second // от 5 до 10 сек
	time.Sleep(pause)

	// 3. Выполняем сценарий: зайти на url, дождаться загрузки, взять HTML
	var html string
	tasks := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(4 * time.Second), // или ждать селектора
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Прокрутить страницу вниз, иногда нужно для lazy-loading
			return chromedp.Run(ctx,
				chromedp.ScrollIntoView(`body`, chromedp.NodeVisible),
			)
		}),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	}
	if err := chromedp.Run(ctx, tasks); err != nil {
		return nil, fmt.Errorf("chromedp run error: %w", err)
	}

	// 4. Парсим HTML через goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("goquery parse error: %w", err)
	}

	// или короче:
	// doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))

	// 5. Извлекаем нужные данные
	// Примерно по тем же селекторам, что в Python
	item := &ParsedItem{}

	title := doc.Find("div.product-card-description__title").First().Text()
	item.Name = title

	price := doc.Find("div.product-buy__price").First().Text()
	item.Price = price

	desc := doc.Find("div.product-card-description-text").First().Text()
	item.Description = desc

	availability := doc.Find("a.order-avail-wrap__link.ui-link.ui-link_blue").First().Text()
	if availability == "" {
		availability = "Товара нет в наличии"
	}
	item.Availability = availability

	// Пример получения главной картинки:
	mainPic, _ := doc.Find("img.product-images-slider__main-img").Attr("src")
	item.MainImage = mainPic

	// Пример парсинга списка картинок
	var pictures []string
	doc.Find("img.product-images-slider__img.loaded.tns-complete").Each(func(i int, s *goquery.Selection) {
		if src, ok := s.Attr("data-src"); ok {
			pictures = append(pictures, src)
		}
	})
	item.Images = pictures

	// Категорию можно искать, например:
	category := "Категория не найдена"
	doc.Find("span").Each(func(i int, s *goquery.Selection) {
		// Логика поиска
		if goquery.NodeName(s) == "span" && s.AttrOr("data-go-back-catalog", "") != "" {
			category = s.Text()
		}
	})
	item.Category = category

	// Пример характеристики:
	var specs []KV
	doc.Find("div.product-characteristics__spec-title").Each(func(i int, s *goquery.Selection) {
		specTitle := s.Text()
		specValue := doc.Find("div.product-characteristics__spec-value").Eq(i).Text()
		specs = append(specs, KV{Key: specTitle, Value: specValue})
	})
	item.Characteristics = specs

	// вернуть результат
	return item, nil
}

// После уже существующего ParseProductPage(...)
func (p *DNSParser) Parse(ctx context.Context, url string) (*usecase.ParsedItem, error) {
	// просто делегируем, но возвращаем нужный usecase.ParsedItem
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

// ParsedItem структура для хранения результата
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

// Пример для хранения ключ-значение:
type KV struct {
	Key   string
	Value string
}
