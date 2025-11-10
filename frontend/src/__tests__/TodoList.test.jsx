import React from "react";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
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

const loginAsDemo = async () => {
  await act(async () => {
    await userEvent.type(screen.getByLabelText(/email/i), "demo@example.com");
    await userEvent.type(screen.getByLabelText(/contraseña/i), "pass123");
    await userEvent.click(screen.getByRole("button", { name: /iniciar sesión/i }));
  });

  await waitFor(() => expect(loginUser).toHaveBeenCalled());
  await waitFor(() => expect(getTodos).toHaveBeenCalled());
};

describe("App integration", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    updateTodo.mockResolvedValue({ todo: { id: "1", title: "Mock", completed: true } });
    deleteTodo.mockResolvedValue({ message: "deleted" });
  });

  it("renderiza el título To-Do List", () => {
    render(<App />);
    expect(screen.getByRole("heading", { name: /to-do list/i })).toBeInTheDocument();
  });

  it("actualiza la lista al crear una tarea", async () => {
    registerUser.mockResolvedValue({ message: "Registrado" });
    loginUser.mockResolvedValue({ message: "Login" });
    getTodos.mockResolvedValueOnce({ todos: [] });
    createTodo.mockResolvedValue({
      todo: { id: "todo-1", title: "Preparar informe", completed: false },
    });

    render(<App />);

    await act(async () => {
      await userEvent.type(screen.getByLabelText(/email/i), "demo@example.com");
      await userEvent.type(screen.getByLabelText(/contraseña/i), "pass123");
      await userEvent.click(screen.getByRole("button", { name: /registrar/i }));
    });

    await waitFor(() => expect(registerUser).toHaveBeenCalled());
    await waitFor(() => expect(getTodos).toHaveBeenCalledWith("demo@example.com"));

    await act(async () => {
      await userEvent.type(
        screen.getByLabelText(/título de la tarea/i),
        "Preparar informe"
      );
      await userEvent.click(screen.getByRole("button", { name: /agregar tarea/i }));
    });

    await waitFor(() => expect(createTodo).toHaveBeenCalledWith({
      email: "demo@example.com",
      title: "Preparar informe",
    }));

    expect(await screen.findByText(/preparar informe/i)).toBeInTheDocument();
  });

  it("permite editar una tarea existente y guarda los cambios", async () => {
    registerUser.mockResolvedValue({ message: "Registrado" });
    loginUser.mockResolvedValue({ message: "Login" });
    getTodos.mockResolvedValueOnce({
      todos: [{ id: "todo-1", title: "Original", completed: false }],
    });
    updateTodo.mockResolvedValue({
      todo: { id: "todo-1", title: "Editada", completed: false },
    });

    render(<App />);

    await loginAsDemo();

    await act(async () => {
      await userEvent.click(screen.getByRole("button", { name: /editar/i }));
    });

    const editInput = await screen.findByLabelText(/editar tarea original/i);
    await act(async () => {
      fireEvent.change(editInput, { target: { value: "Editada" } });
    });
    expect(editInput).toHaveValue("Editada");
    await act(async () => {
      await userEvent.click(screen.getByRole("button", { name: /guardar/i }));
    });

    await waitFor(() =>
      expect(updateTodo).toHaveBeenCalledWith("todo-1", { title: "Editada" })
    );
    expect(await screen.findByText(/editada/i)).toBeInTheDocument();
  });

  it("muestra un mensaje si se intenta guardar una edición vacía", async () => {
    registerUser.mockResolvedValue({ message: "Registrado" });
    loginUser.mockResolvedValue({ message: "Login" });
    getTodos.mockResolvedValueOnce({
      todos: [{ id: "todo-2", title: "Persistente", completed: false }],
    });

    render(<App />);

    await loginAsDemo();

    await act(async () => {
      await userEvent.click(screen.getByRole("button", { name: /editar/i }));
    });

    const editInput = await screen.findByLabelText(/editar tarea persistente/i);
    await act(async () => {
      fireEvent.change(editInput, { target: { value: "" } });
    });
    expect(editInput).toHaveValue("");
    await act(async () => {
      await userEvent.click(screen.getByRole("button", { name: /guardar/i }));
    });

    await waitFor(() => expect(updateTodo).not.toHaveBeenCalled());
    expect(await screen.findByText(/el título es requerido/i)).toBeInTheDocument();
  });
});
