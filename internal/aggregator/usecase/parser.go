package usecase

import "context"

// Parser умеет парсить один URL и возвращать структурированный результат
type Parser interface {
	Parse(ctx context.Context, url string) (*ParsedItem, error)
}

// Скопируйте сюда только нужные поля из dns.ParsedItem:
type ParsedItem struct {
	Price        string
	Availability string
	URL          string
}
