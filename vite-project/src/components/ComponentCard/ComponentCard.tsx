import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { ComponentCardProps } from '../../types/index';
import styles from './ComponentCard.module.css';
import { Modal } from "../Modal/Modal";


export const ComponentCard = ({ component, onSelect, selected }: ComponentCardProps) => {
  const [minPrice, setMinPrice] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [offers, setOffers] = useState<any[]>([]);
  const [isVisible, setIsVisible] = useState(false);


  useEffect(() => {
    const fetchMinPrice = async () => {
      try {
        setIsLoading(true);
        const token = localStorage.getItem('authToken');
        const response1 = await fetch(
          `http://localhost:8080/auth/me`,
          {
            headers: {
              'Authorization': `Bearer ${token}`, // Добавляем токен в заголовки
              'Content-Type': 'application/json'
            }
          }
        );

        console.log(response1)
        const data1 = await response1.json();
        console.log('цены1', data1);

        const response = await fetch(
          `http://localhost:8080/offers?componentId=${component.id}&sort=priceAsc`,
          {
            headers: {
              'Authorization': `Bearer ${token}`, // Добавляем токен в заголовки
              'Content-Type': 'application/json'
            }
          }
        );

        console.log(response)
        const data = await response.json();
        console.log('цены', data); // Debugging line to check the response structure    
        if (data.offers && data.offers.length > 0) {
          setMinPrice(data.offers[0].price);
        }
      } catch (err) {
        setError('Не удалось загрузить цены');
      } finally {
        setIsLoading(false);
      }
    };

    fetchMinPrice();
  }, [component.id]);

  const fetchOffers = async () => {
    try {

      setIsLoading(true);
      setError(null);

      // Получаем токен из localStorage
      const token = localStorage.getItem('authToken');

      if (!token) {
        throw new Error('Требуется авторизация');
      }

      const response = await fetch(
        `http://localhost:8080/offers?componentId=${component.id}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`, // Добавляем токен в заголовки
            'Content-Type': 'application/json'
          }
        }
      );

      if (!response.ok) {
        if (response.status === 401) {
          throw new Error('Сессия истекла, войдите снова');
        }
        throw new Error(`Ошибка сервера: ${response.status}`);
      }
      const data = await response.json();
      //let data = m
      setOffers(data.offers || []);
      setIsModalOpen(true);
    } catch (err) {
      setError('Не удалось загрузить предложения');
    } finally {
      setIsLoading(false);
    }
  };



  return (
    <div className={styles.card}>
      <div className={styles.card__info}>
        <div className={styles.card__title}>
          {component.name}
        </div>

        <div className={styles.card__infoText}>
          {Object.values(component.specs)
            .map(value => String(value).trim())
            .join(', ')}
        </div>

        <div className={styles.card__details}>
          Подробнее
        </div>
      </div>

      <div className={styles.card__price}>
        {isLoading ? (
          'Загрузка...'
        ) : error ? (
          'Цена не доступна'
        ) : minPrice ? (
          `Цена от ${minPrice.toLocaleString()} ₽`
        ) : (
          'Нет в наличии'
        )}
      </div>

      <div className={styles.card__actions}>
        <button
          className={styles.addButton + (selected ? ' ' + styles.addButtonSelected : '')}
          onClick={onSelect}
        >
          {selected ? 'Удалить' : 'Добавить'}
        </button>

        <button
          className={styles.offersButton}
          onClick={() => {
            setIsVisible(true)
            fetchOffers()
          }}
          disabled={isLoading}
        >
          Посмотреть предложения
        </button>
      </div>

      <Modal isOpen={isVisible} onClose={() => setIsVisible(false)}>
        <div className={styles.modalContent}>
          <button className={styles.closeButton} onClick={() => setIsModalOpen(false)}>
            &times;
          </button>
          <h2>Предложения для {component.name}</h2>
          {isLoading ? (
            <div>Загрузка...</div>
          ) : error ? (
            <div>{error}</div>
          ) : offers.length === 0 ? (
            <div>Нет доступных предложений</div>
          ) : (
            <div className={styles.offersList}>
              {offers.map((offer, index) => (
                <div key={index} className={styles.offerItem}>
                  <div className={styles.offerShop}>{offer.shopName}</div>
                  <div className={styles.offerPrice}>
                    {offer.price.toLocaleString()} {offer.currency}
                  </div>
                  <div className={styles.offerAvailability}>{offer.availability}</div>
                  <a href={offer.url} target="_blank" rel="noopener noreferrer" className={styles.offerLink}>
                    Перейти в магазин
                  </a>
                </div>
              ))}
            </div>
          )}
        </div>
      </Modal>




    </div>
  );
};

const m = {
  "componentId": "cpu-12345",
  "offers": [
    {
      "shopId": "shop-001",
      "shopName": "TechStore",
      "price": 349.99,
      "currency": "USD",
      "availability": "In stock",
      "url": "https://techstore.com/cpu-12345"
    },
    {
      "shopId": "shop-002",
      "shopName": "PC Parts",
      "price": 329.95,
      "currency": "USD",
      "availability": "Limited stock",
      "url": "https://pcparts.com/intel-cpu-12345"
    },
    {
      "shopId": "shop-003",
      "shopName": "ElectroWorld",
      "price": 359.00,
      "currency": "EUR",
      "availability": "Pre-order",
      "url": "https://electroworld.eu/cpu-offer-123"
    },
    {
      "shopId": "shop-004",
      "shopName": "ByteMarket",
      "price": 315.50,
      "currency": "USD",
      "availability": "In stock",
      "url": "https://bytemarket.com/products/cpu12345"
    },
    {
      "shopId": "shop-005",
      "shopName": "CompTech",
      "price": 375.00,
      "currency": "GBP",
      "availability": "Out of stock",
      "url": "https://comptech.uk/cpu-deal-123"
    }
  ]
}