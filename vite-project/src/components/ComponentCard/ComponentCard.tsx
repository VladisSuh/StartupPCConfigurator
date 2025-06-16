import { useState, useEffect } from 'react';
import { ComponentCardProps } from '../../types/index';
import styles from './ComponentCard.module.css';
import { Modal } from "../Modal/Modal";
import PriceOffer from '../PriceOffer/PriceOffer';
import ComponentDetails from '../ComponentDetails/ComponentDetails';
import { useAuth } from '../../AuthContext';
import Login from '../Login/component';
import Register from '../Register/component';
import { useConfig } from '../../ConfigContext';
import toast, { Toaster } from 'react-hot-toast';

export const ComponentCard = ({ component, onSelect, selected, onPriceLoaded }: ComponentCardProps) => {
  const [minPrice, setMinPrice] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [offers, setOffers] = useState<any[]>([]);
  const [isOffersVisible, setIsOffersVisible] = useState(false);
  const [isDetailsVisible, setIsDetailsVisible] = useState(false);
  const [showLoginMessage, setShowLoginMessage] = useState(false);
  const [isSubscribed, setIsSubscribed] = useState(false);

  const { isAuthenticated, getToken } = useAuth();
  const [isVisible, setIsVisible] = useState(false);
  const [openComponent, setOpenComponent] = useState('login');
  const { theme } = useConfig();

  const checkSubscriptionStatus = async () => {
    if (!isAuthenticated) return;

    try {
      const token = getToken();
      const response = await fetch(
        `http://localhost:8080/subscriptions/status?ids=${component.id}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        }
      );

      if (!response.ok) {
        throw new Error('Ошибка при проверке подписки');
      }

      const data = await response.json();
      setIsSubscribed(data[component.id] === true);
    } catch (err) {
      console.error('Ошибка при проверке подписки:', err);
    }
  };
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
          onPriceLoaded(data.minPrice);
        }
      } catch (err) {
        setError('Не удалось загрузить минимальную цену');
      } finally {
        setIsLoading(false);
      }
    };

    fetchMinPrice();
    checkSubscriptionStatus();

  }, [component.id]);

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
          onPriceLoaded(data.minPrice);
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

  const handleSubscription = async () => {
    try {
      if (!isAuthenticated) {
        setIsVisible(true);
        setOpenComponent('login');
        return;
      }

      const token = getToken();

      if (isSubscribed) {
        const response = await fetch(
          `http://localhost:8080/subscriptions/${component.id}`,
          {
            method: 'DELETE',
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json'
            }
          }
        );

        if (response.ok) {
          toast.success(`Вы отписались от обновлений компонента ${component.name}`);
          setIsSubscribed(false);
        } else if (response.status === 404) {
          toast.error('Подписка не найдена');
        } else {
          const errorMessage = await response.text();
          toast.error(errorMessage || 'Ошибка при отписке');
        }
      } else {
        const response = await fetch('http://localhost:8080/subscriptions', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
          },
          body: JSON.stringify({ componentId: String(component.id) })
        });

        if (response.ok) {
          toast.success(`Вы подписались на обновления компонента ${component.name}`);
          setIsSubscribed(true);
        } else if (response.status === 400) {
          toast.error('Некорректные данные для подписки');
        } else {
          const errorMessage = await response.text();
          toast.error(errorMessage || 'Ошибка при подписке');
        }
      }
    } catch (error) {
      console.error('Ошибка при подписке:', error);
      toast.error('Произошла ошибка при выполнении запроса');
    }
  };



  return (
    <div className={styles.card}>
      <div className={styles.info}>
        <div className={styles.title}>
          {component.name}
        </div>

        <div className={styles.infoText}>
          {Object.values(component.specs)
            .map(value => String(value).trim())
            .join(', ')}
        </div>

        <div
          className={`${styles.details} ${styles[theme]}`}
          onClick={() => {
            setIsDetailsVisible(true)
            fetchOffers()
          }}>
          Подробнее
        </div>
      </div>

      <div className={styles.price}>
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

        <div onClick={handleSubscription}>
          {isAuthenticated && isSubscribed ? (
            <img
              className={styles.notification}
              src={theme === 'dark'
                ? "src/assets/notifications-off-light.svg"
                : "src/assets/notifications-off-dark.svg"}

              alt="Отписаться от уведомлений"
            />
          ) : (
            <img
              className={styles.notification}
              src={theme === 'dark'
                ? "src/assets/notifications-active-light.svg"
                : "src/assets/notifications-active-dark.svg"}
              alt="Подписаться на уведомления"
            />
          )}
          <Toaster position="top-right" />
        </div>
      </div>

      <div className={styles.actions}>
        <button
          className={`${styles.addButton} ${selected ? styles.addButtonSelected : ''} ${styles[theme]}`}
          onClick={onSelect}
        >
          {selected ? 'Удалить' : 'Добавить'}
        </button>


        <button
          className={`${styles.offersButton} ${styles[theme]}`}
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