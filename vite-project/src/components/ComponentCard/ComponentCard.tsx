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

  const { isAuthenticated, getToken } = useAuth();
  const [isVisible, setIsVisible] = useState(false);
  const [openComponent, setOpenComponent] = useState('login');


  useEffect(() => {
    const fetchMinPrice = async () => {
      try {
        //setIsLoading(true);
        const response = await fetch(
          `http://localhost:8080/offers/min?componentId=${component.id}`,
          {
            method: 'GET',
            headers: {
              'Content-Type': 'application/json',
              'Accept': 'application/json'
            }
          }
        );

        if (!response.ok) {
          throw new Error('Ошибка при загрузке минимальной цены');
        }

        const data = await response.json();
        console.log('Минимальная цена:', data);

        if (data && typeof data.minPrice === 'number') {
          setMinPrice(data.minPrice);
        }
      } catch (err) {
        setError('Не удалось загрузить минимальную цену');
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

      if (!isAuthenticated) {
        throw new Error('Требуется авторизация');
      }

      const token = getToken();

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
      console.log('оферы:', data);
      setOffers(data || []);
    } catch (err) {
      setError('Не удалось загрузить предложения');
    } finally {
      setIsLoading(false);
    }
  };

  const subscribeToComponent = async () => {
    try {
      if (!isAuthenticated) {
        throw new Error('Требуется авторизация');
      }

      const token = getToken();
      console.log('component.id', component.id);
      console.log('token', token);

      console.log('Отправка запроса на подписку на компонент:', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ componentId: String(component.id) })
      });

      const response = await fetch('http://localhost:8080/subscriptions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ componentId: String(component.id) })
      });

      if (!response.ok) {
        if (response.status === 204) {
          console.log('Подписка успешно создана');
        } else if (response.status === 400) {
          console.log(response)
          console.error('Некорректные данные для подписки');
        } else if (response.status === 401) {
          console.error('Пользователь не авторизован');
        } else {
          console.log(response)
          console.error('Неизвестная ошибка:', response.status);
        }
      }

      console.log(response)
    } catch (error) {
      console.error('Ошибка при подписке:', error);
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
        <div>
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

        <img src='src/assets/bell-icon.png' onClick={subscribeToComponent} className={styles.notificationIcon} alt='подписка на уведомления' title='Подписаться на уведомления' />

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
    </div >
  );
}