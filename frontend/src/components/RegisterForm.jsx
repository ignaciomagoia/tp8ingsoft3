import React, { useState } from "react";

const initialState = { email: "", password: "" };

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function RegisterForm({
  onRegister,
  onLogin,
  disabled = false,
  defaultEmail = "",
}) {
  const [values, setValues] = useState({
    ...initialState,
    email: defaultEmail,
  });
  const [errors, setErrors] = useState({});

  const updateField = (field) => (event) => {
    setValues((prev) => ({ ...prev, [field]: event.target.value }));
  };

  const validate = () => {
    const newErrors = {};
    const trimmedEmail = values.email.trim().toLowerCase();

    if (!trimmedEmail) {
      newErrors.email = "El email es requerido";
    } else if (!emailRegex.test(trimmedEmail)) {
      newErrors.email = "El email no tiene un formato válido";
    }

    if (!values.password.trim()) {
      newErrors.password = "La contraseña es requerida";
    }

    setErrors(newErrors);
    return { isValid: Object.keys(newErrors).length === 0, trimmedEmail };
  };

  const handleAction = async (action) => {
    if (disabled) {
      return;
    }

    const { isValid, trimmedEmail } = validate();
    if (!isValid) {
      return;
    }

    const payload = { email: trimmedEmail, password: values.password };
    if (action === "register" && onRegister) {
      await onRegister(payload);
    } else if (action === "login" && onLogin) {
      await onLogin(payload);
    }
    setValues((prev) => ({ ...prev, password: "" }));
  };

  return (
    <section aria-labelledby="auth-section-title" className="panel panel--auth">
      <header className="panel__header">
        <h2 id="auth-section-title" className="panel__title">
          Acceder
        </h2>
        <p className="panel__subtitle">
          Registrate o iniciá sesión para administrar tus tareas.
        </p>
      </header>

      <div className="form-grid">
        <label className="form-field">
          <span>Email</span>
          <input
            type="email"
            value={values.email}
            onChange={updateField("email")}
            aria-label="Email"
            aria-invalid={Boolean(errors.email)}
            disabled={disabled}
          />
          {errors.email && (
            <span role="alert" className="form-error">
              {errors.email}
            </span>
          )}
        </label>

        <label className="form-field">
          <span>Contraseña</span>
          <input
            type="password"
            value={values.password}
            onChange={updateField("password")}
            aria-label="Contraseña"
            aria-invalid={Boolean(errors.password)}
            disabled={disabled}
          />
          {errors.password && (
            <span role="alert" className="form-error">
              {errors.password}
            </span>
          )}
        </label>
      </div>

      <div className="form-actions">
        <button
          type="button"
          className="btn btn--primary"
          onClick={() => handleAction("register")}
          disabled={disabled}
        >
          Registrar
        </button>
        <button
          type="button"
          className="btn btn--outline"
          onClick={() => handleAction("login")}
          disabled={disabled}
        >
          Iniciar sesión
        </button>
      </div>
    </section>
  );
}
