import React from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import TodoList from "../components/TodoList";

describe("TodoList component", () => {
  it("muestra un mensaje cuando no hay tareas", () => {
    render(<TodoList todos={[]} onToggle={jest.fn()} onDelete={jest.fn()} />);
    expect(screen.getByRole("status")).toHaveTextContent(/no hay tareas cargadas/i);
  });

  it("permite alternar y eliminar una tarea", async () => {
    const handleToggle = jest.fn();
    const handleDelete = jest.fn();
    const todos = [{ id: "abc", title: "Estudiar", completed: false }];

    render(<TodoList todos={todos} onToggle={handleToggle} onDelete={handleDelete} />);

    await userEvent.click(
      screen.getByRole("checkbox", { name: /cambiar estado de estudiar/i })
    );
    expect(handleToggle).toHaveBeenCalledWith("abc", false);

    await userEvent.click(screen.getByRole("button", { name: /eliminar/i }));
    expect(handleDelete).toHaveBeenCalledWith("abc");
  });
});
