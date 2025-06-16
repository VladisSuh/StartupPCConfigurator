import { useForm, SubmitHandler } from "react-hook-form"
import styles from "./styles.module.css";
import classNames from "classnames";
import { useState } from "react";
import iconOpenEye from '../../assets/icon-open-eye.png'
import iconClosedEye from '../../assets/icon-closed-eye.png'
import { useConfig } from "../../ConfigContext";


type RegisterData = {
    email: string
    password: string
    name: string
}

type RegisterResponse = {
    user: {
        id: string;
        email: string;
        name: string;
        roles: string[];
    };
    accessToken: string;
};

const emailPattern = /^[\w.-]+@[a-zA-Z\d.-]+\.[a-zA-Z]{2,}$/;
const passwordPattern = /^[A-Za-z\d!@#$%^&*()_+{}\[\]:;'"\\|,.<>\/?~`-]{8,}$/;

const API_URL = "http://localhost:8080/auth/register";

export default function Register({ setOpenComponent, onClose }: { setOpenComponent: (component: string) => void, onClose: () => void }) {

    const {
        register,
        handleSubmit,
        formState: { errors, isValid, isDirty },
        reset,
        trigger
    } = useForm<RegisterData>({
        mode: "onChange"
    })

    const [showPassword, setShowPassword] = useState(false);
    const [loading, setLoading] = useState(false);
    const [errorMessage, setErrorMessage] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);
    const { theme } = useConfig();

    const onSubmit: SubmitHandler<RegisterData> = async (data) => {
        setLoading(true);
        setErrorMessage(null);
        setSuccessMessage(null);

        console.log(data);

        try {
            const response = await fetch(API_URL, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(data),
            });

            console.log(response);

            if (response.ok) {
                const result: RegisterResponse = await response.json();
                setSuccessMessage(`Регистрация успешна! Теперь вы можете войти в аккаунт.`);
                reset();
            } else if (response.status === 400) {
                setErrorMessage("Ошибка валидации. Проверьте введенные данные.");
            } else if (response.status === 409) {
                setErrorMessage("Пользователь с таким email уже существует.");
            } else {
                setErrorMessage("Произошла неизвестная ошибка.");
            }
        } catch (error) {
            setErrorMessage("Ошибка сети. Попробуйте позже.");
        } finally {
            setLoading(false);
        }
    };

    const handleBlur = (fieldName: keyof RegisterData) => {
        trigger(fieldName);
    };


    return (
        <form onSubmit={handleSubmit(onSubmit)} className={classNames(styles.form)}>
            <h2 className={styles.formLabel}>Регистрация</h2>

            <label className={styles.inputLabel} >Имя</label>
            <input {...register("name",
                {
                    required: "Обязательное поле",
                    maxLength: {
                        value: 20,
                        message: "Максимум 20 символов" 
                    },
                })}
                onBlur={() => handleBlur("name")}
                className={`${styles.input} ${styles[theme]} ${errors.email ? styles.errorInput : ''}`}
            />
            {errors.name && (
                <p role="alert" className={`${styles.errorMessage} ${styles[theme]}`}>
                    {errors.name.message}
                </p>
            )}

            <label className={styles.inputLabel}>Email</label>
            <input
                {...register("email", {
                    required: "Обязательное поле",
                    pattern: {
                        value: emailPattern,
                        message: "Неверный формат email"
                    }
                })}
                onBlur={() => handleBlur("email")}
                className={`${styles.input} ${styles[theme]} ${errors.email ? styles.errorInput : ''}`}
            />
            {errors.email && (
                <p role="alert" className={`${styles.errorMessage} ${styles[theme]}`}>
                    {errors.email.message}
                </p>
            )}

            <label className={styles.inputLabel}>Пароль</label>
            <div className={styles.passwordInputContainer}>
                <input
                    type={showPassword ? "text" : "password"}
                    {...register("password", {
                        required: "Обязательное поле",
                        minLength: {
                            value: 8,
                            message: "Минимум 8 символов"
                        },
                        maxLength: {
                            value: 20,
                            message: "Максимум 20 символов"
                        },
                        pattern: {
                            value: passwordPattern,
                            message: "Пароль должен содержать только латинские буквы, цифры и спецсимволы"
                        }
                    })}
                    onBlur={() => handleBlur("password")}
                    className={`${styles.input} ${styles[theme]} ${styles.passwordInput} ${errors.password ? styles.errorInput : ''}`}
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
                <div className={`${styles.errorMessage} ${styles[theme]}`}>
                    {errors.password.message}
                </div>
            )}



            <button
                type="submit"
                disabled={loading || !isDirty || !isValid}
                className={styles.submitButton}
            >
                {loading ? "Загрузка..." : "Зарегистрироваться"}
            </button>
            {errorMessage && <p className={styles.error}>{errorMessage}</p>}
            {successMessage && (
                <div className={styles.success}>
                    <p className={styles.successMessage}>
                        Регистрация успешна! Теперь вы можете
                        <span
                            onClick={() => setOpenComponent('login')}
                            className={styles.loginLink}
                        >
                            войти в аккаунт
                        </span>.
                    </p>
                </div>
            )}
            {!successMessage && (
                <p className={styles.registerLink}>
                    Уже есть аккаунт?{" "}

                    <span
                        onClick={() => setOpenComponent('login')}
                        className={styles.loginLink}
                    >
                        {' '} Войти
                    </span>

                </p>
            )}
        </form>
    )
}