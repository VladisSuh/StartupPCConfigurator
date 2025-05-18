import { Configuration, ResponseComponent } from '../../types';
import ComponentDetails from '../ComponentDetails/ComponentDetails';
import component from '../Login/component';
import { Modal } from '../Modal/Modal';
import styles from './SavedComponentCard.module.css';
import PriceOffer from '../PriceOffer/PriceOffer';
import { useEffect, useState } from 'react';
import { useAuth } from '../../AuthContext';

const SavedComponentCard = ({ component }: { component: ResponseComponent }) => {

    const [minPrice, setMinPrice] = useState<number | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [offers, setOffers] = useState<any[]>([]);
    const [isOffersVisible, setIsOffersVisible] = useState(false);
    const [isDetailsVisible, setIsDetailsVisible] = useState(false);
    const [showLoginMessage, setShowLoginMessage] = useState(false);

    console.log('SavedComponentCard:', component);

    const { isAuthenticated, getToken } = useAuth();

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
            setOffers(data || []);
        } catch (err) {
            setError('Не удалось загрузить предложения');
        } finally {
            setIsLoading(false);
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
                }
            } catch (err) {
                setError('Не удалось загрузить минимальную цену');
            } finally {
                setIsLoading(false);
            }
        };

        fetchMinPrice();

    }, [component.id]);

    return (
        <div className={styles.card}>
            <div className={styles.card__info}>
                <div className={styles.card__title}>
                    {component.name}
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

                <div
                    className={styles.buttonWrapper}
                >
                    <button
                        className={styles.offersButton}
                        onClick={() => {
                            setIsOffersVisible(true);
                            fetchOffers();
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

            {/* <Modal isOpen={isDetailsVisible} onClose={() => setIsDetailsVisible(false)}>
                <ComponentDetails component={component} />
            </Modal> */}


        </div>
    );
}

export default SavedComponentCard;