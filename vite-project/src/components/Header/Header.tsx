import { useEffect, useState } from "react";
import styles from "./Header.module.css";
import { Modal } from "../Modal/Modal";
import Register from "../Register/component";
import Login from "../Login/component";
import { useAuth } from "../../AuthContext";
import { Page, Notification } from "../../types";

const Header = ({ setCurrentPage, currentPage }: { setCurrentPage: (page: Page) => void; currentPage: Page }) => {
    const [isVisible, setIsVisible] = useState(false);
    const [openComponent, setOpenComponent] = useState('login');
    const { logout, isAuthenticated, getToken } = useAuth();
    const [unreadCount, setUnreadCount] = useState(0);
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [showNotifications, setShowNotifications] = useState(false);

    const handleIconClick = () => {
        console.log('Icon clicked');
        if (isAuthenticated) {
            setCurrentPage('account')
            console.log('Личный кабинет');
        } else {
            setIsVisible(true);
        }
    }

    useEffect(() => {

        const fetchUnreadCount = async () => {
            try {
                console.log('Проверка авторизации:', isAuthenticated);
                if (!isAuthenticated) {
                    throw new Error('Требуется авторизация');
                }

                const token = getToken();
                console.log('Получение количества непрочитанных уведомлений с токеном:', token);
                console.log('Запрос',{
                    headers: {
                        'Authorization': `Bearer ${token}`
                    },
                });

                const response = await fetch('http://localhost:8080/notifications/unread-count', {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    },
                });
                if (!response.ok) {
                    console.log('Ошибка запроса:', response);
                    throw new Error('Ошибка запроса');
                }
                const data = await response.json();
                setUnreadCount(data.count || 0);
            } catch (err) {
                console.error("Ошибка загрузки уведомлений:", err);
            }
        };

        fetchUnreadCount();
        const interval = setInterval(fetchUnreadCount, 5000); 

        return () => clearInterval(interval);
    }, []);

    const handleNotificationClick = async () => {
        setShowNotifications(!showNotifications);
        if (!showNotifications) {
            try {
                if (!isAuthenticated) {
                    throw new Error('Требуется авторизация');
                }

                const token = getToken();

                const response = await fetch('http://localhost:8080/notifications?page=1&pageSize=10', {
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token}`
                    },
                });
                const data = await response.json();
                setNotifications(data.items || []);
            } catch (err) {
                console.error('Ошибка загрузки уведомлений:', err);
            }
        }
    };

    return (
        <div className={styles.header}>
            <div className={styles.headerContent}>
                <div className={styles.logo}>UConf</div>

                <div className={styles.navigation}>
                    <button
                        className={`${styles.navButton} ${currentPage === 'configurator' ? styles.active : ''}`}
                        onClick={() => setCurrentPage('configurator')}
                    >
                        Конфигуратор
                    </button>
                    <button
                        className={`${styles.navButton} ${currentPage === 'usecases' ? styles.active : ''}`}
                        onClick={() => setCurrentPage('usecases')}
                    >
                        Готовые сборки
                    </button>
                </div>

                <div className={styles.icons}>
                    {isAuthenticated && (
                        <div className={styles.notificationWrapper} onClick={handleNotificationClick}>
                            <img
                                src="src/assets/icon-ringing.png"
                                className={styles.notificationIcon}
                                alt="Уведомления"
                            />
                            {unreadCount > 0 && <div className={styles.notificationDot}></div>}
                        </div>
                    )}
                    {showNotifications && (
                        <div className={styles.notificationsPopup}>
                            {notifications.length === 0 ? (
                                <div className={styles.empty}>Уведомлений нет</div>
                            ) : (
                                notifications.map(notification => (
                                    <div key={notification.id} className={styles.notificationItem}>
                                        <div className={styles.notificationText}>
                                            <strong>{notification.title}</strong>
                                            <p>{notification.message}</p>
                                            <span className={styles.date}>{new Date(notification.createdAt).toLocaleString()}</span>
                                        </div>
                                        {!notification.isRead && <div className={styles.unreadDot}></div>}
                                    </div>
                                ))
                            )}
                        </div>
                    )}
                    <img
                        src={
                            currentPage === 'account'
                                ? 'src/assets/user-icon-active.svg'
                                : 'src/assets/user-icon-white.svg'
                        }
                        alt="Аккаунт"
                        className={styles.icon}
                        onClick={handleIconClick}
                    />


                    {isAuthenticated && (
                        <img
                            src="src/assets/logout.png"
                            className={styles.logoutIcon}
                            alt="Выход"
                            onClick={logout}
                        />
                    )}
                </div>
            </div>

            {/* <button onClick={() => setIsVisible(true)}>вход</button> */}
            <Modal isOpen={isVisible} onClose={() => {
                setIsVisible(false)
            }}>
                {openComponent === 'register' ? (
                    <Register setOpenComponent={setOpenComponent} onClose={() => {
                        setIsVisible(false)
                    }} />
                ) : (
                    <Login setOpenComponent={setOpenComponent} onClose={() => {
                        setIsVisible(false)
                    }} />
                )}

            </Modal>
            {/* {isVisible && <Modal onClose={() => setIsVisible(false)}></Modal>} */}
        </div>
    );
}

export default Header;