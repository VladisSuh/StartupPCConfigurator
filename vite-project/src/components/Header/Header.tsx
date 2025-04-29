import { useState } from "react";
import styles from "./Header.module.css";
import { Modal } from "../Modal/Modal";
import Register from "../Register/component";
import Login from "../Login/component";

const Header = () => {
    const [isVisible, setIsVisible] = useState(false);
    const [openComponent, setOpenComponent] = useState('register');

    return (
        <div className={styles.header}>
            <div className={styles.headerContent}>
                <div className={styles.logo}>UConf</div>
                <img
                    src="src/components/Header/r.png"
                    className={styles.icon}
                    alt="Иконка"
                />
            </div>
            <button onClick={() => setIsVisible(true)}>вход</button>
            <Modal isOpen={isVisible} onClose={() => {
                setIsVisible(false)
                setOpenComponent('register')
            }}>
                {openComponent === 'register' ? (
                    <Register setOpenComponent={setOpenComponent} />
                ) : (
                    <Login setOpenComponent={setOpenComponent} />
                )}

            </Modal>
            {/* {isVisible && <Modal onClose={() => setIsVisible(false)}></Modal>} */}
        </div>
    );
}

export default Header;