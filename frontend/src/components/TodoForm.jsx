import React, { useState } from "react";

export default function TodoForm({ onAdd, disabled = false }) {
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async (event) => {
    event.preventDefault();
    const trimmed = title.trim();

    if (!trimmed) {
      setError("El título es requerido");
      return;
    }

    setError("");
    if (onAdd) {
      await onAdd(trimmed);
    }
    setTitle("");
  };

  return (
    <form onSubmit={handleSubmit} className="todo-form">
      <label className="form-field todo-field">
        <span>Nueva tarea</span>
        <input
          type="text"
          placeholder="Ej: Comprar víveres"
          value={title}
          onChange={(event) => setTitle(event.target.value)}
          aria-label="Título de la tarea"
          disabled={disabled}
          data-cy="todo-input"
        />
      </label>
      <button type="submit" className="btn btn--primary" disabled={disabled} data-cy="add-btn">
        Agregar tarea
      </button>
      {error && (
        <span role="alert" className="form-error">
          {error}
        </span>
      )}
    </form>
  );
}
