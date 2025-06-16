import { useEffect, useState } from "react";
import styles from "./Header.module.css";
import { Modal } from "../Modal/Modal";
import Register from "../Register/component";
import Login from "../Login/component";
import { useAuth } from "../../AuthContext";
import { Page, Notification } from "../../types";
import { useConfig } from "../../ConfigContext";

const Header = ({ setCurrentPage, currentPage }: { setCurrentPage: (page: Page) => void; currentPage: Page }) => {
    const [isVisible, setIsVisible] = useState(false);
    const [openComponent, setOpenComponent] = useState('login');
    const { logout, isAuthenticated, getToken } = useAuth();
    const [unreadCount, setUnreadCount] = useState(0);
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [showNotifications, setShowNotifications] = useState(false);
    const { setTheme, theme } = useConfig();



    const handleIconClick = () => {
        console.log('Icon clicked');
        if (isAuthenticated) {
            setCurrentPage('account')
            console.log('Личный кабинет');
        } else {
            setIsVisible(true);
        }
    }

    const fetchUnreadCount = async () => {
        try {
            console.log('Проверка авторизации:', isAuthenticated);
            if (!isAuthenticated) {
                throw new Error('Требуется авторизация');
            }

            const token = getToken();
            console.log('Получение количества непрочитанных уведомлений с токеном:', token);
            console.log('Запрос', {
                headers: {
                    'Authorization': `Bearer ${token}`
                },
            });

            const response = await fetch('http://localhost:8080/notifications/count', {
                headers: {
                    'Authorization': `Bearer ${token}`
                },
            });
            if (!response.ok) {
                console.log('Ошибка запроса:', response);
                throw new Error('Ошибка запроса');
            }
            const data = await response.json();
            console.log('fetchUnreadCount', data)
            setUnreadCount(data.unread || 0);
            console.log('unreadCount', unreadCount)
        } catch (err) {
            console.error("Ошибка загрузки уведомлений:", err);
        }
    };

    useEffect(() => {
        const interval = setInterval(fetchUnreadCount, 10000);
        return () => clearInterval(interval);
    }, []);

    const handleNotificationClick = async () => {
        setShowNotifications(!showNotifications);
        if (!showNotifications) {
            try {
                console.log('Проверка авторизации перед загрузкой уведомлений:', isAuthenticated);

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
                console.log('Загруженные уведомления:', data);
                setNotifications(data.items || []);
                console.log('notifications', notifications)
                console.log('data || []', data || [])
            } catch (err) {
                console.error('Ошибка загрузки уведомлений:', err);
            }
        }
    };

    const handleNotificationsClose = async () => {
        setShowNotifications(false);

        const unreadNotifications = notifications.filter(notification => !notification.isRead);

        if (unreadNotifications.length === 0) {
            return;
        }

        const token = getToken();
        if (!token) {
            console.error('Токен не найден');
            return;
        }

        await Promise.all(
            unreadNotifications.map(async (notification) => {
                try {
                    const response = await fetch(
                        `http://localhost:8080/notifications/${notification.id}/read`,
                        {
                            method: 'POST',
                            headers: {
                                'Authorization': `Bearer ${token}`,
                                'Content-Type': 'application/json'
                            }
                        }
                    );

                    if (!response.ok) {
                        throw new Error(`Ошибка для уведомления ${notification.id}`);
                    }

                    return response;
                } catch (error) {
                    console.error(`Ошибка при отметке уведомления ${notification.id} как прочитанного:`, error);
                    return null;
                }
            })
        );
        fetchUnreadCount();
    }

    return (
        <div className={styles.header}>
            <div className={styles.headerContent}>
                <div className={styles.logo}>UConf</div>

                <div className={styles.navigation}>
                    <button
                        className={`${styles.navButton} ${currentPage === 'configurator' ? styles.active : ''} ${styles[theme]}`}
                        onClick={() => setCurrentPage('configurator')}
                    >
                        Конфигуратор
                    </button>
                    <button
                        className={`${styles.navButton} ${currentPage === 'usecases' ? styles.active : ''} ${styles[theme]}`}
                        onClick={() => setCurrentPage('usecases')}
                    >
                        Готовые сборки
                    </button>
                </div>

                <div className={styles.iconContainer}>
                    {isAuthenticated && (
                        <div className={styles.notificationWrapper} onClick={handleNotificationClick}>
                            {theme === 'dark' ? (
                                <img
                                    src="src/assets/notifications-light.svg"
                                    alt="Уведомления"
                                />
                            ) : (
                                <img
                                    src="src/assets/notifications-dark.svg"
                                    alt="Уведомления"
                                />
                            )}
                            {unreadCount > 0 && <div className={styles.notificationDot}></div>}
                        </div>
                    )}

                    {theme === 'dark' ? (
                        <img
                            src="src/assets/account-light.svg"
                            alt="Аккаунт"
                            onClick={handleIconClick}
                        />
                    ) : (
                        <img
                            src="src/assets/account-dark.svg"
                            alt="Аккаунт"
                            onClick={handleIconClick}
                        />
                    )}


                    {isAuthenticated && (
                        theme === 'dark' ? (
                            <img
                                src="src/assets/logout-light.svg"
                                alt="Выход"
                                onClick={logout}
                            />
                        ) : (
                            <img
                                src="src/assets/logout-dark.svg"
                                alt="Выход"
                                onClick={logout}
                            />
                        )
                    )}


                    {theme === 'dark' ? (
                        <img
                            src="src/assets/light_mode.svg"
                            alt="Светлая тема"
                            onClick={() => setTheme('light')}
                        />
                    ) : (
                        <img
                            src="src/assets/dark_mode.svg"
                            alt="Темная тема"
                            onClick={() => setTheme('dark')}
                        />
                    )}
                </div>

                {showNotifications && (
                    <div className={styles.overlay} onClick={handleNotificationsClose}>
                        <div className={styles.notificationsPopup} onClick={(e) => e.stopPropagation()}>
                            {notifications.length === 0 ? (
                                <div className={styles.empty}>Уведомлений нет</div>
                            ) : (
                                notifications.map(notification => (
                                    <div key={notification.id} className={styles.notificationItem}>
                                        <div className={styles.notificationText}>
                                            <strong className={styles.notificationTitle}>{notification.componentName}</strong>
                                            <p className={styles.notificationContent}>Цена изменилась:</p>
                                            <p className={styles.notificationContent}>{notification.oldPrice} {' -> '} {notification.newPrice}</p>
                                            <span className={styles.date}>{new Date(notification.createdAt).toLocaleString()}</span>
                                        </div>
                                        {!notification.isRead && <div className={styles.unreadDot}></div>}
                                    </div>
                                ))
                            )}
                        </div>
                    </div>
                )}
            </div>

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
        </div>
    );
}

export default Header;