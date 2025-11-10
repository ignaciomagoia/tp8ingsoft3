import React, { useEffect, useState } from "react";

export default function TodoItem({ todo, onToggle, onDelete, onUpdate }) {
  const [isEditing, setIsEditing] = useState(false);
  const [draft, setDraft] = useState(todo.title);

  useEffect(() => {
    setDraft(todo.title);
  }, [todo.title]);

  const startEditing = () => {
    setIsEditing(true);
  };

  const cancelEditing = () => {
    setDraft(todo.title);
    setIsEditing(false);
  };

  const saveChanges = () => {
    const trimmed = draft.trim();
    if (onUpdate) {
      onUpdate(todo.id, trimmed);
    }
    if (!trimmed) {
      setDraft(todo.title);
      setIsEditing(false);
      return;
    }
    setIsEditing(false);
  };

  return (
    <li className="todo-item" data-cy="todo-item">
      <label className="todo-item__content">
        <input
          type="checkbox"
          checked={todo.completed}
          onChange={() => onToggle(todo.id, todo.completed)}
          aria-label={`Cambiar estado de ${todo.title}`}
        />
        {isEditing ? (
          <input
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            aria-label={`Editar tarea ${todo.title}`}
            data-cy={`edit-input-${todo.id}`}
          />
        ) : (
          <span className={todo.completed ? "todo-title todo-title--done" : "todo-title"}>
            {todo.title}
          </span>
        )}
      </label>
      <div className="todo-item__actions">
        {isEditing ? (
          <>
            <button
              type="button"
              className="btn btn--primary"
              onClick={saveChanges}
              data-cy={`save-${todo.id}`}
            >
              Guardar
            </button>
            <button type="button" className="btn btn--ghost" onClick={cancelEditing}>
              Cancelar
            </button>
          </>
        ) : (
          <>
            <button
              type="button"
              className="btn btn--ghost"
              onClick={startEditing}
              data-cy={`edit-${todo.id}`}
            >
              Editar
            </button>
            <button type="button" className="btn btn--ghost" onClick={() => onDelete(todo.id)}>
              Eliminar
            </button>
          </>
        )}
      </div>
    </li>
  );
}
