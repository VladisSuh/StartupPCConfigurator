import { useForm, SubmitHandler } from "react-hook-form"
import styles from "./styles.module.css";
import classNames from "classnames";
import { useState } from "react";

import iconOpenEye from '../../assets/icon-open-eye.png'
import iconClosedEye from '../../assets/icon-closed-eye.png'
import { useAuth } from "../../AuthContext";

type LoginData = {
    email: string
    password: string
}

const emailPattern = /^[\w.-]+@[a-zA-Z\d.-]+\.[a-zA-Z]{2,}$/;


export default function Login({ setOpenComponent, onClose, message }:
    {
        setOpenComponent: (component: string) => void,
        onClose: () => void,
        message?: string
    }) {

    const {
        register,
        handleSubmit,
        formState: { errors, isValid, isDirty },
        reset,
        trigger
    } = useForm<LoginData>({
        mode: "onChange"

    })
    const [showPassword, setShowPassword] = useState(false);

    const [loading, setLoading] = useState(false);
    const [errorMessage, setErrorMessage] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);

    const { login } = useAuth();

    const onSubmit: SubmitHandler<LoginData> = async (data) => {
        setLoading(true);
        setErrorMessage(null);
        setSuccessMessage(null);

        try {
            const response = await fetch('http://localhost:8080/auth/login', {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Accept": "application/json"
                },
                body: JSON.stringify({
                    email: data.email,
                    password: data.password
                }),
            });

            console.log(response);

            const responseData = await response.json();

            console.log('responseData', responseData);

            if (response.ok) {

                console.log('Login successful:', responseData);
                login(responseData.token.access_token);
                onClose()
                //setSuccessMessage(`Добро пожаловать!`);
                reset();
            } else if (response.status === 401) {
                setErrorMessage("Неверный email или пароль");
            } else if (response.status === 400) {
                setErrorMessage(responseData.message || "Некорректный запрос");
            } else {
                setErrorMessage(responseData.message || "Произошла ошибка при входе");
            }
        } catch (error) {
            console.error('Login error:', error);
            setErrorMessage("Ошибка соединения с сервером");
        } finally {
            setLoading(false);
        }
    };

    const handleBlur = (fieldName: keyof LoginData) => {
        trigger(fieldName);
    };

    return (
        <form onSubmit={handleSubmit(onSubmit)} className={classNames(styles.form)}>
            {message && <div className={styles.message}>{message}</div>}

            <h2 className={styles.formLabel}>Авторизация</h2>


            <label>Email</label>
            <input
                {...register("email", {
                    required: "Email обязателен",
                    pattern: {
                        value: emailPattern,
                        message: "Неверный формат email"
                    }
                })}
                onBlur={() => handleBlur("email")}
                className={errors.email ? styles.errorInput : ''}
            />
            {errors.email && (
                <p role="alert" className={styles.errorMessage}>
                    {errors.email.message}
                </p>
            )}


            <label>Пароль</label>
            <div className={styles.passwordInputContainer}>
                <input
                    type={showPassword ? "text" : "password"}
                    {...register("password", {
                        required: "Пароль обязателен",
                    })}
                    onBlur={() => handleBlur("password")}
                    className={`${styles.passwordInput} ${errors.password ? styles.errorInput : ''}`}
                />
                <button
                    type="button"
                    className={styles.showPasswordButton}
                    onClick={() => setShowPassword(prev => !prev)}
                >
                    <img
                        src={showPassword ? iconClosedEye : iconOpenEye}
                        alt={showPassword ? "Скрыть пароль" : "Показать пароль"}
                        className={styles.eyeIcon}
                    />
                </button>
            </div>
            {errors.password && (
                <p role="alert" className={styles.errorMessage}>
                    {errors.password.message}
                </p>
            )}


            <button
                type="submit"
                disabled={loading || !isValid || !isDirty}
                className={styles.submitButton}
            >
                {loading ? "Загрузка..." : "Войти"}
            </button>
            {errorMessage && <p className={styles.error}>{errorMessage}</p>}
            {successMessage && <p className={styles.success}>{successMessage}</p>}

            <p className={styles.registerLink}>
                Еще нет аккаунта?{" "}

                <span
                    onClick={() => setOpenComponent('register')}
                    className={styles.loginLink}
                >
                    {' '} Зарегистрироваться
                </span>

            </p>
        </form>
    )
}