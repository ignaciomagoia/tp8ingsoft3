import React, { useEffect } from "react";

const TOAST_TIMEOUT = 3500;

export default function Toast({ message, type = "info", onClose }) {
  useEffect(() => {
    if (!message) return undefined;

    const timer = setTimeout(() => {
      if (onClose) {
        onClose();
      }
    }, TOAST_TIMEOUT);

    return () => clearTimeout(timer);
  }, [message, onClose]);

  if (!message) {
    return null;
  }

  return (
    <div className={`toast toast--${type}`} role="alert" aria-live="assertive">
      <span>{message}</span>
      <button type="button" onClick={onClose} aria-label="Cerrar notificación">
        ×
      </button>
    </div>
  );
}
