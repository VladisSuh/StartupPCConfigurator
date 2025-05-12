import { useState, useEffect } from 'react';
import { ComponentCardProps } from '../../types/index';
import styles from './ComponentCard.module.css';
import { Modal } from "../Modal/Modal";
import PriceOffer from '../PriceOffer/PriceOffer';
import ComponentDetails from '../ComponentDetails/ComponentDetails';
import { useAuth } from '../../AuthContext';
import Login from '../Login/component';
import Register from '../Register/component';


export const ComponentCard = ({ component, onSelect, selected }: ComponentCardProps) => {
  const [minPrice, setMinPrice] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [offers, setOffers] = useState<any[]>([]);
  const [isOffersVisible, setIsOffersVisible] = useState(false);
  const [isDetailsVisible, setIsDetailsVisible] = useState(false);
  const [showLoginMessage, setShowLoginMessage] = useState(false);

  const { isAuthenticated } = useAuth();
  const [isVisible, setIsVisible] = useState(false);
  const [openComponent, setOpenComponent] = useState('login');


  useEffect(() => {
    const fetchMinPrice = async () => {
      try {
        setIsLoading(true);
        const token = localStorage.getItem('authToken');

        const response = await fetch(
          `http://localhost:8080/offers?componentId=${component.id}&sort=priceAsc`,
          {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json'
            }
          }
        );

        const data = await response.json();
        console.log('цены', data);

        if (data && data.length > 0) {
          setMinPrice(data[0].price);
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

      const token = localStorage.getItem('authToken');

      if (!token) {
        throw new Error('Требуется авторизация');
      }

      const response = await fetch(
        `http://localhost:8080/offers?componentId=${component.id}&sort=priceAsc`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
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
      setOffers(data || []);
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

        <div
          className={styles.card__details}
          onClick={() => {
            setIsDetailsVisible(true)
            fetchOffers()
          }}>
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

        <div
          className={styles.buttonWrapper}
        >
          <button
            className={styles.offersButton}
            onClick={() => {
              if (isAuthenticated) {
                setIsOffersVisible(true);
                fetchOffers();
              } else {
                setIsVisible(true);
                setOpenComponent('login');
                setShowLoginMessage(true);
              }
            }}
            disabled={isLoading}
          >
            Посмотреть предложения
          </button>

        </div>
      </div>

      <Modal isOpen={isOffersVisible} onClose={() => setIsOffersVisible(false)}>
        <div className={styles.modalContent}>

          <h2 className={styles.modalTitle}>{component.name}</h2>
          {isLoading ? (
            <div>Загрузка...</div>
          ) : error ? (
            <div>{error}</div>
          ) : offers.length === 0 ? (
            <div>Нет доступных предложений</div>
          ) : (
            <PriceOffer offers={offers} />
          )}
        </div>
      </Modal>

      <Modal isOpen={isDetailsVisible} onClose={() => setIsDetailsVisible(false)}>
        <ComponentDetails component={component} />
      </Modal>

      <Modal isOpen={isVisible} onClose={() => setIsVisible(false)}>
        {openComponent === 'register' ? (
          <Register
            setOpenComponent={(component) => {
              setOpenComponent(component);
              setShowLoginMessage(false); 
            }}
            onClose={() => setIsVisible(false)}
          />
        ) : (
          <Login
            setOpenComponent={(component) => {
              setOpenComponent(component);
              setShowLoginMessage(false); 
            }}
            onClose={() => setIsVisible(false)}
            message={showLoginMessage ? 'Посмотреть предложения можно после авторизации' : undefined}
          />
        )}
      </Modal>
    </div>
  );
}