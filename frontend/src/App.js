import React, { useCallback, useEffect, useState } from "react";
import "./App.css";
import RegisterForm from "./components/RegisterForm";
import TodoForm from "./components/TodoForm";
import TodoList from "./components/TodoList";
import Toast from "./components/Toast";
import {
  registerUser,
  loginUser,
  getTodos,
  createTodo,
  updateTodo,
  deleteTodo,
} from "./services/api";

const emptyToast = { message: "", type: "info" };

function App() {
  const [currentUser, setCurrentUser] = useState("");
  const [todos, setTodos] = useState([]);
  const [isLoadingTodos, setIsLoadingTodos] = useState(false);
  const [toast, setToast] = useState(emptyToast);

  const showToast = useCallback((message, type = "info") => {
    setToast({ message, type });
  }, []);

  const clearToast = useCallback(() => {
    setToast(emptyToast);
  }, []);

  const loadTodos = useCallback(
    async (email) => {
      if (!email) {
        setTodos([]);
        return;
      }
      setIsLoadingTodos(true);
      try {
        const response = await getTodos(email);
        setTodos(response.todos ?? []);
      } catch (error) {
        showToast(error.message, "error");
      } finally {
        setIsLoadingTodos(false);
      }
    },
    [showToast]
  );

  useEffect(() => {
    loadTodos(currentUser);
  }, [currentUser, loadTodos]);

  const handleAuthSuccess = useCallback(
    (email, message) => {
      setCurrentUser(email);
      showToast(message ?? "Sesión iniciada", "success");
    },
    [showToast]
  );

  const handleRegister = async ({ email, password }) => {
    try {
      const response = await registerUser({ email, password });
      handleAuthSuccess(email, response.message ?? "Usuario registrado correctamente");
    } catch (error) {
      showToast(error.message, "error");
      throw error;
    }
  };

  const handleLogin = async ({ email, password }) => {
    try {
      const response = await loginUser({ email, password });
      handleAuthSuccess(email, response.message ?? "Inicio de sesión exitoso");
    } catch (error) {
      showToast(error.message, "error");
      throw error;
    }
  };

  const handleLogout = () => {
    setCurrentUser("");
    setTodos([]);
    showToast("Sesión cerrada", "info");
  };

  const handleCreateTodo = async (title) => {
    if (!currentUser) {
      showToast("Debes iniciar sesión para crear tareas", "warning");
      return;
    }

    try {
      const response = await createTodo({ email: currentUser, title });
      const created = response.todo ?? response;
      setTodos((prev) => [...prev, created]);
      showToast("Tarea creada", "success");
    } catch (error) {
      showToast(error.message, "error");
    }
  };

  const handleToggleTodo = async (id, completed) => {
    try {
      const response = await updateTodo(id, { completed: !completed });
      const updated = response.todo ?? response;
      setTodos((prev) => prev.map((todo) => (todo.id === id ? updated : todo)));
    } catch (error) {
      showToast(error.message, "error");
    }
  };

  const handleRenameTodo = async (id, title) => {
    const trimmed = title.trim();
    if (!trimmed) {
      showToast("El título es requerido", "warning");
      return;
    }

    try {
      const response = await updateTodo(id, { title: trimmed });
      const updated = response.todo ?? response;
      setTodos((prev) => prev.map((todo) => (todo.id === id ? updated : todo)));
      showToast("Tarea actualizada", "success");
    } catch (error) {
      showToast(error.message, "error");
    }
  };

  const handleDeleteTodo = async (id) => {
    try {
      await deleteTodo(id);
      setTodos((prev) => prev.filter((todo) => todo.id !== id));
      showToast("Tarea eliminada", "info");
    } catch (error) {
      showToast(error.message, "error");
    }
  };

  return (
    <div className="app">
      <header className="app__header">
        <h1>To-Do List</h1>
        <p>Gestioná tus tareas pendientes y mantené todo bajo control.</p>
      </header>

      <main className="app__content">
        <RegisterForm
          onRegister={handleRegister}
          onLogin={handleLogin}
          disabled={Boolean(currentUser)}
          defaultEmail={currentUser}
        />

        {currentUser ? (
          <section aria-labelledby="todos-section-title" className="panel todo-panel">
            <div className="todo-panel__header">
              <div>
                <h2 id="todos-section-title" className="section__title">
                  Mis tareas
                </h2>
                <p className="section__subtitle">
                  Sesión activa como <strong>{currentUser}</strong>
                </p>
              </div>

              <button type="button" className="btn btn--outline" onClick={handleLogout}>
                Cerrar sesión
              </button>
            </div>

            <TodoForm onAdd={handleCreateTodo} disabled={isLoadingTodos} />

            {isLoadingTodos ? (
              <p role="status">Cargando tareas...</p>
            ) : (
              <TodoList
                todos={todos}
                onToggle={handleToggleTodo}
                onDelete={handleDeleteTodo}
                onUpdate={handleRenameTodo}
              />
            )}
          </section>
        ) : (
          <section className="panel panel--placeholder">
            <p className="panel__subtitle">
              Registrate o iniciá sesión para comenzar a agregar tareas y hacer seguimiento de tus pendientes.
            </p>
          </section>
        )}
      </main>

      <Toast message={toast.message} type={toast.type} onClose={clearToast} />
    </div>
  );
}

export default App;
