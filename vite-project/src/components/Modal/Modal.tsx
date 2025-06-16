import React from 'react';
import styles from './Modal.module.css';
import { useConfig } from '../../ConfigContext';

interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    children: React.ReactNode;
}

export const Modal = ({ isOpen, onClose, children }: ModalProps) => {
    if (!isOpen) return null;
    const { theme } = useConfig();

    return (
        <div className={styles.overlay} onClick={onClose}>
            <div className={`${styles.modal} ${styles[theme]}`} onClick={e => e.stopPropagation()}>
                <button className={`${styles.closeButton} ${styles[theme]}`} onClick={onClose}>
                    ×
                </button>
                {children}
            </div>
        </div>
    );
};
