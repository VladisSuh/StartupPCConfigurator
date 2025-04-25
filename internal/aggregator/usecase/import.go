package usecase

// ImportRecord — одна строка из Excel-прайс-листа
type ImportRecord struct {
	ComponentID  string  // код или ID компонента
	ShopCode     string  // уникальный код магазина (из shops.code)
	Price        float64 // цена
	Currency     string  // валюта, напр. "USD"
	Availability string  // статус наличия
	URL          string  // ссылка на товар
}
