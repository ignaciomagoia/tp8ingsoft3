import React from "react";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import TodoForm from "../components/TodoForm";
import { createTodo } from "../services/api";

describe("TodoForm", () => {
  it("muestra error cuando el título está vacío", async () => {
    render(<TodoForm onAdd={jest.fn()} />);

    await act(async () => {
      await userEvent.click(screen.getByRole("button", { name: /agregar tarea/i }));
    });

    expect(screen.getByText(/el título es requerido/i)).toBeInTheDocument();
  });

  it("ejecuta onAdd con el título sin espacios", async () => {
    const handleAdd = jest.fn();
    render(<TodoForm onAdd={handleAdd} />);

    await act(async () => {
      await userEvent.type(
        screen.getByLabelText(/título de la tarea/i),
        "  Comprar pan  "
      );
      await userEvent.click(screen.getByRole("button", { name: /agregar tarea/i }));
    });

    expect(handleAdd).toHaveBeenCalledWith("Comprar pan");
  });
});

describe("api service", () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      headers: { get: () => "application/json" },
      json: () => Promise.resolve({ todo: { id: "1", title: "Test", completed: false } }),
    });
  });

  afterEach(() => {
    jest.resetAllMocks();
    global.fetch = originalFetch;
  });

  it("envía la petición correcta al crear una tarea", async () => {
    await createTodo({ email: "demo@example.com", title: "Test" });

    expect(global.fetch).toHaveBeenCalledWith("http://localhost:8080/todos", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: "demo@example.com", title: "Test" }),
    });
  });
});
