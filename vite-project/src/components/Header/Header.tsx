import { useEffect, useState } from "react";
import styles from "./Header.module.css";
import { Modal } from "../Modal/Modal";
import Register from "../Register/component";
import Login from "../Login/component";
import { useAuth } from "../../AuthContext";

const Header = () => {
    const [isVisible, setIsVisible] = useState(false);
    const [openComponent, setOpenComponent] = useState('login');
    const { logout, isAuthenticated } = useAuth();

    const handleIconClick = () => {
        console.log('Icon clicked');
        if (isAuthenticated) {
            console.log('Личный кабинет');
        } else {
            setIsVisible(true);
        }
    }

    return (
        <div className={styles.header}>
            <div className={styles.headerContent}>
                <div className={styles.logo}>UConf</div>
                <div>
                    <img
                        src="src/assets/user-icon.png"
                        className={styles.icon}
                        alt="Аккаунт"
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