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


export default function Login({ setOpenComponent }: { setOpenComponent: (component: string) => void }) {

    const { register, handleSubmit, formState: { errors, isValid, isDirty }, reset } = useForm<LoginData>()
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
                // Успешная авторизация
                /* const result = responseData as {
                    user: {
                        id: string;
                        email: string;
                        name: string;
                        roles: string[];
                    };
                    accessToken: string;
                }; */
                console.log('Login successful:', responseData);
                login(responseData.token.access_token); // Сохраняем токен в контексте

                /* setSuccessMessage(`Добро пожаловать, ${result.user.name}!`); */

                setSuccessMessage(`Добро пожаловать!`);

                // Сохраняем токен (пример)
                //localStorage.setItem('accessToken', result.accessToken);

                // Перенаправляем пользователя (пример)
                // navigate('/dashboard');

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

    return (
        <form onSubmit={handleSubmit(onSubmit)} className={classNames(styles.form)}>
            <h2 className={styles.formLabel}>Авторизация</h2>

            <label>Email</label>
            <input {...register("email", { required: true, pattern: emailPattern })} />
            {errors.email?.type === "required" && (<p role="alert">Email is required</p>)}
            {errors.email?.type === "pattern" && (<p role="alert">Invalid email address</p>)}

            <label>Пароль</label>
            <div className={styles.passwordInputContainer}>
                <input
                    type={showPassword ? "text" : "password"}
                    {...register("password", { required: true, maxLength: 20 })}
                    className={styles.passwordInput}
                />
                {errors.password?.type === "required" && (
                    <p role="alert" className={styles.errorMessage}>Обязательное поле</p>
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

            {errors.password?.type === "required" && (<p role="alert">Password is required</p>)}
            {/* {errors.password?.type === "required" && (<p role="alert">Password is required</p>)} */}


            <button
                type="submit"
                disabled={loading || !isValid || !isDirty}
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