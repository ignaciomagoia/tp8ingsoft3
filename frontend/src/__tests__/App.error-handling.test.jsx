import React from "react";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "../App";
import {
  registerUser,
  loginUser,
  getTodos,
  createTodo,
  updateTodo,
  deleteTodo,
} from "../services/api";

jest.mock("../services/api");

describe("App error handling", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    registerUser.mockResolvedValue({ message: "Registrado" });
    loginUser.mockResolvedValue({ message: "Inicio de sesi\u00f3n exitoso" });
    updateTodo.mockResolvedValue({ todo: { id: "1", title: "Mock", completed: true } });
    deleteTodo.mockResolvedValue({ message: "ok" });
  });

  const loginAs = async (email, password) => {
    await act(async () => {
      await userEvent.type(screen.getByLabelText(/email/i), email);
      await userEvent.type(screen.getByLabelText(/contrase\u00f1a/i), password);
      await userEvent.click(screen.getByRole("button", { name: /iniciar sesi\u00f3n/i }));
    });
  };

  it("muestra un toast de error cuando falla la carga de tareas", async () => {
    getTodos.mockRejectedValueOnce(new Error("Fallo carga"));
    render(<App />);

    await loginAs("demo@example.com", "123456");

    await waitFor(() => {
      expect(getTodos).toHaveBeenCalledWith("demo@example.com");
    });

    expect(await screen.findByText(/fallo carga/i)).toBeInTheDocument();
  });

  it("muestra un toast cuando no se puede crear una tarea", async () => {
    getTodos.mockResolvedValueOnce({ todos: [] });
    createTodo.mockRejectedValueOnce(new Error("No se pudo crear"));
    render(<App />);

    await loginAs("tester@example.com", "123456");

    await act(async () => {
      await userEvent.type(
        screen.getByLabelText(/t\u00edtulo de la tarea/i),
        "Preparar demo"
      );
      await userEvent.click(screen.getByRole("button", { name: /agregar tarea/i }));
    });

    expect(await screen.findByText(/no se pudo crear/i)).toBeInTheDocument();
  });
});
