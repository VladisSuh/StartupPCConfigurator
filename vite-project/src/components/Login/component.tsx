import { useForm, SubmitHandler } from "react-hook-form"
import styles from "./styles.module.css";
import classNames from "classnames";
import { useState } from "react";

import eyeOpenDark from '../../assets/eye-open-dark.svg'
import eyeOpenLight from '../../assets/eye-open-light.svg'
import eyeClosedDark from '../../assets/eye-closed-dark.svg'
import eyeClosedLight from '../../assets/eye-closed-light.svg'

import { useAuth } from "../../AuthContext";
import { useConfig } from "../../ConfigContext";

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
    const { theme } = useConfig();

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
            {message && <div className={`${styles.message} ${styles[theme]}`}>{message}</div>}

            <h2 className={styles.formLabel}>Авторизация</h2>

            <label>Email</label>
            <input
                className={`${styles.input} ${styles[theme]} ${errors.email ? styles.errorInput : ''}`}
                {...register("email", {
                    required: "Email обязателен",
                    pattern: {
                        value: emailPattern,
                        message: "Неверный формат email"
                    }
                })}
                onBlur={() => handleBlur("email")}
            />
            {errors.email && (
                <p role="alert" className={`${styles.errorMessage} ${styles[theme]}`}>
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
                    className={`${styles.input} ${styles.passwordInput} ${styles[theme]} ${errors.email ? styles.errorInput : ''}`}
                />
                <button
                    type="button"
                    className={styles.showPasswordButton}
                    onClick={() => setShowPassword(prev => !prev)}
                >
                    {theme === 'dark' ? (
                        <img
                            className={`${styles.eyeIcon}`}
                            src={showPassword ? eyeClosedLight : eyeOpenLight}
                        />
                    ) : (
                        <img
                            className={`${styles.eyeIcon}`}
                            src={showPassword ? eyeClosedDark : eyeOpenDark}
                        />
                    )}
                </button>
            </div>

            {errors.password && (
                <p role="alert" className={`${styles.errorMessage} ${styles[theme]}`}>
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