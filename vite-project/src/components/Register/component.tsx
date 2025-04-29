import { useForm, SubmitHandler } from "react-hook-form"
import styles from "./styles.module.css";
import classNames from "classnames";
import { useState } from "react";
import iconOpenEye from '../../assets/icon-open-eye.png'
import iconClosedEye from '../../assets/icon-closed-eye.png'


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

export default function Register({ setOpenComponent }: { setOpenComponent: (component: string) => void }) {

    const { register, handleSubmit, formState: { errors, isValid, isDirty }, reset } = useForm<RegisterData>()
    const [showPassword, setShowPassword] = useState(false);
    const [loading, setLoading] = useState(false);
    const [errorMessage, setErrorMessage] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);

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
                setSuccessMessage(`Регистрация успешна! Добро пожаловать, ${result.user.name}`);
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


    return (
        <form onSubmit={handleSubmit(onSubmit)} className={classNames(styles.form)}>
            <h2 className={styles.formLabel}>Регистрация</h2>

            <label>Имя</label>
            <input {...register("name", { required: true, maxLength: 20 })} />
            {errors.name?.type === "required" && (
                <p role="alert">Обязательное поле</p>
            )}

            <label>Email</label>
            <input {...register("email", { required: true, pattern: emailPattern })} />
            {errors.email?.type === "required" && (<p role="alert">Обязательное поле</p>)}
            {errors.email?.type === "pattern" && (<p role="alert">Неверный формат email</p>)}

            <label>Пароль</label>
            <div className={styles.passwordInputContainer}>
                <input
                    type={showPassword ? "text" : "password"}
                    {...register("password", {
                        required: true,
                        minLength: 8,
                        maxLength: 20,
                        pattern: passwordPattern
                    })}
                    className={styles.passwordInput}
                />
                {errors.password?.type === "required" && (
                    <p role="alert" className={styles.errorMessage}>Обязательное поле</p>
                )}
                {errors.password?.type === "minLength" && (
                    <p role="alert" className={styles.errorMessage}>Минимум 8 символов</p>
                )}
                {errors.password?.type === "maxLength" && (
                    <p role="alert" className={styles.errorMessage}>Максимум 20 символов</p>
                )}
                {errors.password?.type === "pattern" && (
                    <p role="alert" className={styles.errorMessage}>
                        Пароль должен содержать только латинские буквы, цифры и спецсимволы
                    </p>
                )}

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


            <button
                type="submit"
                disabled={loading || !isDirty}
            >
                {loading ? "Загрузка..." : "Зарегистрироваться"}
            </button>
            {errorMessage && <p className={styles.error}>{errorMessage}</p>}
            {successMessage && (
                <div className={styles.success}>
                    <p>
                        {successMessage}
                        <span
                            onClick={() => setOpenComponent('login')}
                            className={styles.loginLink}
                        >
                            {' '} Войти
                        </span>
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